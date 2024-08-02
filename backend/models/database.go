package models

import (
	"github.com/dgrijalva/jwt-go"
)

// later, break this into two models: Video and Schedule
type Video struct {
	Base
	Topic           string `json:"topic"`
	Description     string `json:"description"`
	Narrator        string `json:"narrator"`
	VideoStyle      string `json:"videoStyle"`
	PostingSchedule string `json:"postingSchedule"`
	IsOneTime       bool   `json:"isOneTime"`
	VideoURL 	  string `json:"videoURL" gorm:"null"`
}