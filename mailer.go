package main

import (
	"fmt"
	"net/smtp"
	"os"
)

func sendMail(to, subject, body string) error {
	from := os.Getenv("GMAIL_ADDRESS")
	password := os.Getenv("GMAIL_APP_PASSWORD")

	if from == "" || password == "" {
		return fmt.Errorf("GMAIL_ADDRESS or GMAIL_APP_PASSWORD is not set")
	}

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	auth := smtp.PlainAuth(
		"",
		from,
		password,
		smtpHost,
	)

	message := []byte(
		"From: " + from + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=UTF-8\r\n" +
			"\r\n" +
			body,
	)

	return smtp.SendMail(
		smtpHost+":"+smtpPort,
		auth,
		from,
		[]string{to},
		message,
	)
}
