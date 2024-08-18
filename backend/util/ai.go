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
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"sync"
	"mime/multipart"

	models "go-authentication-boilerplate/models"

	"github.com/anthropics/anthropic-sdk-go"
	anthropicOpts "github.com/anthropics/anthropic-sdk-go/option"

	openai "github.com/sashabaranov/go-openai"
	"github.com/google/generative-ai-go/genai"
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

type PromptResponse struct {
	DallePrompt string `json:"dalle_prompt"`
}

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

// type ASR struct {
// 	Sentences []struct {
// 		Text string `json:"text"`
// 	} `json:"sentences"`
// }


func getVideoFolderPath(videoID string) string {
	return filepath.Join(os.Getenv("HOME"), "Desktop", "reels", videoID)
}

func SaveVideoError(video *models.Video, err error) error {
	video.Error = err.Error()
	_, err = SetVideo(video)
	return err
}

func CreateVideo(video *models.Video, recreate bool) (*models.Video, error) {
	startTime := time.Now()

	client := openai.NewClient(OPENAI_API_KEY)

	if recreate {
		video.Error = ""
		// if video.DALLEPromptGenerated && video.DALLEGenerated && video.TTSGenerated {
		// 	log.Printf("[INFO] Video already processed. Let's try to stitch it again: %s", video.ID)
		// 	videoPtr, err := StitchVideo(*video);
		// 	if err != nil {
		// 		log.Printf("[ERROR] Error stitching video: %v", err)
		// 		return nil, SaveVideoError(video, err)
		// 	}

		// 	video = &videoPtr

		// 	// Update video progress
		// 	video.Progress = 100
		// 	video.VideoStitched = true

		// 	video, err := SetVideo(video)
		// 	if err != nil {
		// 		log.Printf("[ERROR] Error saving video: %v", err)
		// 		return nil, SaveVideoError(video, err)
		// 	}
		// }

		folderPath := getVideoFolderPath(video.ID)
		if err := os.RemoveAll(folderPath); err != nil {
			log.Printf("[ERROR] Error deleting folder: %v", err)
		}

		if err := os.MkdirAll(folderPath, 0755); err != nil {
			log.Printf("[ERROR] Error creating folder: %v", err)
			return nil, err
		}

		video.Progress = 0
		video.ScriptGenerated = false
		video.DALLEPromptGenerated = false
		video.DALLEGenerated = false
		video.TTSGenerated = false
		video.VideoStitched = false
		video.SRTGenerated = false
		video.VideoUploaded = false
		video.Error = ""
		video.TTSURL = ""
		video.StitchedVideoURL = ""

		var err error
		video, err = SetVideo(video)
		if err != nil {
			log.Printf("[ERROR] Error saving video: %v", err)
			return nil, err
		}
	}

	log.Printf("[INFO] Processing content for video: %s", video.ID)

	// cleanedTopic, script, essence, err := processContent(client, video.Topic, video.Description)
	cleanedTopic, script, essence, err := GenerateScriptClaude(video.Topic, video.Description)
	if err != nil {
		log.Printf("[ERROR] Error processing content: %v", err)
		return nil, SaveVideoError(video, err)
	}

	log.Printf("[INFO] Processed content for video: %s", video.ID)

	video.Topic = cleanedTopic
	video.Script = script
	video.Essence = essence
	video.ScriptGenerated = true
	video.Progress = 10

	video, err = SetVideo(video)
	if err != nil {
		log.Printf("[ERROR] Error saving video: %v", err)
		return nil, SaveVideoError(video, err)
	}

	log.Printf("[INFO] Generating TTS for video: %s", video.ID)

	if err := generateTTSForScript(client, video); err != nil {
		log.Printf("[ERROR] Error generating TTS: %v", err)
		return nil, SaveVideoError(video, err)
	}

	log.Printf("[INFO] Generated TTS for video: %s", video.ID)

	video.Progress = 30
	video.TTSGenerated = true

	video, err = SetVideo(video)
	if err != nil {
		log.Printf("[ERROR] Error saving video: %v", err)
		return nil, SaveVideoError(video, err)
	}

	log.Printf("[INFO] Generating SRT for video: %s", video.ID)

	asrSentences, err := generateSRTForTTSTranscript(video)
	if err != nil {
		log.Printf("[ERROR] Error generating SRT: %v", err)
		return nil, SaveVideoError(video, err)
	}

	log.Printf("[INFO] Generated SRT for video: %s", video.ID)

	video.Progress = 50
	video.SRTGenerated = true
	video.SRTURL = filepath.Join(getVideoFolderPath(video.ID), "subtitles", "subtitles.json")

	video, err = SetVideo(video)
	if err != nil {
		log.Printf("[ERROR] Error saving video: %v", err)
		return nil, SaveVideoError(video, err)
	}

	forceAI := false

	if video.MediaType == "stock" {
		pexelsVideos, err := fetchPexelsVideos(*video, asrSentences)
		if err != nil {
			log.Printf("[ERROR] Error fetching Pexels videos: %v", err)
			forceAI = true
		} else {
			log.Printf("[INFO] Fetched %d Pexels videos", len(pexelsVideos))

			selectedVideos, err := matchVideosToSentences(pexelsVideos, asrSentences)
			if err != nil {
				log.Printf("[ERROR] Error matching videos to sentences: %v", err)
				forceAI = true
			} else {
				log.Printf("[INFO] Matched %d videos to sentences", len(selectedVideos))

				// one must try to download these as well soon. 
				// too tired to try.

				video.MediaType = "stock"
				video.Progress = 60
				video, err = SetVideo(video)
				if err != nil {
					log.Printf("[ERROR] Error saving video: %v", err)
					return nil, SaveVideoError(video, err)
				}
			}
		}
	} 

	if forceAI || video.MediaType == "ai" {
		log.Printf("[INFO] Generating images (after generating prompt for each sentence) for video: %s", video.ID)

		err = generateAndSaveImagesForScript(client, video)
		if err != nil {
			log.Printf("[ERROR] Error generating images: %v", err)
			return nil, SaveVideoError(video, err)
		}

		video.Progress = 80
		video.DALLEGenerated = true
		video.DALLEPromptGenerated = true

		video, err = SetVideo(video)
		if err != nil {
			log.Printf("[ERROR] Error saving video: %v", err)
			return nil, SaveVideoError(video, err)
		}

		log.Printf("[INFO] Generated images for video: %s", video.ID)
	}

	log.Printf("[INFO] Going to try to stitch video now: %s", video.ID)

	videoPtr, err := StitchVideo(*video)
	if err != nil {
		log.Printf("[ERROR] Error stitching video: %v", err)
		return nil, SaveVideoError(video, err)
	}

	video = &videoPtr

	log.Printf("[INFO] Stitched video for video: %s", video.ID)

	video.Progress = 100
	video.VideoStitched = true

	video, err = SetVideo(video)
	if err != nil {
		log.Printf("[ERROR] Error saving video: %v", err)
		return nil, SaveVideoError(video, err)
	}

	endTime := time.Now()

	log.Printf("[INFO] Video processing completed in %v", endTime.Sub(startTime))

	return video, nil
}

func fetchPexelsVideos(video models.Video, asrSentences []ASRSentences) ([]PexelsVideo, error) {
	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.pexels.com/videos/search", nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", os.Getenv("PEXELS_API_KEY"))
	req.Header.Set("Accept", "application/json")

	query := url.Values{}
	query.Add("query", video.Essence)
	query.Add("per_page", fmt.Sprintf("%d", len(asrSentences)*2)) // Fetch more videos than sentences for better matching
	req.URL.RawQuery = query.Encode()

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var pexelsResponse PexelsResponse
	err = json.Unmarshal(body, &pexelsResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}

	return pexelsResponse.Videos, nil
}

func matchVideosToSentences(videos []PexelsVideo, sentences []ASRSentences) ([]PexelsVideo, error) {
	if len(videos) == 0 {
		return nil, fmt.Errorf("no videos fetched from Pexels")
	}

	// Calculate total duration of all sentences
	totalSentenceDuration := sentences[len(sentences)-1].End - sentences[0].Start

	// Sort videos by duration, closest to total sentence duration first
	// sort.Slice(videos, func(i, j int) bool {
	// 	return math.Abs(float64(videos[i].Duration*1000-totalSentenceDuration)) <
	// 		math.Abs(float64(videos[j].Duration*1000-totalSentenceDuration))
	// })

	// Select videos that cover the total duration
	var selectedVideos []PexelsVideo
	currentDuration := float64(0)
	for _, video := range videos {
		if currentDuration >= totalSentenceDuration {
			break
		}
		selectedVideos = append(selectedVideos, video)
		currentDuration += float64(video.Duration) * float64(1000) // Convert to milliseconds
	}

	return selectedVideos, nil
}

func generateTTSForScript(client *openai.Client, video *models.Video) error {
	audioData, err := generateTTSForFullScript(client, video.Script, video.Narrator)
	if err != nil {
		return fmt.Errorf("error generating TTS for script: %v", err)
	}

	folderPath := filepath.Join(getVideoFolderPath(video.ID), "audio")
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return fmt.Errorf("error creating audio folder: %v", err)
	}

	filename := "full_audio.mp3"
	filePath := filepath.Join(folderPath, filename)

	return ioutil.WriteFile(filePath, audioData, 0644)
}

func generateTTSForFullScript(client *openai.Client, script, narrator string) ([]byte, error) {
	req := openai.CreateSpeechRequest{
		Model: openai.TTSModel1HD,
		Input: script,
		Voice: openai.VoiceAlloy,
	}

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
	defer resp.Close()

	return io.ReadAll(resp)
}

func generateSRTForTTSTranscript(video *models.Video) ([]ASRSentences, error) {
	audioFilePath := filepath.Join(getVideoFolderPath(video.ID), "audio", "full_audio.mp3")

	asrSentences := []ASRSentences{}

	srtContent, err := generateSRTWithWhisper(audioFilePath, video.Script)
	if err != nil {
		return asrSentences, fmt.Errorf("error generating SRT with Whisper: %v", err)
	}

	srtFolderPath := filepath.Join(getVideoFolderPath(video.ID), "subtitles")
	if err := os.MkdirAll(srtFolderPath, 0755); err != nil {
		return asrSentences, fmt.Errorf("error creating subtitles folder: %v", err)
	}

	srtFilePath := filepath.Join(srtFolderPath, "subtitles.json")
	err = ioutil.WriteFile(srtFilePath, []byte(srtContent), 0644)
	if err != nil {
		return asrSentences, fmt.Errorf("error writing SRT file: %v", err)
	}

	var asr ASR
	err = json.Unmarshal([]byte(srtContent), &asr)
	if err != nil {
		log.Printf("[ERROR] Error unmarshalling ASR content: %v", err)
		return asrSentences, err
	}

	asrSentences = asr.Sentences

	return asrSentences, err
}

func generateSRTWithWhisper(audioFilePath string, script string) (string, error) {
	file, err := os.Open(audioFilePath)
	if err != nil {
		return "", fmt.Errorf("error opening audio file: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("audio", filepath.Base(audioFilePath))
	if err != nil {
		return "", fmt.Errorf("error creating form file: %v", err)
	}
	if _, err = io.Copy(part, file); err != nil {
		return "", fmt.Errorf("error copying file to form: %v", err)
	}

	_ = writer.WriteField("original_script", script)

	writer.Close()
	

	req, err := http.NewRequest("POST", "http://localhost:5000/generate_asr", body)
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	srtContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %v", err)
	}

	return string(srtContent), nil
}

func generateImageForPrompt(prompt string, style ImageStyle, numImages int) ([]byte, error) {
	fullPrompt := prompt

	apiKey := os.Getenv("ACIDRAIN_OLA_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ACIDRAIN_OLA_KEY environment variable not set")
	}

	log.Printf("Generating image for prompt: %s", fullPrompt)

	seed := 1075943719
	seedPtr := &seed

	reqBody := SDXLRequest{
		ModelName:         "diffusion1XL",
		Prompt:            fullPrompt,
		ImageHeight:       1024,
		ImageWidth:        1024,
		NumOutputImages:   numImages,
		GuidanceScale:     10,
		NumInferenceSteps: 50,
		Seed:              seedPtr,
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

	body, err := io.ReadAll(resp.Body)
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

	imageData, err := base64.StdEncoding.DecodeString(sdxlResp.Data[0].B64JSON)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 image data: %v", err)
	}

	return imageData, nil
}

func generateAndSaveImagesForScript(client *openai.Client, video *models.Video) error {
	srtFilePath := filepath.Join(getVideoFolderPath(video.ID), "subtitles", "subtitles.json")
	srtContent, err := ioutil.ReadFile(srtFilePath)
	if err != nil {
		return fmt.Errorf("error reading SRT file: %v", err)
	}
	var asr ASR
	err = json.Unmarshal(srtContent, &asr)
	if err != nil {
		return fmt.Errorf("error unmarshalling ASR content: %v", err)
	}
	sentences := SplitScriptASRIntoSentences(asr.Sentences)
	var wg sync.WaitGroup
	errorChan := make(chan error, len(sentences))
	// Semaphore to limit the number of concurrent goroutines
	semaphore := make(chan struct{}, 20) // Adjust this number based on your needs and API rate limits
	folderPath := filepath.Join(getVideoFolderPath(video.ID), "images")
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return fmt.Errorf("error creating images folder: %v", err)
	}

	// retryDelays := []time.Duration{5 * time.Second, 10 * time.Second, 15 * time.Second}
	retryDelays := []time.Duration{
		5 * time.Second,
	}

	for i := 1; i <= 30; i++ {
		retryDelays = append(retryDelays, time.Duration(i*10)*time.Second)
	}

	for i, sentence := range sentences {
		wg.Add(1)
		go func(index int, s string) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }() // Release semaphore

			var prompt string
			var imageData []byte
			var err error

			lastSentence := ""

			if len(sentences) > 1 {
				lastSentence = sentences[len(sentences)-1]
			}

			// Retry loop for prompt generation
			for retryCount := 0; retryCount <= len(retryDelays); retryCount++ {
				prompt, err = generateDallEPromptForSentence(client, s, video, lastSentence)
				if err == nil {
					break
				}
				if retryCount < len(retryDelays) {
					log.Printf("Error generating prompt for sentence '%s', retrying in %v: %v", s, retryDelays[retryCount], err)
					time.Sleep(retryDelays[retryCount])
				}
			}
			if err != nil {
				errorChan <- fmt.Errorf("[ERROR] failed to generate prompt for sentence '%s' after all retries: %v", s, err)
				return
			}

			// Retry loop for image generation
			for retryCount := 0; retryCount <= len(retryDelays); retryCount++ {
				imageData, err = generateImageForPrompt(prompt, ImageStyle(video.VideoStyle), 1)
				if err == nil {
					break
				}
				if retryCount < len(retryDelays) {
					log.Printf("[ERROR] Error generating image for prompt %d, retrying in %v: %v", index+1, retryDelays[retryCount], err)
					time.Sleep(retryDelays[retryCount])
				}
			}
			if err != nil {
				errorChan <- fmt.Errorf("failed to generate image for prompt %d after all retries: %v", index+1, err)
				return
			}

			// Save image
			filename := fmt.Sprintf("image_%d.png", index+1)
			filePath := filepath.Join(folderPath, filename)
			if err := ioutil.WriteFile(filePath, imageData, 0644); err != nil {
				errorChan <- fmt.Errorf("error saving image %d: %v", index+1, err)
				return
			}
		}(i, sentence)
	}
	wg.Wait()
	close(errorChan)

	// Check for errors
	var errors []string
	for err := range errorChan {
		if err != nil {
			errors = append(errors, err.Error())
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("errors occurred during image generation: %s", strings.Join(errors, "; "))
	}
	return nil
}

func generateDallEPromptForSentence(client *openai.Client, formattedSentence string, video *models.Video, lastSentence string) (string, error) {
	if isDevMode() {
		return generateDallEPromptForSentenceGemini(formattedSentence, video, lastSentence)
	}

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

	styleInstruction := getStyleInstruction(video.VideoStyle)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are an AI assistant specialized in creating prompts for DALL-E 3 image generation based on individual sentences from a video script.",
				},
				{
					Role: openai.ChatMessageRoleUser,
					Content: fmt.Sprintf(`Generate detailed SDXL prompts for each of the following sentences from a video script.
                    Follow these guidelines for each prompt:
					Emphasize visual elements and atmosphere rather than literal interpretation of the sentence.
					IF Talking about a person, PLEASE instruct the prompt to keep their mouth closed.
					Use rich, descriptive language to convey mood, lighting, and textures.
					Incorporate specific artistic styles, techniques, or historical art movements mentioned in the style instruction.
					Try to convert any concerning "sexually explicit" content into a more general and safe-for-work context. Don't just remove the content, but try to replace it with something that fits the context.
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

					Most importantly, your prompt must start with the artistic style: %s

					The topic of the video is: %s
					The description of the video is: %s

					The sentence to generate a prompt for is:
                    %s
					`, styleInstruction, video.Topic, video.Description, formattedSentence),
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
		log.Printf("Result from AI: %s", functionArgs)
		return "", fmt.Errorf("error parsing AI response: %v", err)
	}
	return result.Prompt, nil
}

// bad function
func generateDallEPromptsWithClaude(client *anthropic.Client, sentences []string, video *models.Video) ([]string, error) {
	var lineBrokenSentences string
	for _, sentence := range sentences {
		lineBrokenSentences += sentence + "\n"
	}

	systemMessage := `You are an expert in creating visually appealing and creative DALL-E prompts. Your goal is to generate prompts that result in fun, pretty, and engaging images. Focus on visual elements, atmosphere, and artistic style rather than literal interpretations. Maintain consistency across the entire video narrative.`

	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(fmt.Sprintf(`Generate SDXL prompts for a video with the following details:
Topic: %s
Essence: %s
Video Description: %s
Style: %s

I will provide you with sentences from the video script. For each sentence, generate a SDXL prompt that captures the essence of that part of the video. Maintain consistency across all prompts to create a cohesive visual narrative.

Guidelines for crafting the prompts:
1. Create visually striking and cohesive images that capture the essence of each sentence and the overall video.
2. Use rich, descriptive language to convey mood, lighting, and textures.
3. Incorporate the specified artistic style and techniques.
4. Avoid requesting text, specific logos, or sexually explicit content.
5. Use a format like "[Subject], [Setting], [Mood/Atmosphere], [Style], [Additional details]".
6. Include relevant details from the topic and description to enhance context.
7. Specify camera angles, perspectives, or composition when appropriate.
8. Mention color palettes or lighting conditions that fit the overall theme and style.
9. Include details about materials, textures, or surface qualities.
10. Use evocative adjectives and sensory language to make the prompts more vivid.
11. Specify the desired level of detail or realism.
12. Keep each prompt concise but descriptive, aiming for 1-2 sentences maximum.
13. Ensure consistency across all prompts to create a cohesive visual narrative.
14. IF talking about a person, instruct to keep their mouth closed.

For each sentence I provide, respond with an list containing the SDXL prompt (Must be a simple list. No key/value pairs for EACH sentence. No bullets, or numbering, I will be parsing your response directly. Nothing other than prompts themselves, not even a paragraph briefing on what you're doing. The same number of responses as the sentences provided is very important, even if the line is 1-2 words. Use the context from the last sentence in that case, but respect the rule of equal number of responses). Each sentence is below, separated by a line break:
%s
`, video.Topic, video.Essence, video.Description, getStyleInstruction(video.VideoStyle), sentences))),
	}

	var prompts []string

	message, err := client.Messages.New(context.TODO(), anthropic.MessageNewParams{
		Model:     anthropic.F(anthropic.ModelClaude_3_5_Sonnet_20240620),
		MaxTokens: anthropic.Int(1024),
		System: anthropic.F([]anthropic.TextBlockParam{
			anthropic.NewTextBlock(systemMessage),
		}),
		Messages: anthropic.F(messages),
	})

	if err != nil {
		return nil, fmt.Errorf("error generating content with Claude: %v", err)
	}

	if len(message.Content) == 0 || message.Content[0].Type != "text" {
		return nil, fmt.Errorf("unexpected response format from Claude")
	}

	// print the response in string
	log.Println(message.Content[0].Text)

	// Split the response into prompts
	prompts = strings.Split(message.Content[0].Text, "\n")

	// make sure to remove any empty prompts
	var cleanedPrompts []string
	for _, prompt := range prompts {
		if prompt != "" {
			cleanedPrompts = append(cleanedPrompts, prompt)
		}
	}

	if len(cleanedPrompts) != len(sentences) {
		log.Printf("Sentences: %d, Prompts: %d", len(sentences), len(cleanedPrompts))
	
		return nil, fmt.Errorf("number of prompts generated does not match number of sentences")
	}

	return cleanedPrompts, nil
}

func generateDallEPromptForSentenceGemini(formattedSentence string, video *models.Video, lastSentence string) (string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		return "", fmt.Errorf("error creating Gemini client: %v", err)
	}
	defer client.Close()

	// model := client.GenerativeModel("gemini-pro")
	// use gemini flash
	model := client.GenerativeModel("gemini-1.5-flash")

	styleInstruction := getStyleInstruction(video.VideoStyle)

	prompt := fmt.Sprintf(`Generate a detailed SDXL prompt based on the following information:

Focus on the sentence, last sentence, topic and theme given by the user.

Sentence: %s
Topic: %s
Last sentence (Please use this information to create a coherent narrative): %s
Essence of the video (Please use this very loosely): %s 

Style Instruction: %s
Guidelines for crafting the prompt:
The sentence given to you is almost a sentence. Try to understand the context and generate a prompt that is visually appealing and creatively.

Emphasize visual elements and atmosphere rather than literal interpretation of the sentence.
IF Talking about a person, PLEASE instruct the prompt to keep their mouth closed.
Use rich, descriptive language to convey mood, lighting, and textures.
Incorporate specific artistic styles, techniques, or historical art movements mentioned in the style instruction.
Avoid requesting text or specific logos. Focus on creating a vivid scene or concept.
Try to convert any concerning "sexually explicit" content into a more general and safe-for-work context. Don't just remove the content, but try to replace it with something that fits the context.
Use a format like "[Subject], [Setting], [Mood/Atmosphere], [Style], [Additional details]" to structure the prompt.
Include relevant details from the topic and description to enhance context, but prioritize visual appeal.
Specify camera angles, perspectives, or composition when appropriate (e.g. "close-up view", "wide-angle shot", "birds-eye perspective").
Mention color palettes or lighting conditions that fit the overall theme and style.
Include details about materials, textures, or surface qualities to enhance realism or artistic effect.
Use evocative adjectives and sensory language to make the prompt more vivid. Be creative like a person who wants to grab your attention.
Specify the desired level of detail or realism (e.g. "photorealistic", "impressionistic", "highly detailed").
Avoid unnecessary formatting or markdown. Present the prompt as plain text.
Keep the prompt concise but descriptive, aiming for 2-3 sentences maximum.
Focus on creating a cohesive, visually striking image that captures the essence of the sentence and context.
`, formattedSentence, video.Topic, video.Essence, lastSentence, styleInstruction)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("error generating content: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	jsonResponse := resp.Candidates[0].Content.Parts[0].(genai.Text)
	jsonResponseStr := string(jsonResponse)

	return jsonResponseStr, nil
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
		return "Create a prompt for an ultra-realistic image with high detail, vivid colors, and dramatic lighting. The style should be photorealistic, similar to high-end editorial photography or the works of photorealistic painters like Chuck Close."
	}
}

func processContent(client *openai.Client, topic, description string) (string, string, string, error) {
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
				},
				"essence": {
					"type": "string",
					"description": "1-2 word essence of the video for the stock footage"
				}
			},
			"required": ["cleaned_topic", "script", "essence"]
		}`),
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a script writer for social media to help with tiktok, instagram, youtube shorts, and other short-form video content. You have been asked to create a cleaned topic and script for a short-form video based on the following topic and description. Do not include hashtags, links, emojis or any guidance on how to shoot the video or the camera angle. The script should be engaging and informative.",
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
		return "", "", "", fmt.Errorf("error creating chat completion: %v", err)
	}

	functionArgs := resp.Choices[0].Message.FunctionCall.Arguments
	// clean functionArgs of \n
	functionArgs = strings.ReplaceAll(functionArgs, "\n", "")

	var result struct {
		CleanedTopic string `json:"cleaned_topic"`
		Script       string `json:"script"`
		Essence 	string `json:"essence"`
	}

	err = json.Unmarshal([]byte(functionArgs), &result)
	if err != nil {
		log.Printf("Result from AI: %s", functionArgs)
		return "", "", "", fmt.Errorf("error parsing AI response: %v", err)
	}

	return result.CleanedTopic, result.Script, result.Essence, nil
}

func GenerateScriptClaude(topic, description string) (string, string, string, error) {
	client := anthropic.NewClient(
		anthropicOpts.WithAPIKey(
			os.Getenv("ANTHROPIC_API_KEY"),
		),
	)

	systemMessage := "You are a good script writer for social media reels. Have a personality that reflects in your writing. According to the information provided, choose a personality which is professional or funny or charming or casual. Or a mix of these"

	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(fmt.Sprintf(`Process the following content to create a cleaned topic and script for short-form video:

Original topic: %s
Description: %s

Create a cleaned topic and script based on the given topic and description. The script should be engaging and informative and have 150-170 words.

Format your response as a JSON object with the following structure:
{
    "cleaned_topic": "A more attractive and engaging version of the original topic",
    "script": "A 60-80 second script for the video (150-170 words)."
    "essence": "1-2 word essence of the video for the stock footage"
}

Do not include hashtags, links, emojis, or any guidance on how to shoot the video or camera angles in the script.`, topic, description))),
	}

	message, err := client.Messages.New(context.TODO(), anthropic.MessageNewParams{
		Model:     anthropic.F(anthropic.ModelClaude_3_5_Sonnet_20240620),
		MaxTokens: anthropic.Int(1024),
		System: anthropic.F([]anthropic.TextBlockParam{
			anthropic.NewTextBlock(systemMessage),
		}),
		Messages: anthropic.F(messages),
	})
	if err != nil {
		return "", "", "", fmt.Errorf("error generating content with Claude: %v", err)
	}

	var result struct {
		CleanedTopic string `json:"cleaned_topic"`
		Script       string `json:"script"`
		Essence      string `json:"essence"`
	}

	if len(message.Content) == 0 || message.Content[0].Type != "text" {
		return "", "", "", fmt.Errorf("unexpected response format from Claude")
	}

	err = json.Unmarshal([]byte(message.Content[0].Text), &result)
	if err != nil {
		return "", "", "", fmt.Errorf("error parsing Claude response: %v", err)
	}

	return result.CleanedTopic, result.Script, result.Essence, nil
}

func processContentGemini(topic, description string) (string, string, string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		return "", "", "", fmt.Errorf("error creating Gemini client: %v", err)
	}
	defer client.Close()

	// model := client.GenerativeModel("gemini-pro")
	model := client.GenerativeModel("gemini-1.5-flash")

	prompt := fmt.Sprintf(`Process the following content to create a cleaned topic and script for short-form video:

Original topic: %s
Description: %s

You are a script writer for social media to help with TikTok, Instagram, YouTube Shorts, and other short-form video content. Create a cleaned topic and script based on the given topic and description. The script should be engaging and informative and have 140-170 words.

Make sure to not include any [ ] or any guidance on how to shoot the video or camera angles in the script.

Please format your response as a JSON object with the following structure: (Make sure to remove the backticks, JSON formatting, new lines, and double quotes) 
{
    "cleaned_topic": "A more attractive and engaging version of the original topic",
    "script": "A 60-80 second script for the video (more than 200 words)"
	"essence": "1-2 word essence of the video for the stock footage"
}

Do not include hashtags, links, emojis, or any guidance on how to shoot the video or camera angles in the script.`, topic, description)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", "", "", fmt.Errorf("error generating content: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", "", "", fmt.Errorf("no content generated")
	}

	jsonResponse := resp.Candidates[0].Content.Parts[0].(genai.Text)


	// jsonResponse := fmt.Sprintf(`{
	// 	"cleaned_topic": "%s",
	// 	"script": "%s"
	// }`, topic, description)


	jsonResponseStr := string(jsonResponse)
	jsonResponseStr = strings.ToLower(jsonResponseStr)
	jsonResponseStr = strings.ReplaceAll(jsonResponseStr, "```json", "")
	jsonResponseStr = strings.ReplaceAll(jsonResponseStr, "```", "")
	jsonResponseFinal := strings.ReplaceAll(jsonResponseStr, "\n", "")

// 	log.Printf("jsonResponse: %s", jsonResponse)

	var result struct {
		CleanedTopic string `json:"cleaned_topic"`
		Script       string `json:"script"`
		Essence	  string `json:"essence"`
	}

	err = json.Unmarshal([]byte(jsonResponseFinal), &result)
	if err != nil {
		return "", "", "", fmt.Errorf("error parsing Gemini response: %v", err)
	}

	return result.CleanedTopic, result.Script, result.Essence, nil
}