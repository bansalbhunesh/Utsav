package otp

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type ResilientSender struct {
	Inner            Sender
	MaxRetries       int
	InitialBackoff   time.Duration
	Cooldown         time.Duration
	FailureThreshold int

	mu              sync.Mutex
	consecutiveFail int
	openUntil       time.Time
	probing         int32 // atomic: allows a single half-open probe goroutine
}

func NewResilientSender(inner Sender) *ResilientSender {
	return &ResilientSender{
		Inner:            inner,
		MaxRetries:       3,
		InitialBackoff:   300 * time.Millisecond,
		Cooldown:         30 * time.Second,
		FailureThreshold: 5,
	}
}

func (s *ResilientSender) SendOTP(ctx context.Context, phone, code string) error {
	if s == nil || s.Inner == nil {
		return fmt.Errorf("otp sender unavailable")
	}
	if err := s.guardCircuit(); err != nil {
		return err
	}

	var lastErr error
	backoff := s.InitialBackoff
	if backoff <= 0 {
		backoff = 300 * time.Millisecond
	}
	retries := s.MaxRetries
	if retries < 1 {
		retries = 1
	}
	for attempt := 1; attempt <= retries; attempt++ {
		err := s.Inner.SendOTP(ctx, strings.TrimSpace(phone), strings.TrimSpace(code))
		if err == nil {
			s.markSuccess()
			return nil
		}
		lastErr = err
		if attempt == retries {
			break
		}
		select {
		case <-ctx.Done():
			s.markFailure()
			return ctx.Err()
		case <-time.After(backoff):
			backoff = backoff * 2
		}
	}
	s.markFailure()
	return fmt.Errorf("otp send failed after retries: %w", lastErr)
}

func (s *ResilientSender) guardCircuit() error {
	now := time.Now()
	s.mu.Lock()
	isOpen := !s.openUntil.IsZero() && now.Before(s.openUntil)
	cooldownExpired := !s.openUntil.IsZero() && !now.Before(s.openUntil)
	if isOpen {
		s.mu.Unlock()
		return fmt.Errorf("otp provider circuit breaker open")
	}
	if cooldownExpired {
		// Only one goroutine is allowed to probe after cooldown expires.
		if !atomic.CompareAndSwapInt32(&s.probing, 0, 1) {
			s.mu.Unlock()
			return fmt.Errorf("otp provider circuit breaker probing")
		}
		s.openUntil = time.Time{}
	}
	s.mu.Unlock()
	return nil
}

func (s *ResilientSender) markSuccess() {
	atomic.StoreInt32(&s.probing, 0)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.consecutiveFail = 0
	s.openUntil = time.Time{}
}

func (s *ResilientSender) markFailure() {
	atomic.StoreInt32(&s.probing, 0)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.consecutiveFail++
	threshold := s.FailureThreshold
	if threshold <= 0 {
		threshold = 5
	}
	if s.consecutiveFail >= threshold {
		cd := s.Cooldown
		if cd <= 0 {
			cd = 30 * time.Second
		}
		s.openUntil = time.Now().Add(cd)
		s.consecutiveFail = 0
	}
}
