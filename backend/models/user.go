package models

import (
	"github.com/dgrijalva/jwt-go"
)

// User represents a User schema
type User struct {
	Base
	Email	string `json:"email" gorm:"unique;not null"`
}

// UserErrors represent the error format for user routes
type UserErrors struct {
	Err      bool   `json:"error"`
	Email    string `json:"email"`
}

// Claims represent the structure of the JWT token
type Claims struct {
	jwt.StandardClaims
}
