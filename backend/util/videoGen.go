package util

import (
	"fmt"
	"io/ioutil"
	"log"

	"net/http"
	"bytes"
	"encoding/json"
)

func StitchVideo(videoID string) (error) {
	log.Printf("[INFO] Creating slideshow with subtitles..")

	err := callStitchingAPI(videoID)
	if err != nil {
		return fmt.Errorf("failed to call stitching API: %v", err)
	}

	return nil
}

func callStitchingAPI(videoID string) error {
	// @app.route('/create_slideshow', methods=['POST'], from localhost:5000
	req, err := http.NewRequest("POST", "http://localhost:8080/create_slideshow", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	type SlideshowRequest struct {
		VideoID string `json:"video_id"`
	}

	// Create the request body
	slideshowRequest := SlideshowRequest{
		VideoID: videoID,
	}

	// Marshal the request body
	body, err := json.Marshal(slideshowRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Set the request body
	req.Body = ioutil.NopCloser(bytes.NewReader(body))

	// Set the content type
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		log.Printf("[ERROR] Failed to create slideshow with subtitles. The response body is: %v with status code: %v", res.Body, res.StatusCode)

		return fmt.Errorf("failed to send request: %v", err)
	}

	return nil // for now
}