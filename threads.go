package pachca

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"time"
)

// Threads
// Объект для работы с обсуждениями (тредами)
type Threads struct {
	client *resty.Client
}

// Thread
// Объект описывающий тред
type Thread struct {
	ID            int       `json:"id"`
	ChatID        int       `json:"chat_id"`
	MessageID     int       `json:"message_id"`
	MessageChatID int       `json:"message_chat_id"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ThreadResponseRaw
// Объект для хранения ответов от API с описанием треда
type ThreadResponseRaw struct {
	Data Thread `json:"data"`
}

// New
// Метод для создания нового треда к сообщению.
// Если у сообщения уже был создан тред, то в ответе на запрос вернётся информация об уже созданном раннее треде.
func (t *Threads) New(ctx context.Context, messageID int) (*Thread, *resty.Response, error) {
	if messageID <= 0 {
		return nil, nil, fmt.Errorf("invalid message ID: %d", messageID)
	}

	url := fmt.Sprintf("%v/%v/thread", messagesURL, messageID)
	resp, err := t.client.R().
		SetContext(ctx).
		Post(url)
	if err != nil {
		return nil, resp, err
	}

	if resp.StatusCode() != 201 {
		return nil, resp, fmt.Errorf("%w: %d", ErrResponseCode, resp.StatusCode())
	}

	var r ThreadResponseRaw
	err = json.Unmarshal(resp.Body(), &r)
	if err != nil {
		return nil, resp, fmt.Errorf("%w: %w", ErrResponseDecode, err)
	}

	return &r.Data, resp, nil
}

// Get
// Метод для получения информации о треде.
func (t *Threads) Get(ctx context.Context, threadID int) (*Thread, *resty.Response, error) {
	if threadID <= 0 {
		return nil, nil, fmt.Errorf("invalid thread ID: %d", threadID)
	}

	url := fmt.Sprintf("%v/%v", threadsURL, threadID)
	resp, err := t.client.R().
		SetContext(ctx).
		Get(url)
	if err != nil {
		return nil, resp, err
	}

	if resp.StatusCode() != 200 {
		return nil, resp, fmt.Errorf("%w: %d", ErrResponseCode, resp.StatusCode())
	}

	var r ThreadResponseRaw
	err = json.Unmarshal(resp.Body(), &r)
	if err != nil {
		return nil, resp, fmt.Errorf("%w: %w", ErrResponseDecode, err)
	}

	return &r.Data, resp, nil
}
