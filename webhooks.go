package pachca

import "time"

type WebhookType struct {
	Type string `json:"type"`
}

type WebhookMessage struct {
	Type            string    `json:"type"`
	Event           string    `json:"event"`
	ChatID          int       `json:"chat_id"`
	Content         string    `json:"content"`
	UserID          int       `json:"user_id"`
	ID              int       `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	ParentMessageID *int      `json:"parent_message_id"`
	EntityType      string    `json:"entity_type"`
	EntityID        int       `json:"entity_id"`
	URL             string    `json:"url"`
	Thread          *struct {
		MessageID     int `json:"message_id"`
		MessageChatID int `json:"message_chat_id"`
	} `json:"thread"`
	WebhookTimestamp int `json:"webhook_timestamp"`
}

type WebhookButton struct {
	Type             string `json:"type"`
	Event            string `json:"event"`
	MessageID        int    `json:"message_id"`
	TriggerID        string `json:"trigger_id"`
	Data             string `json:"data"`
	UserID           int    `json:"user_id"`
	ChatID           int    `json:"chat_id"`
	WebhookTimestamp int    `json:"webhook_timestamp"`
}
