package main

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)

func sendEmail(email string, code string) {
    url := "https://encnotes.onrender.com"
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
            a {
                background-color: rgb(77, 77, 238);
                color: white;
                padding: 15px;
                text-decoration: none;
                display: inline-block;
                border-radius: 5px;
                box-shadow: 1px 1px 5px gray;
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
            <a target='_blank' href='`+ url + `/api/verify?email=` + email + `&code=` + code + `'>Click here to Verify</a>
            <p>This verification link will expire in 15 minutes. You do not need to reply to this E-Mail</p>
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
		log.Fatal(err)
	}
	fmt.Println("Sent email")
}
