package pachca

import (
	"encoding/json"
	"time"
)

// WebhookType
// Объект для определения типа входящего вебхука
type WebhookType struct {
	Type string `json:"type"`
}

// WebhookMessage
// Объект вебхука о событии сообщения (создание/обновление/удаление)
type WebhookMessage struct {
	Type            string     `json:"type"`
	Event           string     `json:"event"`
	ChatID          int        `json:"chat_id"`
	Content         string     `json:"content"`
	UserID          int        `json:"user_id"`
	ID              int        `json:"id"`
	CreatedAt       time.Time  `json:"created_at"`
	ChangedAt       *time.Time `json:"changed_at"`
	ParentMessageID *int       `json:"parent_message_id"`
	EntityType      string     `json:"entity_type"`
	EntityID        int        `json:"entity_id"`
	URL             string     `json:"url"`
	Thread          *struct {
		MessageID     int `json:"message_id"`
		MessageChatID int `json:"message_chat_id"`
	} `json:"thread"`
	WebhookTimestamp int `json:"webhook_timestamp"`
}

// WebhookButton
// Объект вебхука о нажатии на кнопку в сообщении
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

// WebhookView
// Объект вебхука об отправке формы (модального окна)
type WebhookView struct {
	Type             string          `json:"type"`
	Event            string          `json:"event"`
	PrivateMetadata  string          `json:"private_metadata"`
	CallbackID       string          `json:"callback_id"`
	ChatID           *int            `json:"chat_id"`
	UserID           int             `json:"user_id"`
	ViewID           string          `json:"view_id"`
	SubmitID         string          `json:"submit_id"`
	Data             json.RawMessage `json:"data"`
	WebhookTimestamp int             `json:"webhook_timestamp"`
}

// WebhookChatMember
// Объект вебхука о добавлении/удалении участников чата или треда
type WebhookChatMember struct {
	Type             string    `json:"type"`
	Event            string    `json:"event"`
	ChatID           int       `json:"chat_id"`
	ThreadID         *int      `json:"thread_id"`
	UserIDs          []int     `json:"user_ids"`
	CreatedAt        time.Time `json:"created_at"`
	WebhookTimestamp int       `json:"webhook_timestamp"`
}

// WebhookCompanyMember
// Объект вебхука о событии участника компании (приглашение/подтверждение/обновление/блокировка/разблокировка/удаление)
type WebhookCompanyMember struct {
	Type             string    `json:"type"`
	Event            string    `json:"event"`
	UserIDs          []int     `json:"user_ids"`
	CreatedAt        time.Time `json:"created_at"`
	WebhookTimestamp int       `json:"webhook_timestamp"`
}

// WebhookLink
// Ссылка, упомянутая в сообщении, в событии WebhookLinkShared
type WebhookLink struct {
	URL    string `json:"url"`
	Domain string `json:"domain"`
	Skip   bool   `json:"skip"`
}

// WebhookLinkShared
// Объект вебхука об упоминании ссылки в сообщении
type WebhookLinkShared struct {
	Type             string        `json:"type"`
	Event            string        `json:"event"`
	ChatID           int           `json:"chat_id"`
	MessageID        int           `json:"message_id"`
	Links            []WebhookLink `json:"links"`
	UserID           int           `json:"user_id"`
	CreatedAt        time.Time     `json:"created_at"`
	WebhookTimestamp int           `json:"webhook_timestamp"`
}

// WebhookVideoCallMember
// Участник видеозвонка в событии WebhookVideoCall
type WebhookVideoCallMember struct {
	UserID   int        `json:"user_id"`
	JoinedAt time.Time  `json:"joined_at"`
	LeftAt   *time.Time `json:"left_at"`
}

// WebhookVideoCall
// Объект вебхука о событии видеозвонка (начало/завершение/готовность записи)
type WebhookVideoCall struct {
	Type        string `json:"type"`
	Event       string `json:"event"`
	VideoRoomID int    `json:"video_room_id"`
	ChatID      int    `json:"chat_id"`
	OwnerID     int    `json:"owner_id"`
	Thread      *struct {
		ID            int `json:"id"`
		ChatID        int `json:"chat_id"`
		MessageID     int `json:"message_id"`
		MessageChatID int `json:"message_chat_id"`
	} `json:"thread"`
	StartedAt        time.Time                `json:"started_at"`
	FinishedAt       time.Time                `json:"finished_at"`
	Duration         int                      `json:"duration"`
	Members          []WebhookVideoCallMember `json:"members"`
	RecordingID      int                      `json:"recording_id"`
	FileID           *int                     `json:"file_id"`
	URL              string                   `json:"url"`
	Size             int64                    `json:"size"`
	WebhookTimestamp int                      `json:"webhook_timestamp"`
}
