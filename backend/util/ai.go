package util

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"path"
	"io"
	"bytes"
	"encoding/base64"
	"io/ioutil"

	models "go-authentication-boilerplate/models"

	openai "github.com/sashabaranov/go-openai"
	storage "cloud.google.com/go/storage"
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

func SaveVideoError(video *models.Video, err error) error {
	video.Error = err.Error()
	_, err = SetVideo(video)
	if err != nil {
		return err
	}

	return nil
}

func CreateVideo(video *models.Video) (*models.Video, error) {
	client := openai.NewClient(OPENAI_API_KEY)
	storageClient, err := GetGCPClient("/Users/aditya/Documents/OSS/zappush/shortpro/backend/gcp_credentials.json")
	if err != nil {
		log.Printf("[ERROR] Error creating storage client: %v", err)
		return nil, err
	}

	cleanedtopic, script, err := processContent(client, video.Topic, video.Description)
	if err != nil {
		log.Printf("[ERROR] Error processing content: %v", err)
		return nil, err
	}

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

	prompts, err := generateDallEPromptsForScript(client, video.Script, DefaultStyle)
	if err != nil {
		log.Printf("[ERROR] Error generating DALL-E prompts: %v", err)
		_ = SaveVideoError(video, err)
		return nil, err
	}

	video.Progress = 25
	video.DALLEPromptGenerated = true

	video, err = SetVideo(video)
	if err != nil {
		log.Printf("[ERROR] Error saving video: %v", err)
		_ = SaveVideoError(video, err)
		return nil, err
	}

	log.Printf("Generated %d prompts for the video script", len(prompts))
	log.Printf("prompts: %v", prompts)

	// now, we can create the image
	// Create a new context and storage client
	ctx := context.Background()
	if err != nil {
		log.Printf("[ERROR] Error creating storage client: %v", err)
		_ = SaveVideoError(video, err)
		return nil, err
	}

	bucketName := os.Getenv("ACIDRAIN_GCP_BUCKET_NAME")
	videoID := video.ID

	_, err = generateImagesForScript(ctx, client, storageClient, bucketName, videoID, prompts, DefaultStyle)
	if err != nil {
		log.Printf("[ERROR] Error generating images: %v", err)
		_ = SaveVideoError(video, err)
		return nil, err
	}

	video.Progress = 60
	video.DALLEGenerated = true

	video, err = SetVideo(video)
	if err != nil {
		log.Printf("[ERROR] Error saving video: %v", err)
		_ = SaveVideoError(video, err)
		return nil, err
	}

	_, err = generateTTSForScript(ctx, client, storageClient, bucketName, videoID, video.Script, video.Narrator)
	if err != nil {
		log.Printf("[ERROR] Error generating TTS: %v", err)
		_ = SaveVideoError(video, err)
		return nil, err
	}

	video.Progress = 70
	video.TTSGenerated = true

	video, err = SetVideo(video)
	if err != nil {
		log.Printf("[ERROR] Error saving video: %v", err)
		_ = SaveVideoError(video, err)
		return nil, err
	}

	return video, nil
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
		imageData, err := generateImageForPrompt(client, sp.Prompt, style)
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

func generateImageForPrompt(client *openai.Client, prompt string, style ImageStyle) ([]byte, error) {
	stylePrompt := getStylePrompt(style)
	fullPrompt := fmt.Sprintf("%s %s", stylePrompt, prompt)

	req := openai.ImageRequest{
		Prompt:         fullPrompt,
		Size:           openai.CreateImageSize1024x1024,
		ResponseFormat: openai.CreateImageResponseFormatB64JSON,
		N:              1,
	}

	resp, err := client.CreateImage(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("image creation failed: %v", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no image data received")
	}

	imgBytes, err := base64.StdEncoding.DecodeString(resp.Data[0].B64JSON)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image data: %v", err)
	}

	return imgBytes, nil
}

func getStylePrompt(style ImageStyle) string {
	switch style {
	case AnimeStyle:
		return "Create an anime style image of"
	case CartoonStyle:
		return "Create a cartoon style image of"
	case WatercolorStyle:
		return "Create a watercolor style painting of"
	default:
		return "Create a realistic, detailed image of"
	}
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


func generateDallEPromptsForScript(client *openai.Client, script string, style ImageStyle) ([]SentencePrompt, error) {
    // Strip all emojis away from the script
    script = StripEmoji(script)
    sentences := splitIntoSentences(script)

    functionDescription := openai.FunctionDefinition{
        Name:        "generate_dalle_prompts",
        Description: "Generate DALL-E 3 prompts for multiple sentences",
        Parameters: json.RawMessage(`{
            "type": "object",
            "properties": {
                "prompts": {
                    "type": "array",
                    "items": {
                        "type": "object",
                        "properties": {
                            "sentence": {
                                "type": "string",
                                "description": "The original sentence from the script"
                            },
                            "prompt": {
                                "type": "string",
                                "description": "A detailed prompt for DALL-E 3 to generate an image based on the sentence and style"
                            }
                        },
                        "required": ["sentence", "prompt"]
                    }
                }
            },
            "required": ["prompts"]
        }`),
    }

    styleInstruction := getStyleInstruction(style)
    
    // Prepare the sentences as a formatted string
    formattedSentences := ""
    for i, sentence := range sentences {
        formattedSentences += fmt.Sprintf("%d. %s\n", i+1, sentence)
    }

    resp, err := client.CreateChatCompletion(
        context.Background(),
        openai.ChatCompletionRequest{
            Model: openai.GPT4,
            Messages: []openai.ChatCompletionMessage{
                {
                    Role:    openai.ChatMessageRoleSystem,
                    Content: "You are an AI assistant specialized in creating prompts for DALL-E 3 image generation based on sentences from a video script.",
                },
                {
                    Role: openai.ChatMessageRoleUser,
                    Content: fmt.Sprintf(`Generate detailed DALL-E 3 prompts for each of the following sentences from a video script. 
                        Each prompt should describe a single, striking image that captures the essence of the corresponding sentence. 
                        %s
                        Focus on visual elements, colors, and composition. The prompts should be descriptive but concise.
                        
                        Sentences:
                        %s`, styleInstruction, formattedSentences),
                },
            },
            Functions: []openai.FunctionDefinition{
                functionDescription,
            },
            FunctionCall: openai.FunctionCall{
                Name: "generate_dalle_prompts",
            },
        },
    )

    if err != nil {
        return nil, fmt.Errorf("error creating chat completion: %v", err)
    }

    if len(resp.Choices) == 0 {
        return nil, fmt.Errorf("no choices returned from the API")
    }

    functionArgs := resp.Choices[0].Message.FunctionCall.Arguments

    var result struct {
        Prompts []SentencePrompt `json:"prompts"`
    }
    err = json.Unmarshal([]byte(functionArgs), &result)
    if err != nil {
        return nil, fmt.Errorf("error parsing AI response: %v", err)
    }

    // Ensure the order of prompts matches the order of sentences
    orderedPrompts := make([]SentencePrompt, len(sentences))
    for _, prompt := range result.Prompts {
        for i, sentence := range sentences {
            if prompt.Sentence == sentence {
                orderedPrompts[i] = prompt
                break
            }
        }
    }

    return orderedPrompts, nil
}

func getStyleInstruction(style ImageStyle) string {
	switch style {
	case AnimeStyle:
		return "Create the prompt in an anime art style."
	case CartoonStyle:
		return "Create the prompt in a cartoon art style."
	case WatercolorStyle:
		return "Create the prompt in a watercolor painting style."
	default:
		return "Create the prompt in a realistic, detailed style."
	}
}

func splitIntoSentences(text string) []string {
	// This is a simple sentence splitter. For more accurate results, consider using a natural language processing library.
	sentences := strings.FieldsFunc(text, func(r rune) bool {
		return r == '.' || r == '!' || r == '?'
	})
	
	for i, s := range sentences {
		sentences[i] = strings.TrimSpace(s)
	}
	
	return sentences
}

func processContent(client *openai.Client, topic, description string) (string, string, error) {
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