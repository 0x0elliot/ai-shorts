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
	TTSGenerated    bool `json:"ttsGenerated" gorm:"default:false"`
	VideoGenerated  bool `json:"videoGenerated" gorm:"default:false"`
	VideoStitched   bool `json:"videoStitched" gorm:"default:false"`

	TTSURL           string `json:"ttsURL" gorm:"null"`
	StitchedVideoURL string `json:"stitchedVideoURL" gorm:"null"`

	OwnerID string `json:"ownerID"`
	Owner   User   `json:"owner" gorm:"foreignKey:OwnerID;references:ID"`
}
