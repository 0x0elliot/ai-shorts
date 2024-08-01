package util

import (
	"github.com/resend/resend-go/v2"
	"os"
	"log"

	"errors"
)

func SendEmail(from string, to []string, subject string, body string, html string, cc []string, bcc []string, replyto string, service string) error {
	if service == "" {
		service = "resend"
	}

	if service == "resend" {
		apiKey := os.Getenv("ACIDRAIN_RESEND_API_KEY")
		fromEmail := os.Getenv("ACIDRAIN_RESEND_EMAIL")
		if apiKey == "" || fromEmail == "" {
			log.Printf("[ERROR] Couldn't send email: %s", "API Key or Email not set")
			return errors.New("API Key or Email not set")
		}

		if len(html) != 0{
			body = html
		}

		r := resend.NewClient(apiKey)
		params := &resend.SendEmailRequest{
			From:    fromEmail,
			To:      to,
			Html:    html,
			Subject: subject,
			Cc: 	 cc,
			Bcc:     bcc,
			ReplyTo: replyto,
		}

		sent, err := r.Emails.Send(params)
		if err != nil {
			log.Printf("[ERROR] Couldn't send email: %v", err)
			return err
		}

		log.Printf("[INFO] Email sent: %v", sent)
		return nil
	}

	log.Printf("[ERROR] Couldn't send email: %s", "Invalid service")
	return nil
}


