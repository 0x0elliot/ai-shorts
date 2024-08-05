package util

import (
	"encoding/base64"
	"os"
	"strings"
	"bufio"
)

type ASRSentences struct {
	// End int `json:"end"`
	// Start int `json:"start"`
	Text string `json:"text"`
}

type ASR struct {
	Sentences []ASRSentences `json:"sentences"`
}

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

func SplitScriptASRIntoSentences(sentences []ASRSentences) []string {
	var result []string
	for _, sentence := range sentences {
		// skip the first character of the sentence
		// because it is always a space
		result = append(result, sentence.Text[1:])
	}
	return result
}

func SplitSRTIntoSentences(wordLevelSRT string) []string {
    var sentences []string
    var currentSentence []string

    scanner := bufio.NewScanner(strings.NewReader(wordLevelSRT))
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        
        // Skip empty lines and timestamp lines
        if line == "" || strings.Contains(line, "-->") {
            continue
        }
        
        // Add word to current sentence
        currentSentence = append(currentSentence, line)
        
        // Check if the word ends a sentence
        if strings.HasSuffix(line, ".") || strings.HasSuffix(line, "!") || strings.HasSuffix(line, "?") {
            sentences = append(sentences, strings.Join(currentSentence, " "))
            currentSentence = nil
        }
    }

    // Add any remaining words as a sentence
    if len(currentSentence) > 0 {
        sentences = append(sentences, strings.Join(currentSentence, " "))
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
