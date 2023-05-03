package main

import (
	"os"

	"github.com/kataras/hcaptcha"
)

func verifyCaptcha(token string) bool {
	secret := os.Getenv("HCAPTCHA_SECRET")
	client := hcaptcha.New(secret)
	resp := client.VerifyToken(token)

	return resp.Success
}
