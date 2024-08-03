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

	ScriptGenerated bool `json:"scriptGenerated" gorm:"default:false"`
	DALLEPromptGenerated bool `json:"dallePromptGenerated" gorm:"default:false"`
	DALLEGenerated  bool `json:"dalleGenerated" gorm:"default:false"`
	TTSGenerated    bool `json:"ttsGenerated" gorm:"default:false"`
	SVTGenerated    bool `json:"svtGenerated" gorm:"default:false"`
	VideoStitched   bool `json:"videoStitched" gorm:"default:false"`

	Progress 	  int  `json:"progress" gorm:"default:0"`

	// full progress of the video
	VideoGenerated  bool `json:"videoGenerated" gorm:"default:false"`
	VideoUploaded   bool `json:"videoUploaded" gorm:"default:false"`

	// maybe, hide this from the user
	Error 		  string `json:"error" gorm:"null"`

	TTSURL           string `json:"ttsURL" gorm:"null"`
	SVTURL           string `json:"svtURL" gorm:"null"`
	StitchedVideoURL string `json:"stitchedVideoURL" gorm:"null"`

	OwnerID string `json:"ownerID"`
	Owner   User   `json:"owner" gorm:"foreignKey:OwnerID;references:ID"`
}
