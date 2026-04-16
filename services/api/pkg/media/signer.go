package media

import (
	"net/url"
	"strings"
	"time"
)

type PresignRequest struct {
	ObjectKey   string
	ContentType string
	ExpiresIn   time.Duration
}

type PresignResponse struct {
	Method     string            `json:"method"`
	URL        string            `json:"url"`
	Headers    map[string]string `json:"headers"`
	ObjectKey  string            `json:"object_key"`
	ExpiresInS int64             `json:"expires_in_seconds"`
}

type Signer interface {
	PresignPut(req PresignRequest) (PresignResponse, error)
	PublicObjectURL(objectKey string) string
}

type URLSigner struct {
	BaseURL string
}

func (s URLSigner) PresignPut(req PresignRequest) (PresignResponse, error) {
	expires := req.ExpiresIn
	if expires <= 0 {
		expires = 10 * time.Minute
	}
	headers := map[string]string{}
	if strings.TrimSpace(req.ContentType) != "" {
		headers["Content-Type"] = req.ContentType
	}
	return PresignResponse{
		Method:     "PUT",
		URL:        s.PublicObjectURL(req.ObjectKey),
		Headers:    headers,
		ObjectKey:  req.ObjectKey,
		ExpiresInS: int64(expires.Seconds()),
	}, nil
}

func (s URLSigner) PublicObjectURL(objectKey string) string {
	base := strings.TrimRight(strings.TrimSpace(s.BaseURL), "/")
	key := strings.TrimLeft(strings.TrimSpace(objectKey), "/")
	if base == "" {
		return "/" + url.PathEscape(key)
	}
	parts := strings.Split(key, "/")
	escaped := make([]string, 0, len(parts))
	for _, p := range parts {
		escaped = append(escaped, url.PathEscape(p))
	}
	return base + "/" + strings.Join(escaped, "/")
}
