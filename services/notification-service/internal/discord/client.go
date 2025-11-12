package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"toxictoast/services/notification-service/internal/domain"

	"github.com/toxictoast/toxictoastgo/shared/logger"
)

// Client handles Discord webhook deliveries with embeds
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new Discord client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Embed represents a Discord embed message
type Embed struct {
	ThreadTitle string       `json:"thread_title,omitempty"`
	Title       string       `json:"title,omitempty"`
	Description string       `json:"description,omitempty"`
	URL         string       `json:"url,omitempty"`
	Color       int          `json:"color,omitempty"`
	Fields      []EmbedField `json:"fields,omitempty"`
	Footer      *EmbedFooter `json:"footer,omitempty"`
	Timestamp   string       `json:"timestamp,omitempty"` // ISO 8601 timestamp
	Thumbnail   *EmbedImage  `json:"thumbnail,omitempty"`
	Image       *EmbedImage  `json:"image,omitempty"`
	Author      *EmbedAuthor `json:"author,omitempty"`
}

// EmbedField represents a field in an embed
type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// EmbedFooter represents the footer of an embed
type EmbedFooter struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url,omitempty"`
}

// EmbedImage represents an image in an embed
type EmbedImage struct {
	URL string `json:"url"`
}

// EmbedAuthor represents the author of an embed
type EmbedAuthor struct {
	Name    string `json:"name,omitempty"`
	URL     string `json:"url,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

// WebhookPayload represents the payload sent to Discord
type WebhookPayload struct {
	Content   string  `json:"content,omitempty"`
	Username  string  `json:"username,omitempty"`
	AvatarURL string  `json:"avatar_url,omitempty"`
	Embeds    []Embed `json:"embeds,omitempty"`
}

// DiscordResponse represents Discord's response
type DiscordResponse struct {
	ID        string `json:"id"`
	Type      int    `json:"type"`
	ChannelID string `json:"channel_id"`
}

// SendNotification sends a notification to Discord with an embed
func (c *Client) SendNotification(ctx context.Context, webhookURL string, embed Embed) (string, int, string, error) {
	payload := WebhookPayload{
		Embeds: []Embed{embed},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", 0, "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", 0, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ToxicToastGo-Notification/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", 0, "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024))
	if err != nil {
		logger.Info(fmt.Sprintf("Failed to read response body: %v", err))
		bodyBytes = []byte{}
	}
	responseBody := string(bodyBytes)

	// Discord returns 204 No Content on success (or 200 with message ID)
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
		var discordResp DiscordResponse
		if len(bodyBytes) > 0 {
			if err := json.Unmarshal(bodyBytes, &discordResp); err == nil {
				return discordResp.ID, resp.StatusCode, responseBody, nil
			}
		}
		return "", resp.StatusCode, responseBody, nil
	}

	return "", resp.StatusCode, responseBody, fmt.Errorf("Discord API error: HTTP %d - %s", resp.StatusCode, responseBody)
}

// BuildEmbedFromEvent builds a Discord embed from an event
func BuildEmbedFromEvent(event *domain.Event, color int) Embed {
	embed := Embed{
		ThreadTitle: fmt.Sprintf("%s Event", event.Type),
		Title:       fmt.Sprintf("📢 %s", event.Type),
		Description: fmt.Sprintf("Event from %s", event.Source),
		Color:       color,
		Timestamp:   event.Timestamp.Format(time.RFC3339),
		Footer: &EmbedFooter{
			Text: fmt.Sprintf("Event ID: %s", event.ID),
		},
	}

	// Add event data as fields
	fields := make([]EmbedField, 0)
	for key, value := range event.Data {
		// Convert value to string
		valueStr := fmt.Sprintf("%v", value)
		if len(valueStr) > 1024 {
			valueStr = valueStr[:1021] + "..."
		}

		fields = append(fields, EmbedField{
			Name:   key,
			Value:  valueStr,
			Inline: true,
		})
	}

	embed.Fields = fields
	return embed
}

// GetEmbedColor returns color based on event type
func GetEmbedColor(eventType string) int {
	// Color codes (decimal):
	// 3066993 = Green (success)
	// 15158332 = Red (error)
	// 16776960 = Yellow (warning)
	// 3447003 = Blue (info)
	// 10181046 = Purple (stream)

	if len(eventType) > 4 {
		prefix := eventType[:4]
		switch prefix {
		case "blog":
			return 3447003 // Blue
		case "twitch":
			return 10181046 // Purple
		case "link":
			return 3066993 // Green
		}
	}

	return 5814783 // Default gray
}
