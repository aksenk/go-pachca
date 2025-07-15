package pachca

import (
	"time"
)

type WebhookReaction struct {
	Type      string    `json:"type"`
	Event     string    `json:"event"`
	MessageID int       `json:"message_id"`
	Code      string    `json:"code"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
}
