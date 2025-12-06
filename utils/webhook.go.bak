package utils

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

// WebhookPayload represents the payload sent to webhook
type WebhookPayload struct {
	Action    string `json:"action"` // "create", "update", "delete"
	ArticleID string `json:"article_id"`
	Timestamp string `json:"timestamp"`
}

// SendBuildWebhook sends a webhook notification to trigger a build
func SendBuildWebhook(action, articleID string) {
	webhookURL := os.Getenv("WEBHOOK_URL")
	if webhookURL == "" {
		log.Println("WEBHOOK_URL not configured, skipping webhook notification")
		return
	}

	payload := WebhookPayload{
		Action:    action,
		ArticleID: articleID,
		Timestamp: "",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal webhook payload: %v", err)
		return
	}

	go func() {
		resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("Failed to send webhook: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			log.Printf("Webhook sent successfully for action: %s, article: %s", action, articleID)
		} else {
			log.Printf("Webhook failed with status: %d", resp.StatusCode)
		}
	}()
}
