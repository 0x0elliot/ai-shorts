package util

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"log"
	"path/filepath"
	"strings"
	"time"
	"strconv"
	"bufio"

	"net/http"
	"bytes"
	"encoding/json"

	"cloud.google.com/go/storage"
)

type Subtitle struct {
    Number   int
    Start    time.Time
    End      time.Time
    Duration time.Duration
    Text     string
}

const (
    VIDEO_WIDTH  = 1080
    VIDEO_HEIGHT = 1920
    FONT_SIZE    = 48
)


var tempDir = filepath.Join("/tmp")


func StitchVideo(ctx context.Context, storageClient *storage.Client, bucketName, videoID string) (string, error) {

	var audioPath string
	var subtitlesPath string
	var imagePaths []string

	// Create a temporary directory for working files
	// use tempDir to store the directory path to "/tmp" ONLY
	// if the directory does not exist, create it
	// if _, err := os.Stat(tempDir); os.IsNotExist(err) {
	// 	os.Mkdir(tempDir, 0755)
	// }

	// defer os.RemoveAll(tempDir)

	// download all images ( in the format of image_0.png, image_1.png, etc) from the bucket
	// it := storageClient.Bucket(bucketName).Objects(ctx, &storage.Query{
	// 	Prefix: videoID + "/images/",
	// })

	// get all objects in the bucket "video/{id}/images"
	it := storageClient.Bucket(bucketName).Objects(ctx, &storage.Query{
		Prefix: "videos/" + videoID,
	})

	for {
		attrs, err := it.Next()
		if err != nil {
			// return "", fmt.Errorf("failed to iterate over objects: %v", err)
			if strings.Contains(err.Error(), "no more items in iterator") {
				break
			}

			return "", fmt.Errorf("failed to iterate over objects: %v", err)
		}


		// Download the object
		objectPath := attrs.Name
		objectPath = strings.Replace(objectPath, "gs://"+bucketName+"/", "", 1)
		objectPath = strings.Replace(objectPath, "/", string(filepath.Separator), -1)

		objectName := strings.Split(objectPath, "/")[len(strings.Split(objectPath, "/"))-1]

		outputPath := filepath.Join(tempDir, objectName)

		log.Printf("Downloading object %s to %s", attrs.Name, outputPath)

		// download the object
		reader, err := storageClient.Bucket(bucketName).Object(objectPath).NewReader(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to get object reader: %v", err)
		}

		// Save the image
		imageData, err := ioutil.ReadAll(reader)
		if err != nil {
			return "", fmt.Errorf("failed to read image data: %v", err)
		}

		err = ioutil.WriteFile(outputPath, imageData, 0644)
		if err != nil {
			return "", fmt.Errorf("failed to save image file: %v", err)
		}

		if strings.Contains(objectPath, "audio") {
			audioPath = outputPath
		} else if strings.Contains(objectPath, "subtitles") {
			subtitlesPath = outputPath
		} else {
			// Append the path to the list of image paths
			imagePaths = append(imagePaths, outputPath)
		}
	}

	// now, print all the paths obtained: imagePaths, audioPath, subtitlesPath
	fmt.Println("Image Paths: ", imagePaths)
	fmt.Println("Audio Path: ", audioPath)
	fmt.Println("Subtitles Path: ", subtitlesPath)

	// sort the imagePaths
	imagePaths = sortImagePaths(imagePaths)

	log.Printf("Creating slideshow with subtitles..")

	callStitchingAPI(imagePaths, audioPath, subtitlesPath, filepath.Join(tempDir, "output.mp4"))

	return filepath.Join(tempDir, "output.mp4"), nil
}

func callStitchingAPI(imagePaths []string, audioPath, subtitlesPath, outputPath string) error {
	// @app.route('/create_slideshow', methods=['POST'], from localhost:5000
	req, err := http.NewRequest("POST", "http://localhost:5000/create_slideshow", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	type SlideshowRequest struct {
		ImagePaths    []string `json:"image_paths"`
		AudioPath     string   `json:"audio_path"`
		SubtitlesPath string   `json:"subtitles_path"`
	}

	// Create the request body
	slideshowRequest := SlideshowRequest{
		ImagePaths:    imagePaths,
		AudioPath:     audioPath,
		SubtitlesPath: subtitlesPath,
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
	_, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}

	return nil // for now
}


func parseSRTFile(filename string) ([]Subtitle, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, fmt.Errorf("failed to open SRT file: %w", err)
    }
    defer file.Close()

    var subtitles []Subtitle
    scanner := bufio.NewScanner(file)
    var currentSub Subtitle

    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" {
            if currentSub.Number != 0 {
                subtitles = append(subtitles, currentSub)
                currentSub = Subtitle{}
            }
            continue
        }

        if currentSub.Number == 0 {
            currentSub.Number, _ = strconv.Atoi(line)
        } else if currentSub.Start.IsZero() {
            times := strings.Split(line, " --> ")
            currentSub.Start, _ = time.Parse("15:04:05,000", times[0])
            currentSub.End, _ = time.Parse("15:04:05,000", times[1])
            currentSub.Duration = currentSub.End.Sub(currentSub.Start)
        } else {
            currentSub.Text += line + " "
        }
    }

    if currentSub.Number != 0 {
        subtitles = append(subtitles, currentSub)
    }

    if err := scanner.Err(); err != nil {
        return nil, fmt.Errorf("error reading SRT file: %w", err)
    }

    return subtitles, nil
}

// sort the imagePaths
func sortImagePaths(imagePaths []string) []string {
	// images are in the format of image_0.png, image_1.png, etc
	sorted := make([]string, len(imagePaths))
	for _, path := range imagePaths {
		// get the index
		index := strings.Split(filepath.Base(path), "_")[1]
		// strip the extension
		index = strings.Split(index, ".")[0]
		i, _ := strconv.Atoi(index)

		log.Printf("Sorting image %s at index %d", path, i)
		sorted[i - 1] = path
	}

	return sorted
}

func splitIntoLines(text string, maxChars int) []string {
    var lines []string
    words := strings.Fields(text)
    currentLine := ""

    for _, word := range words {
        if len(currentLine)+len(word) > maxChars {
            lines = append(lines, strings.TrimSpace(currentLine))
            currentLine = word
        } else {
            if currentLine != "" {
                currentLine += " "
            }
            currentLine += word
        }
    }

    if currentLine != "" {
        lines = append(lines, strings.TrimSpace(currentLine))
    }

    return lines
}