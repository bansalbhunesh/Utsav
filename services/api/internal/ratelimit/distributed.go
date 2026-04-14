package ratelimit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Limiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

type InMemoryLimiter struct {
	win *Window
}

func NewInMemoryLimiter(max int, window time.Duration) *InMemoryLimiter {
	return &InMemoryLimiter{win: New(max, window)}
}

func (l *InMemoryLimiter) Allow(_ context.Context, key string) (bool, error) {
	return l.win.Allow(key), nil
}

type UpstashRESTLimiter struct {
	restURL string
	token   string
	max     int
	window  time.Duration
	client  *http.Client
	mu      sync.Mutex
}

func NewUpstashRESTLimiter(restURL, token string, max int, window time.Duration) *UpstashRESTLimiter {
	return &UpstashRESTLimiter{
		restURL: strings.TrimRight(strings.TrimSpace(restURL), "/"),
		token:   strings.TrimSpace(token),
		max:     max,
		window:  window,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

type upstashCommand struct {
	Command []string `json:"command"`
}

type upstashResult struct {
	Result any    `json:"result"`
	Error  string `json:"error"`
}

func (l *UpstashRESTLimiter) Allow(ctx context.Context, key string) (bool, error) {
	if l.restURL == "" || l.token == "" {
		return false, fmt.Errorf("upstash not configured")
	}
	windowSec := int(l.window.Seconds())
	if windowSec <= 0 {
		windowSec = 60
	}

	pipeline := []upstashCommand{
		{Command: []string{"INCR", key}},
		{Command: []string{"EXPIRE", key, strconv.Itoa(windowSec), "NX"}},
	}
	body, _ := json.Marshal(pipeline)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, l.restURL+"/pipeline", bytes.NewReader(body))
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+l.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := l.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return false, fmt.Errorf("upstash limiter failed with status %d", resp.StatusCode)
	}

	var out []upstashResult
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return false, err
	}
	if len(out) == 0 {
		return false, fmt.Errorf("upstash limiter empty response")
	}
	if out[0].Error != "" {
		return false, fmt.Errorf("upstash limiter error: %s", out[0].Error)
	}

	var count int
	switch v := out[0].Result.(type) {
	case float64:
		count = int(v)
	case string:
		n, _ := strconv.Atoi(v)
		count = n
	default:
		return false, fmt.Errorf("upstash limiter unexpected response")
	}
	return count <= l.max, nil
}
