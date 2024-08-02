package util 

import (
	"encoding/base64"
	"strings"
)

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

// user agent to device type
func GetUserAgentDeviceType(userAgent string) string {
	if userAgent == "" {
		return "unknown"
	}

	if strings.Contains(userAgent, "Android") {
		return "mobile"
	}

	if strings.Contains(userAgent, "iPhone") {
		return "mobile"
	}

	if strings.Contains(userAgent, "iPad") {
		return "tablet"
	}

	if strings.Contains(userAgent, "Macintosh") {
		return "desktop"
	}

	if strings.Contains(userAgent, "Windows") {
		return "desktop"
	}

	if strings.Contains(userAgent, "Linux") {
		return "desktop"
	}

	return "unknown"
}

// User agent to OS
func GetUserAgentOS(userAgent string) string {
	if userAgent == "" {
		return "unknown"
	}

	if strings.Contains(userAgent, "Android") {
		return "Android"
	}

	if strings.Contains(userAgent, "iPhone") {
		return "iOS"
	}

	if strings.Contains(userAgent, "iPad") {
		return "iOS"
	}

	if strings.Contains(userAgent, "Macintosh") {
		return "MacOS"
	}

	if strings.Contains(userAgent, "Windows") {
		return "Windows"
	}

	if strings.Contains(userAgent, "Linux") {
		return "Linux"
	}

	return "unknown"
}
