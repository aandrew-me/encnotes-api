# Encnotes API
Notes api written in Go for [Encnotes](https://github.com/aandrew-me/encnotes-client)

[![Deploy on Railway](https://railway.app/button.svg)](https://railway.app/template/5ZAVcm?referralCode=Z03iPU)

# API Endpoints

### GET
- /api/ping
- /api/notes
- /api/notes/[id]
- /api/info
- /api/verify

### POST
- /api/register
- /api/login
- /api/logout
- /api/notes
- /api/sendEmail


### Delete
- /api/notes

### PUT
- /api/notes

## Environment Variables Required

`MONGO_URL`

`PORT`

`EMAIL_PASSWORD`

`EMAIL`

`SMTP_HOST`

`HCAPTCHA_SECRET`

## Used Technologies
- Gofiber for server
- MongoDB for database