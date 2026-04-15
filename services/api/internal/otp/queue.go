package otp

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/hibiken/asynq"
)

const TaskTypeSendOTP = "otp:send"

type Dispatcher interface {
	DispatchOTP(ctx context.Context, phone, code string) error
}

type DirectDispatcher struct {
	Sender Sender
}

func (d *DirectDispatcher) DispatchOTP(ctx context.Context, phone, code string) error {
	if d == nil || d.Sender == nil {
		return nil
	}
	return d.Sender.SendOTP(ctx, phone, code)
}

type QueueDispatcher struct {
	client *asynq.Client
}

type otpTaskPayload struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

func NewQueueDispatcher(client *asynq.Client) *QueueDispatcher {
	return &QueueDispatcher{client: client}
}

func (d *QueueDispatcher) DispatchOTP(_ context.Context, phone, code string) error {
	if d == nil || d.client == nil {
		return fmt.Errorf("otp queue client is not configured")
	}
	payload, _ := json.Marshal(otpTaskPayload{Phone: strings.TrimSpace(phone), Code: strings.TrimSpace(code)})
	task := asynq.NewTask(TaskTypeSendOTP, payload)
	_, err := d.client.Enqueue(task, asynq.MaxRetry(5), asynq.Queue("critical"))
	return err
}

func NewOTPTaskHandler(sender Sender) asynq.Handler {
	return asynq.HandlerFunc(func(_ context.Context, task *asynq.Task) error {
		if sender == nil {
			return fmt.Errorf("otp sender is not configured")
		}
		var payload otpTaskPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("decode otp task payload: %w", err)
		}
		return sender.SendOTP(context.Background(), payload.Phone, payload.Code)
	})
}

func RedisClientOptFromURL(raw string) (asynq.RedisClientOpt, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return asynq.RedisClientOpt{}, err
	}
	pass, _ := u.User.Password()
	db := 0
	if q := strings.TrimSpace(u.Query().Get("db")); q != "" {
		if n, convErr := strconv.Atoi(q); convErr == nil {
			db = n
		}
	}
	var tlsCfg *tls.Config
	if strings.EqualFold(u.Scheme, "rediss") {
		tlsCfg = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	return asynq.RedisClientOpt{
		Addr:      u.Host,
		Password:  pass,
		DB:        db,
		TLSConfig: tlsCfg,
	}, nil
}
