package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookService handles webhook delivery.
type WebhookService struct {
	client *http.Client
}

// NewWebhookService creates a new WebhookService.
func NewWebhookService() *WebhookService {
	return &WebhookService{
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

// WebhookPayload is the data sent to webhook endpoints.
type WebhookPayload struct {
	Event     string      `json:"event"`
	Timestamp string      `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// Deliver sends a webhook payload to the given URL with HMAC-SHA256 signature.
func (s *WebhookService) Deliver(url string, secret string, event string, data interface{}) error {
	payload := WebhookPayload{
		Event:     event,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data:      data,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	// Sign with HMAC-SHA256.
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tavern-Signature", sig)
	req.Header.Set("X-Tavern-Event", event)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("deliver webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned %d", resp.StatusCode)
	}
	return nil
}
