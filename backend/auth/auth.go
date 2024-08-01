package auth

import (
	"log"
	"time"

	db "go-authentication-boilerplate/database"
	"go-authentication-boilerplate/models"

	uuidLib "github.com/google/uuid"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"

	email "go-authentication-boilerplate/email"
)

var jwtKey = []byte(db.PRIVKEY)

// GenerateTokens generates the access and refresh tokens
func GenerateTokens(uuid string) (string, string, error) {
	claim, accessToken, err := GenerateAccessClaims(uuid)
	if err != nil {
		log.Printf("[ERROR] Couldn't generate access token: %v", err)
		return "", "", err
	}

	refreshToken, err := GenerateRefreshClaims(claim)
	if err != nil {
		log.Printf("[ERROR] Couldn't generate refresh token: %v", err)
		return "", "", err	
	}

	return accessToken, refreshToken, nil
}

func GeneratePasswordLessLink(user *models.User) error {
	claim, accessToken, err := GenerateAccessClaims(user.ID)
	if err != nil {
		log.Printf("[ERROR] Couldn't generate access token: %v", err)
		return err
	}

	refreshToken, err := GenerateRefreshClaims(claim)
	if err != nil {
		log.Printf("[ERROR] Couldn't generate refresh token: %v", err)
		return err
	}
	
	url := "http://localhost:3000/magic-link-auth?accessToken=" + accessToken + "&refreshToken=" + refreshToken

	log.Printf("[INFO] Sending email to %v", url)

	// (from string, to []string, subject string, body string, html string, cc []string, bcc []string, replyto string, service string) error {
	err = email.SendEmail(
		"wolfwithahat@protonmail.com",
		[]string{user.Email},
		"Passwordless Login Link!",
		"Please click on the link to login: "+url,
		"Please click on the link to login: <a href='"+url+"'>Login</a>",
		[]string{},
		[]string{},
		"",
		"resend",
	)

	if err != nil {
		log.Printf("[ERROR] Couldn't send email (1): %v", err)
		return err
	}

	return nil
}

// GenerateAccessClaims returns a claim and a acess_token string
func GenerateAccessClaims(uuid string) (*models.Claims, string, error) {
	t := time.Now()
	claim := &models.Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    uuid,
			ExpiresAt: t.Add(24 * time.Hour).Unix(),
			Subject:   "access_token",
			IssuedAt:  t.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		log.Printf("[ERROR] Couldn't sign the token: %s", err)
		return claim, "", err
	}

	return claim, tokenString, nil
}

// GenerateRefreshClaims returns refresh_token
func GenerateRefreshClaims(cl *models.Claims) (string, error) {
	result := db.DB.Where(&models.Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer: cl.Issuer,
		},
	}).Find(&models.Claims{})

	// checking the number of refresh tokens stored.
	// If the number is higher than 3, remove all the refresh tokens and leave only new one.
	if result.RowsAffected > 3 {
		db.DB.Where(&models.Claims{StandardClaims: jwt.StandardClaims{Issuer: cl.Issuer}}).Delete(&models.Claims{})
	}

	t := time.Now()
	refreshClaim := &models.Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    cl.Issuer,
			ExpiresAt: t.Add(10 * 24 * time.Hour).Unix(),
			Subject:   "refresh_token",
			IssuedAt:  t.Unix(),
		},
	}

	refreshClaim.Id = uuidLib.New().String()

	// create a claim on DB
	db.DB.Create(&refreshClaim)

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaim)
	refreshTokenString, err := refreshToken.SignedString(jwtKey)
	if err != nil {
		log.Printf("[ERROR] Couldn't sign the token: %s", err)
		return "", err
	}

	return refreshTokenString, nil
}

// SecureAuth returns a middleware which secures all the private routes
func SecureAuth() func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		accessToken := c.Get("Authorization")

		if len(accessToken) > 7 {
			if accessToken[:7] == "Bearer " {
				accessToken = accessToken[7:]
			}
		} else {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "Token not found",
			})
		}

		if len(accessToken) == 0 {
			accessToken = c.Cookies("access_token")
		}

		claims := new(models.Claims)
		token, err := jwt.ParseWithClaims(accessToken, claims,
			func(token *jwt.Token) (interface{}, error) {
				return jwtKey, nil
		})

		if err != nil {
			log.Printf("[ERROR] Couldn't parse the token: %s", accessToken)
			log.Printf("[ERROR] Couldn't parse the token: %s", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "Couldn't parse the token",
			})
		}

		if token.Valid {
			if claims.ExpiresAt < time.Now().Unix() {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   true,
					"message": "Token Expired",
				})
			}
		} else if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				// this is not even a token, we should delete the cookies here
				c.ClearCookie("access_token", "refresh_token")
				return c.SendStatus(fiber.StatusForbidden)
			} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
				// Token is either expired or not active yet
				return c.SendStatus(fiber.StatusUnauthorized)
			} else {
				// cannot handle this token
				c.ClearCookie("access_token", "refresh_token")
				return c.SendStatus(fiber.StatusForbidden)
			}
		}

		c.Locals("id", claims.Issuer)
		return c.Next()
	}
}

// GetAuthCookies sends two cookies of type access_token and refresh_token
func GetAuthCookies(accessToken, refreshToken string) (*fiber.Cookie, *fiber.Cookie) {
	accessCookie := &fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		Secure:   true,
	}

	refreshCookie := &fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  time.Now().Add(10 * 24 * time.Hour),
		HTTPOnly: true,
		Secure:   true,
	}

	return accessCookie, refreshCookie
}
