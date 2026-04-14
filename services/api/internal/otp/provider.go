package otp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Sender interface {
	SendOTP(ctx context.Context, phone string, code string) error
}

type MSG91Sender struct {
	AuthKey string
	Sender  string
	Route   string
	Client  *http.Client
}

func NewMSG91Sender(authKey, sender, route string) *MSG91Sender {
	if strings.TrimSpace(route) == "" {
		route = "4"
	}
	return &MSG91Sender{
		AuthKey: strings.TrimSpace(authKey),
		Sender:  strings.TrimSpace(sender),
		Route:   strings.TrimSpace(route),
		Client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *MSG91Sender) SendOTP(ctx context.Context, phone string, code string) error {
	if strings.TrimSpace(s.AuthKey) == "" || strings.TrimSpace(s.Sender) == "" {
		return fmt.Errorf("msg91 not configured")
	}
	cleanPhone := strings.TrimSpace(phone)
	if cleanPhone == "" {
		return fmt.Errorf("phone is required")
	}
	msg := fmt.Sprintf("%s is your UTSAV OTP. It is valid for 5 minutes.", code)
	q := url.Values{}
	q.Set("authkey", s.AuthKey)
	q.Set("mobiles", cleanPhone)
	q.Set("message", msg)
	q.Set("sender", s.Sender)
	q.Set("route", s.Route)
	q.Set("country", "91")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.msg91.com/api/sendhttp.php?"+q.Encode(), nil)
	if err != nil {
		return err
	}
	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("msg91 send failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}
