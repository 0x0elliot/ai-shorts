package util

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	models "go-authentication-boilerplate/models"

	storage "cloud.google.com/go/storage"
	genai "github.com/google/generative-ai-go/genai"
	openai "github.com/sashabaranov/go-openai"
	"google.golang.org/api/option"
)

var OPENAI_API_KEY = os.Getenv("ACIDRAIN_OPENAI_KEY")

type ImageStyle string

const (
	DefaultStyle    ImageStyle = "default"
	AnimeStyle      ImageStyle = "anime"
	CartoonStyle    ImageStyle = "cartoon"
	WatercolorStyle ImageStyle = "watercolor"
)

type GeneratedImage struct {
	Sentence string
	ImageURL string
}

type SentencePrompt struct {
	Sentence string
	Prompt   string
}

type SDXLRequest struct {
	ModelName         string  `json:"modelName"`
	Prompt            string  `json:"prompt"`
	Prompt2           string  `json:"prompt2,omitempty"`
	ImageHeight       int     `json:"imageHeight"`
	ImageWidth        int     `json:"imageWidth"`
	NegativePrompt    string  `json:"negativePrompt,omitempty"`
	NegativePrompt2   string  `json:"negativePrompt2,omitempty"`
	NumOutputImages   int     `json:"numOutputImages"`
	GuidanceScale     float64 `json:"guidanceScale"`
	NumInferenceSteps int     `json:"numInferenceSteps"`
	Seed              *int    `json:"seed,omitempty"`
	OutputImgType     string  `json:"outputImgType"`
}
type SDXLResponse struct {
	Data []struct {
		B64JSON string `json:"b64_json"`
	} `json:"data"`
}

func SaveVideoError(video *models.Video, err error) error {
	video.Error = err.Error()
	_, err = SetVideo(video)
	if err != nil {
		return err
	}

	return nil
}

func CreateVideo(video *models.Video, recreate bool) (*models.Video, error) {
	client := openai.NewClient(OPENAI_API_KEY)
	storageClient, err := GetGCPClient()
	if err != nil {
		log.Printf("[ERROR] Error creating storage client: %v", err)
		return nil, err
	}

	if recreate {
		// clean up everything from the bucket
		bucketName := os.Getenv("ACIDRAIN_GCP_BUCKET_NAME")
		videoID := video.ID
		folderPath := path.Join("videos", videoID)

		if err := DeleteFolderFromBucket(context.Background(), storageClient, bucketName, folderPath); err != nil {
			log.Printf("[ERROR] Error deleting folder from bucket: %v", err)
		}

		video.Progress = 0
		video.ScriptGenerated = false
		video.DALLEPromptGenerated = false
		video.DALLEGenerated = false
		video.TTSGenerated = false
		video.VideoStitched = false
		video.VideoGenerated = false
		video.VideoUploaded = false
		video.Error = ""
		video.TTSURL = ""
		video.StitchedVideoURL = ""

		video, err = SetVideo(video)
		if err != nil {
			log.Printf("[ERROR] Error saving video: %v", err)
			return nil, err
		}
	}

	log.Printf("[INFO] Processing content for video: %s", video.ID)

	cleanedtopic, script, err := processContent(client, video.Topic, video.Description)
	if err != nil {
		log.Printf("[ERROR] Error processing content: %v", err)
		err = SaveVideoError(video, err)
		return nil, err
	}

	log.Printf("[INFO] Processed content for video: %s", video.ID)

	// Update the video struct with the processed content
	video.Topic = cleanedtopic
	video.Script = script
	video.ScriptGenerated = true

	video.Progress = 10

	// Save the video to the database
	video, err = SetVideo(video)
	if err != nil {
		log.Printf("[ERROR] Error saving video: %v", err)
		_ = SaveVideoError(video, err)
		return nil, err
	}

	// now, we can create the image
	// Create a new context and storage client
	ctx := context.Background()
	if err != nil {
		log.Printf("[ERROR] Error creating storage client: %v", err)
		_ = SaveVideoError(video, err)
		return nil, err
	}

	bucketName := os.Getenv("ACIDRAIN_GCP_BUCKET_NAME")

	log.Printf("[INFO] Generating content for video: %s", video.ID)

	// first generate the TTS for the script
	// then get whispers to generate the SRT file
	_, err = generateTTSForScript(ctx, client, storageClient, bucketName, video.ID, video.Script, video.Narrator)
	if err != nil {
		log.Printf("[ERROR] Error generating TTS: %v", err)
		_ = SaveVideoError(video, err)
		return nil, err
	}

	log.Printf("[INFO] Generated TTS for video: %s", video.ID)

	video.Progress = 30
	video.TTSGenerated = true

	video, err = SetVideo(video)
	if err != nil {
		log.Printf("[ERROR] Error saving video: %v", err)
		_ = SaveVideoError(video, err)
		return nil, err
	}

	log.Printf("[INFO] Generating SRT for video: %s", video.ID)

	srtContent, err := generateSRTForTTSTranscript(ctx, storageClient, bucketName, video.ID)
	if err != nil {
		log.Printf("[ERROR] Error generating SRT: %v", err)
		_ = SaveVideoError(video, err)
		return nil, err
	}

	log.Printf("[INFO] Generated SRT for video: %s", video.ID)

	video.Progress = 40
	video.SRTGenerated = true
	video.SRTURL = fmt.Sprintf("https://storage.googleapis.com/%s/videos/%s/subtitles/subtitles.srt", bucketName, video.ID)

	video, err = SetVideo(video)
	if err != nil {
		log.Printf("[ERROR] Error saving video: %v", err)
		_ = SaveVideoError(video, err)
		return nil, err
	}

	log.Printf("[INFO] Generating DALL-E prompts for video: %s", video.ID)

	prompts, err := generateDallEPromptsForScript(client, srtContent, video)
	if err != nil {
		log.Printf("[ERROR] Error generating DALL-E prompts: %v", err)
		_ = SaveVideoError(video, err)
		return nil, err
	}

	log.Printf("[INFO] Generated DALL-E prompts for video: %s", video.ID)

	video.Progress = 45
	video.DALLEPromptGenerated = true

	video, err = SetVideo(video)
	if err != nil {
		log.Printf("[ERROR] Error saving video: %v", err)
		_ = SaveVideoError(video, err)
		return nil, err
	}

	log.Printf("Generated %d prompts for the video script", len(prompts))
	log.Printf("prompts: %v", prompts)

	log.Printf("[INFO] Generating images for video: %s", video.ID)

	_, err = generateImagesForScript(ctx, client, storageClient, bucketName, video.ID, prompts, DefaultStyle)
	if err != nil {
		log.Printf("[ERROR] Error generating images: %v", err)
		_ = SaveVideoError(video, err)
		return nil, err
	}

	log.Printf("[INFO] Generated images for video: %s", video.ID)

	video.Progress = 60
	video.DALLEGenerated = true

	video, err = SetVideo(video)
	if err != nil {
		log.Printf("[ERROR] Error saving video: %v", err)
		_ = SaveVideoError(video, err)
		return nil, err
	}

	// call StitchVideo function
	log.Printf("[INFO] Stitching video for video: %s", video.ID)


	return video, nil
}

func generateSRTForTTSTranscript(ctx context.Context, client *storage.Client, bucketName, videoID string) (string, error) {
	// Construct the path to the existing audio file
	audioObjectPath := fmt.Sprintf("videos/%s/audio/full_audio.mp3", videoID)

	// Create a signed URL for the existing audio file
	bucket := client.Bucket(bucketName)
	// audioObj := bucket.Object(audioObjectPath)

	// generate the audio URL from bucket
	// example:
	// https://storage.googleapis.com/zappush_public/videos/3291a801-3b3b-4018-b706-670fcce05563/image_1.png
	audioURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, audioObjectPath)

	// Use Whisper to generate SRT content
	srtContent, err := generateSRTWithWhisper(audioURL)
	if err != nil {
		return "", fmt.Errorf("error generating SRT with Whisper: %v", err)
	}

	// Save the SRT file to the bucket
	srtObjectPath := fmt.Sprintf("videos/%s/subtitles/subtitles.srt", videoID)
	_, err = saveSRTToBucket(ctx, bucket, srtObjectPath, srtContent)
	if err != nil {
		return "", fmt.Errorf("error saving SRT to bucket: %v", err)
	}

	return srtContent, nil
}

func generateSRTWithWhisper(audioURL string) (string, error) {
	openaiClient := openai.NewClient(OPENAI_API_KEY)

	// Download the audio file from the bucket
	resp, err := http.Get(audioURL)
	if err != nil {
		return "", fmt.Errorf("error downloading audio file: %v", err)
	}
	defer resp.Body.Close()

	// Create a temporary file to store the audio
	tempFile, err := os.CreateTemp("", "audio*.mp3")
	if err != nil {
		return "", fmt.Errorf("error creating temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up the temp file when we're done

	// Copy the audio data to the temporary file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("error writing to temporary file: %v", err)
	}
	tempFile.Close()

	// Now use the temporary file for the Whisper API request
	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: tempFile.Name(),
		Format:   openai.AudioResponseFormatSRT,
	}

	respOA, err := openaiClient.CreateTranscription(context.Background(), req)
	if err != nil {
		log.Printf("[ERROR] Error creating transcription: %v", err)
		return "", fmt.Errorf("error creating transcription: %v", err)
	}

	return respOA.Text, nil
}

func saveSRTToBucket(ctx context.Context, bucket *storage.BucketHandle, objectPath string, srtContent string) (string, error) {
	obj := bucket.Object(objectPath)
	writer := obj.NewWriter(ctx)
	_, err := writer.Write([]byte(srtContent))
	if err != nil {
		return "", fmt.Errorf("error writing SRT to bucket: %v", err)
	}
	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("error closing writer: %v", err)
	}

	// Make the object publicly accessible
	if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return "", fmt.Errorf("error setting ACL: %v", err)
	}

	// Get the public URL
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return "", fmt.Errorf("error getting object attributes: %v", err)
	}

	return attrs.MediaLink, nil
}

func generateTTSForScript(ctx context.Context, client *openai.Client, storageClient *storage.Client, bucketName, videoID string, script string, narrator string) (string, error) {
	audioData, err := generateTTSForFullScript(client, script, narrator)
	if err != nil {
		return "", fmt.Errorf("error generating TTS for script: %v", err)
	}

	bucket := storageClient.Bucket(bucketName)
	folderPath := path.Join("videos", videoID, "audio")

	// Save the audio to the bucket
	filename := "full_audio.mp3"
	objectPath := path.Join(folderPath, filename)
	audioURL, err := saveAudioToBucket(ctx, bucket, objectPath, audioData)
	if err != nil {
		return "", fmt.Errorf("error saving audio to bucket: %v", err)
	}

	return audioURL, nil
}

func generateTTSForFullScript(client *openai.Client, script, narrator string) ([]byte, error) {
	req := openai.CreateSpeechRequest{
		Model: openai.TTSModel1,
		Input: script,
		Voice: openai.VoiceAlloy, // Default voice, adjust based on narrator preference
	}

	// Map narrator to OpenAI voice options
	switch narrator {
	case "alloy":
		req.Voice = openai.VoiceAlloy
	case "echo":
		req.Voice = openai.VoiceEcho
	case "fable":
		req.Voice = openai.VoiceFable
	case "onyx":
		req.Voice = openai.VoiceOnyx
	case "nova":
		req.Voice = openai.VoiceNova
	case "shimmer":
		req.Voice = openai.VoiceShimmer
	}

	resp, err := client.CreateSpeech(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("speech creation failed: %v", err)
	}

	audioData, err := ioutil.ReadAll(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio data: %v", err)
	}

	return audioData, nil
}

func saveAudioToBucket(ctx context.Context, bucket *storage.BucketHandle, objectPath string, audioData []byte) (string, error) {
	obj := bucket.Object(objectPath)
	writer := obj.NewWriter(ctx)
	writer.ContentType = "audio/mpeg"

	if _, err := writer.Write(audioData); err != nil {
		return "", fmt.Errorf("failed to write audio data to bucket: %v", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %v", err)
	}

	// Make the object publicly accessible
	if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return "", fmt.Errorf("failed to set ACL: %v", err)
	}

	// Get the public URL
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get object attributes: %v", err)
	}

	return attrs.MediaLink, nil
}

func generateImagesForScript(ctx context.Context, client *openai.Client, storageClient *storage.Client, bucketName, videoID string, sentencePrompts []SentencePrompt, style ImageStyle) ([]GeneratedImage, error) {
	var generatedImages []GeneratedImage

	bucket := storageClient.Bucket(bucketName)
	folderPath := path.Join("videos", videoID)

	for i, sp := range sentencePrompts {
		imageData, err := generateImageForPrompt(sp.Prompt, style, 1)
		if err != nil {
			return nil, fmt.Errorf("error generating image for prompt %d: %v", i+1, err)
		}

		// Save the image to the bucket
		filename := fmt.Sprintf("image_%d.png", i+1)
		objectPath := path.Join(folderPath, filename)
		imageURL, err := saveImageToBucket(ctx, bucket, objectPath, imageData)
		if err != nil {
			return nil, fmt.Errorf("error saving image %d to bucket: %v", i+1, err)
		}

		generatedImages = append(generatedImages, GeneratedImage{
			Sentence: sp.Sentence,
			ImageURL: imageURL,
		})
	}

	return generatedImages, nil
}

func generateImageForPrompt(prompt string, style ImageStyle, numImages int) ([]byte, error) {
	// stylePrompt := getStylePrompt(style)
	fullPrompt := prompt

	apiKey := os.Getenv("ACIDRAIN_OLA_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ACIDRAIN_OLA_KEY environment variable not set")
	}

	log.Printf("Generating image for prompt: %s", fullPrompt)

	reqBody := SDXLRequest{
		ModelName:         "diffusion1XL",
		Prompt:            fullPrompt,
		ImageHeight:       1024,
		ImageWidth:        1024,
		NumOutputImages:   numImages,
		GuidanceScale:     7.5,
		NumInferenceSteps: 10,
		OutputImgType:     "pil",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", "https://cloud.olakrutrim.com/v1/images/generations/diffusion", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(body))
	}

	var sdxlResp SDXLResponse
	err = json.Unmarshal(body, &sdxlResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if len(sdxlResp.Data) == 0 {
		return nil, fmt.Errorf("no image data received")
	}

	// remember, output image type is "pil", so we need to decode the base64 string
	b64JSON := sdxlResp.Data[0].B64JSON
	imageData, err := base64.StdEncoding.DecodeString(b64JSON)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 image data: %v", err)
	}

	return imageData, nil
}

func saveImageToBucket(ctx context.Context, bucket *storage.BucketHandle, objectPath string, imageData []byte) (string, error) {
	obj := bucket.Object(objectPath)
	writer := obj.NewWriter(ctx)
	writer.ContentType = "image/png"

	if _, err := io.Copy(writer, bytes.NewReader(imageData)); err != nil {
		return "", fmt.Errorf("failed to copy image data to bucket: %v", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %v", err)
	}

	// Make the object publicly accessible
	if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return "", fmt.Errorf("failed to set ACL: %v", err)
	}

	// Get the public URL
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get object attributes: %v", err)
	}

	return attrs.MediaLink, nil
}

func generateDallEPromptsForScript(client *openai.Client, srtContent string, video *models.Video) ([]SentencePrompt, error) {
	var script string

	if len(srtContent) == 0 {
		script = video.Script
	} else {
		// get the script content
		resp, err := http.Get(video.SRTURL)
		if err != nil {
			return nil, fmt.Errorf("error downloading script content: %v", err)
		}
		defer resp.Body.Close()

		scriptBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading script content: %v", err)
		}

		script = string(scriptBytes)
	}

	// strip all emojis away from the script
	// script = StripEmoji(script)
	sentences := SplitSRTIntoSentences(script)

	var results []SentencePrompt
	for _, sentence := range sentences {
		prompt, err := generateDallEPromptForSentence(client, sentence, video)
		if err != nil {
			return nil, fmt.Errorf("error generating prompt for sentence '%s': %v", sentence, err)
		}
		results = append(results, SentencePrompt{Sentence: sentence, Prompt: prompt})
	}

	return results, nil
}

func generateDallEPromptForSentenceGemini(formattedSentence string, video *models.Video) (string, error) {
	topic := video.Topic
	description := video.Description
	style := video.VideoStyle

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("ACIDRAIN_GEMNI_KEY")))
	if err != nil {
		return "", fmt.Errorf("error creating Gemini client: %v", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-pro")

	styleInstruction := getStyleInstruction(style)

	prompt := fmt.Sprintf(`Generate a detailed SDXL prompt based on the following information:
Sentence: %s
Topic: %s
Description: %s
Style Instruction: %s
Guidelines for crafting the prompt:

Emphasize visual elements and atmosphere rather than literal interpretation of the sentence.
Use rich, descriptive language to convey mood, lighting, and textures.
Incorporate specific artistic styles, techniques, or historical art movements mentioned in the style instruction.
Avoid requesting text or specific logos. Focus on creating a vivid scene or concept.
Use a format like "[Subject], [Setting], [Mood/Atmosphere], [Style], [Additional details]" to structure the prompt.
Include relevant details from the topic and description to enhance context, but prioritize visual appeal.
Specify camera angles, perspectives, or composition when appropriate (e.g. "close-up view", "wide-angle shot", "birds-eye perspective").
Mention color palettes or lighting conditions that fit the overall theme and style.
Include details about materials, textures, or surface qualities to enhance realism or artistic effect.
Use evocative adjectives and sensory language to make the prompt more vivid.
Specify the desired level of detail or realism (e.g. "photorealistic", "impressionistic", "highly detailed").
Avoid unnecessary formatting or markdown. Present the prompt as plain text.
Keep the prompt concise but descriptive, aiming for 2-3 sentences maximum.
If applicable, mention specific artists or art styles that align with the desired outcome.
Focus on creating a cohesive, visually striking image that captures the essence of the sentence and context.
`, formattedSentence, topic, description, styleInstruction)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("error generating content: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	jsonResponse := resp.Candidates[0].Content.Parts[0].(genai.Text)

	log.Printf("jsonResponse: %s", jsonResponse)

	// var result struct {
	//     Prompt string `json:"prompt"`
	// }

	// strip any new line characters.
	jsonResponseStr := string(jsonResponse)
	// jsonResponseStr = strings.ReplaceAll(jsonResponseStr, "```json", "")
	// jsonResponseStr = strings.ReplaceAll(jsonResponseStr, "```", "")
	// jsonResponseFinal := strings.ReplaceAll(jsonResponseStr, "\n", "")

	// err = json.Unmarshal([]byte(jsonResponseFinal), &result)
	// if err != nil {
	//     return "", fmt.Errorf("error parsing Gemini response: %v", err)
	// }

	return jsonResponseStr, nil
}

func generateDallEPromptForSentence(client *openai.Client, formattedSentence string, video *models.Video) (string, error) {
	if isDevMode() {
		return generateDallEPromptForSentenceGemini(formattedSentence, video)
	}

	topic := video.Topic
	description := video.Description
	style := video.VideoStyle

	// functionDescription := openai.FunctionDefinition{
	//     Name:        "generate_dalle_prompt",
	//     Description: "Generate a DALL-E 3 prompt based on the given sentence and style",
	//     Parameters: json.RawMessage(`{
	//         "type": "object",
	//         "properties": {
	//             "prompt": {
	//                 "type": "string",
	// 				"description": "A detailed, DALL-E friendly prompt that focuses on a single, clear subject. Include specific artistic style, lighting, and mood, but avoid complex scenes or text requests."
	//         },
	//         "required": ["prompt"]
	//     }`),
	// }

	functionDescription := openai.FunctionDefinition{
		Name:        "generate_dalle_prompt",
		Description: "Generate a DALL-E 3 prompt based on the given sentence and style",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"prompt": {
					"type": "string",
					"description": "A detailed, DALL-E friendly prompt that focuses on a single, clear subject. Include specific artistic style, lighting, and mood, but avoid complex scenes or text requests."
				}
			},
			"required": ["prompt"]
		}`),
	}

	styleInstruction := getStyleInstruction(style)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are an AI assistant specialized in creating prompts for DALL-E 3 image generation based on individual sentences from a video script.",
				},
				{
					Role: openai.ChatMessageRoleUser,
					Content: fmt.Sprintf(`Generate detailed DALL-E 3 prompts for each of the following sentences from a video script. 
                    Follow these guidelines for each prompt:
                    1. Focus on a single, clear subject or concept from the sentence.
                    2. Don't request text! Try to not include any text in the prompt. Go after the visual aspect of the sentence.
					3. When you get a single sentence, without much context, Always refer to the full script to understand the context and generate a prompt that fits the overall theme of the video.
					4. Never come up with prompts that show details of a screen. Focus on other aspects of the sentence.
					5. Avoid prompts that ask for specific text or logos, banners etc. Focus on the visual aspect of the prompt.
					6. The images you generate should feel like they belong in a high-quality video. They should be visually appealing and solid.
					7. FOCUS on the ARTISTIC STYLE: %s

					The topic of the video is: %s
					The description of the video is: %s

					The sentence to generate a prompt for is:
                    %s
					`, styleInstruction, topic, description, formattedSentence),
				},
			},
			Functions: []openai.FunctionDefinition{
				functionDescription,
			},
			FunctionCall: openai.FunctionCall{
				Name: "generate_dalle_prompt",
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("error creating chat completion: %v", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from the API")
	}

	functionArgs := resp.Choices[0].Message.FunctionCall.Arguments
	var result struct {
		Prompt string `json:"prompt"`
	}

	err = json.Unmarshal([]byte(functionArgs), &result)
	if err != nil {
		return "", fmt.Errorf("error parsing AI response: %v", err)
	}
	return result.Prompt, nil
}

func getStyleInstruction(style string) string {
	switch style {
	case "anime":
		return "Create the prompt in the style of a high-quality anime key visual, with vibrant colors, dynamic lighting, and attention to fine details. Think of works by Studio Ghibli or Makoto Shinkai."
	case "cartoon":
		return "Design the prompt in the style of a modern, polished cartoon, reminiscent of high-end 3D animated films. Include bold colors, exaggerated features, and a touch of whimsy, similar to works by Pixar or DreamWorks."
	case "watercolor":
		return "Envision the prompt as a delicate watercolor painting, with soft, translucent colors blending seamlessly. Incorporate visible brush strokes and paper texture, inspired by the ethereal works of J.M.W. Turner or the nature studies of Albrecht Dürer."
	case "digital":
		return "Craft the prompt as a cutting-edge digital artwork, with crisp lines, vibrant gradients, and a futuristic feel. Think of works by Beeple or the sleek aesthetics of sci-fi concept art."
	case "vintage":
		return "Frame the prompt as a vintage illustration from the mid-20th century, with slightly faded colors, visible halftone dots, and the charm of retro advertising posters or classic book covers."
	case "minimalist":
		return "Conceptualize the prompt as a minimalist design, focusing on clean lines, negative space, and a limited color palette. Draw inspiration from modern graphic design and abstract art movements."
	case "photorealistic":
		return "Envision the prompt as a hyper-realistic photograph, with incredible detail, dramatic lighting, and perfect composition. Think of high-end editorial photography or the works of photorealistic painters like Chuck Close."
	default:
		return "Create the prompt as a vivid, high-definition digital painting, balancing realism with artistic flair. Include rich textures, dramatic lighting, and a cinematic quality to the composition."
	}
}

func processContentGemini(topic, description string) (string, string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("ACIDRAIN_GEMNI_KEY")))
	if err != nil {
		return "", "", fmt.Errorf("error creating Gemini client: %v", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-pro")

	prompt := fmt.Sprintf(`Process the following content to create a cleaned topic and script for short-form video:

Original topic: %s
Description: %s

You are a script writer for social media to help with TikTok, Instagram, YouTube Shorts, and other short-form video content. Create a cleaned topic and script based on the given topic and description. The script should be engaging and informative, suitable for a 60-80 second video.

Please format your response as a JSON object with the following structure:
{
    "cleaned_topic": "A more attractive and engaging version of the original topic",
    "script": "A 60-80 second script for the video (more than 200 words)"
}

Do not include hashtags, links, emojis, or any guidance on how to shoot the video or camera angles in the script.`, topic, description)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", "", fmt.Errorf("error generating content: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", "", fmt.Errorf("no content generated")
	}

	jsonResponse := resp.Candidates[0].Content.Parts[0].(genai.Text)

	jsonResponseStr := string(jsonResponse)

	// lowercase the response
	jsonResponseStr = strings.ToLower(jsonResponseStr)

	jsonResponseStr = strings.ReplaceAll(jsonResponseStr, "```json", "")
	jsonResponseStr = strings.ReplaceAll(jsonResponseStr, "```", "")
	jsonResponseFinal := strings.ReplaceAll(jsonResponseStr, "\n", "")

	log.Printf("jsonResponse: %s", jsonResponse)

	var result struct {
		CleanedTopic string `json:"cleaned_topic"`
		Script       string `json:"script"`
	}

	err = json.Unmarshal([]byte(jsonResponseFinal), &result)
	if err != nil {
		return "", "", fmt.Errorf("error parsing Gemini response: %v", err)
	}

	return result.CleanedTopic, result.Script, nil
}

func processContent(client *openai.Client, topic, description string) (string, string, error) {
	if isDevMode() {
		return processContentGemini(topic, description)
	}

	functionDescription := openai.FunctionDefinition{
		Name:        "process_content",
		Description: "Process a topic and description to create a cleaned topic and script for short-form video content",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"cleaned_topic": {
					"type": "string",
					"description": "Take the topic given and make it more attractive and engaging. This could involve rephrasing, adding more context, or making it more specific. Aim for a topic that is clear, concise, and interesting to a general audience."
				},
				"script": {
					"type": "string",
					"description": "A script for a 60-80 second video based on the topic and description. Must be engaging and informative, and suitable for a short-form video format. Pretend you are writing a script for a video on Instagram or YouTube Shorts, aiming for growth and engagement. Remember, that this is a script for the reel and NOT the caption. Maintain a tone accordingly. MUST be more than 200 words. DO NOT include any links or hashtags. DO NOT include any guidance on how to shoot the video OR any emojis. Try to be as creative and informative as possible."
				}
			},
			"required": ["cleaned_topic", "script"]
		}`),
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a script writer for social media to help with tiktok, instagram, youtube shorts, and other short-form video content. You have been asked to create a cleaned topic and script for a short-form video based on the following topic and description. Do not include hastags, links, emojis or any guidance on how to shoot the video or the camera angle. The script should be engaging and informative.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf("Process the following content to create a cleaned topic and script for short-form video:\n\nOriginal topic: %s\nDescription: %s", topic, description),
				},
			},
			Functions: []openai.FunctionDefinition{
				functionDescription,
			},
			FunctionCall: openai.FunctionCall{
				Name: "process_content",
			},
		},
	)

	if err != nil {
		return "", "", fmt.Errorf("error creating chat completion: %v", err)
	}

	functionArgs := resp.Choices[0].Message.FunctionCall.Arguments

	var result struct {
		CleanedTopic string `json:"cleaned_topic"`
		Script       string `json:"script"`
	}
	err = json.Unmarshal([]byte(functionArgs), &result)
	if err != nil {
		return "", "", fmt.Errorf("error parsing AI response: %v", err)
	}

	return result.CleanedTopic, result.Script, nil
}
