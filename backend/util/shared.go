package util

import (
	"encoding/base64"
	"os"
	"regexp"
)

func isDevMode() bool {
	return os.Getenv("USE_GEMINI") == "true"
}

func StripEmoji(s string) string {
	// https://stackoverflow.com/a/13785978/13201408
	var newRunes []rune
	for _, r := range s {
		if r > 0x7F {
			continue
		}
		newRunes = append(newRunes, r)
	}
	return string(newRunes)
}

func Contains(arr []string, str string) bool {

	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func SplitSRTIntoSentences(text string) []string {
	// Define a regular expression to match the SRT blocks
	// Each block contains an index number, a timestamp range, and the actual sentence
	re := regexp.MustCompile(`(?m)^\d+\s+\d{2}:\d{2}:\d{2},\d{3} --> \d{2}:\d{2}:\d{2},\d{3}\s+(.*)$`)

	// Find all matches
	matches := re.FindAllStringSubmatch(text, -1)

	// Extract the sentences
	var sentences []string
	for _, match := range matches {
		if len(match) > 1 {
			sentences = append(sentences, match[1])
		}
	}

	return sentences
}

func ContainsInt64(arr []int64, num int64) bool {
	for _, a := range arr {
		if a == num {
			return true
		}
	}
	return false
}

func IsBase64Image(base64Image string) bool {
	return "data:image/" == base64Image[:11]
}

// CalculateBase64ImageSizeMB takes a base64 encoded string and returns its size in megabytes
func CalculateBase64ImageSizeMB(base64String string) (float64, error) {
	// Decode the base64 string
	data, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return 0, err
	}

	// Calculate the size in bytes
	sizeInBytes := len(data)

	// Convert the size to megabytes (1 MB = 1024 * 1024 bytes)
	sizeInMB := float64(sizeInBytes) / (1024 * 1024)

	return sizeInMB, nil
}
