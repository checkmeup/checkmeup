package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	botToken   string
	httpClient *http.Client
}

func NewClient(botToken string) *Client {
	return &Client{
		botToken: botToken,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type sendMessageRequest struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

type apiResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

// SetWebhook registers webhookURL with the Telegram Bot API and sets a
// secret_token so Telegram includes X-Telegram-Bot-Api-Secret-Token on every
// incoming update. Pass an empty secret to skip token verification (dev only).
func (c *Client) SetWebhook(webhookURL, secret string) error {
	if c.botToken == "" {
		return fmt.Errorf("telegram bot token not configured")
	}
	payload := map[string]string{"url": webhookURL}
	if secret != "" {
		payload["secret_token"] = secret
	}
	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/setWebhook", c.botToken)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("setWebhook request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	var result apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	if !result.OK {
		return fmt.Errorf("setWebhook error: %s", result.Description)
	}
	return nil
}

// WebhookUpdate is a minimal subset of the Telegram Update object.
type WebhookUpdate struct {
	UpdateID int              `json:"update_id"`
	Message  *WebhookMessage `json:"message"`
}

type WebhookMessage struct {
	Text string `json:"text"`
	Chat struct {
		ID int64 `json:"id"`
	} `json:"chat"`
}

// HandleUpdate processes an incoming webhook update from Telegram.
// It replies to /start with the chat's numeric ID so users can paste it into Settings.
func (c *Client) HandleUpdate(update WebhookUpdate) {
	if update.Message == nil {
		return
	}
	if !strings.HasPrefix(update.Message.Text, "/start") {
		return
	}
	chatID := fmt.Sprintf("%d", update.Message.Chat.ID)
	text := fmt.Sprintf(
		"Your Chat ID is:\n\n<code>%s</code>\n\nPaste this into <b>checkmeup → Settings → Telegram</b> to receive monitoring alerts.",
		chatID,
	)
	_ = c.SendMessage(chatID, text)
}

func (c *Client) SendMessage(chatID, text string) error {
	if c.botToken == "" {
		return fmt.Errorf("telegram bot token not configured")
	}

	body, err := json.Marshal(sendMessageRequest{ChatID: chatID, Text: text, ParseMode: "HTML"})
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.botToken)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("telegram request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("telegram response decode failed: %w", err)
	}
	if !result.OK {
		return fmt.Errorf("telegram API error: %s", result.Description)
	}
	return nil
}
