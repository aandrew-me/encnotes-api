package main

import (
	"fmt"
	"net/smtp"
	"os"
)

func sendEmail(email string, code string) {
	text := `<!DOCTYPE html>
    <html lang='en'>
    <head>
        <meta charset='UTF-8'>
        <meta http-equiv='X-UA-Compatible' content='IE=edge'>
        <meta name='viewport' content='width=device-width, initial-scale=1.0'>
        <title>Email Verification</title>
        <style>
            body {
                text-align: center;
                font-family: system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
            }
            #box {
                background-color: aliceblue;
                padding: 35px;
            }
        </style>
    </head>
    <body>
        <div id='box'>
            <p>Hello, To complete your registration, you need to verify your Email</p>
            <p>Your verification code is <b>` + code + `</b></p>
            <p>This verification code will expire in 15 minutes. You do not need to reply to this E-Mail</p>
        </div>
        <p>If you didn't intent to receive this email, just ignore it.</p>
    </body></html>
    `

	from := os.Getenv("EMAIL")
	password := os.Getenv("EMAIL_PASSWORD")
	to := []string{email}
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := "587"

	msg := "Subject: Email Verification for EncNotes\r\nMIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n" + text

	messageByte := []byte(msg)

	// Create authentication
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Send actual message
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, messageByte)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Sent email")
	}
}
