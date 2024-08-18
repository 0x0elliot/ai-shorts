package util

import (
	"fmt"
	"io/ioutil"
	"log"

	models "go-authentication-boilerplate/models"

	"net/http"
	"bytes"
	"encoding/json"
)

type StitchingAPIResponse struct {
	OutputFile string `json:"output_file"`
}

func StitchVideo(video models.Video) (models.Video, error) {
	log.Printf("[INFO] Creating slideshow with subtitles..")

	videoID := video.ID

	outputURL, err := callStitchingAPI(videoID, video.BackgroundMusic)
	if err != nil {
		return video, fmt.Errorf("failed to call stitching API: %v", err)
	}

	videoPtr, err := GetVideoById(videoID)
	if err != nil {
		return video, fmt.Errorf("failed to get video by ID: %v", err)
	}

	video = *videoPtr

	video.VideoURL = outputURL
	video.VideoUploaded = true
	video.VideoStitched = true
	video.TTSGenerated = true
	video.SRTGenerated = true
	video.ScriptGenerated = true
	video.DALLEPromptGenerated = true
	video.DALLEGenerated = true
	video.StitchedVideoURL = outputURL

	videoPtr, err = SetVideo(&video)
	if err != nil {
		log.Printf("[ERROR] Failed to set video: %v", err)
		return video, fmt.Errorf("failed to set video: %v", err)
	}

	log.Printf("[INFO] Successfully created slideshow with subtitles..")

	video = *videoPtr
	
	return video, nil
}

func callStitchingAPI(videoID string, musicFile string) (outputUrl string, err error) {
	req, err := http.NewRequest("POST", "http://127.0.0.1:8080/create_slideshow", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	type SlideshowRequest struct {
		VideoID string `json:"video_id"`
		MusicFile string `json:"music"`
	}

	log.Printf("[INFO] Music file: %v", musicFile)

	// Create the request body
	slideshowRequest := SlideshowRequest{
		VideoID: videoID,
		MusicFile: musicFile,
	}

	// Marshal the request body
	body, err := json.Marshal(slideshowRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Set the request body
	req.Body = ioutil.NopCloser(bytes.NewReader(body))

	// Set the content type
	req.Header.Set("Content-Type", "application/json")

	log.Printf("[INFO] Sending request to create slideshow with subtitles..")

	// Send the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Printf("[ERROR] Failed to create slideshow with subtitles. The response body is: %v with status code: %v", res.Body, res.StatusCode)
		return "", fmt.Errorf("failed to send request: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		log.Printf("[ERROR] Failed to create slideshow with subtitles. The response body is: %v with status code: %v", res.Body, res.StatusCode)
		return "", fmt.Errorf("failed to send request: %v", err)
	}

	log.Printf("[INFO] Successfully created slideshow with subtitles..")

	// Decode the response
	var response StitchingAPIResponse

	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		log.Printf("[ERROR] Failed to decode response: %v", res.Body)
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("[INFO] Output file: %v", response.OutputFile)

	return response.OutputFile, nil
}