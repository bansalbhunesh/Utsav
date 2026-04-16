package model

// OTP request/verify payloads used by HTTP handlers.
type OTPRequest struct {
	Phone string `json:"phone" binding:"required"`
}

type OTPVerify struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

type RefreshBody struct {
	RefreshToken string `json:"refresh_token"`
}

