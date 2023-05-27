Notes api written in go.

[![Deploy on Railway](https://railway.app/button.svg)](https://railway.app/template/5ZAVcm?referralCode=Z03iPU)

API uses MongoDB to store user Data.

# API Endpoints

## GET
- /api/ping
- /api/notes
- /api/notes/[id]
- /api/info
- /api/verify

## POST
- /api/register
- /api/login
- /api/logout
- /api/notes
- /api/sendEmail


## Delete
- /api/notes

## PUT
- /api/notes

# Environment Variables Required

`MONGO_URL`

`PORT`

`EMAIL_PASSWORD`

`EMAIL`

`SMTP_HOST`

`HCAPTCHA_SECRET`