package models

import (
	pq "github.com/lib/pq"
)

// later, break this into two models: Video and Schedule
type Video struct {
	Base
	Topic         string         `json:"topic"`
	Description   string         `json:"description"`
	Narrator      string         `json:"narrator"`
	VideoStyle    string         `json:"videoStyle"`
	PostingMethod pq.StringArray `json:"postingMethod" gorm:"type:text[]"`
	IsOneTime     bool           `json:"isOneTime"`
	VideoURL      string         `json:"videoURL" gorm:"null"`
	Script        string         `json:"script" gorm:"null"`
	VideoTheme    string         `json:"videoTheme"`

	ScriptGenerated      bool `json:"scriptGenerated" gorm:"default:false"`
	DALLEPromptGenerated bool `json:"dallePromptGenerated" gorm:"default:false"`
	DALLEGenerated       bool `json:"dalleGenerated" gorm:"default:false"`
	TTSGenerated         bool `json:"ttsGenerated" gorm:"default:false"`
	SRTGenerated         bool `json:"srtGenerated" gorm:"default:false"`
	VideoStitched        bool `json:"videoStitched" gorm:"default:false"`
	// full progress of the video
	VideoUploaded  bool `json:"videoUploaded" gorm:"default:false"`

	Progress int `json:"progress" gorm:"default:0"`

	MediaType string `json:"mediaType" gorm:"default:ai"` // ai or stock (from pexels)

	Essence string `json:"essence" gorm:"null"` // the essence of the video

	BackgroundMusic string `json:"backgroundMusic" gorm:"null"`

	// maybe, hide this from the user
	Error string `json:"error" gorm:"null"`

	TTSURL           string `json:"ttsURL" gorm:"null"`
	SRTURL           string `json:"srtURL" gorm:"null"`
	StitchedVideoURL string `json:"stitchedVideoURL" gorm:"null"`

	OwnerID string `json:"ownerID"`
	Owner   User   `json:"owner" gorm:"foreignKey:OwnerID;references:ID"`
}
