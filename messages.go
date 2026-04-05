package pachca

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

// ---------- Пагинация ----------

type CursorPagination struct {
	Limit  int    `json:"limit,omitempty"`  // количество записей (1–50, по умолчанию 50)
	Cursor string `json:"cursor,omitempty"` // курсор следующей страницы из meta.paginate.next_page
}

// ChatMessagesOptions параметры для получения сообщений чата
type ChatMessagesOptions struct {
	ChatID int
	CursorPagination
	SortField string // "id" или "created_at" (опционально)
	SortOrder string // "asc" или "desc" (по умолчанию "desc")
}

// ---------- Модели данных ----------

// Message основная структура для отправки/редактирования
type Message struct {
	EntityType         string            `json:"entity_type,omitempty"`          // "discussion", "thread", "user"
	EntityID           int               `json:"entity_id"`                      // ID чата/треда/пользователя
	Content            string            `json:"content"`                        // текст сообщения
	Files              []MessageFile     `json:"files,omitempty"`                // вложения
	Buttons            [][]MessageButton `json:"buttons,omitempty"`              // кнопки
	ParentMessageID    *int              `json:"parent_message_id,omitempty"`    // ответ на сообщение
	DisplayAvatarURL   string            `json:"display_avatar_url,omitempty"`   // кастомная аватарка (только для ботов)
	DisplayName        string            `json:"display_name,omitempty"`         // кастомное имя (только для ботов)
	SkipInviteMentions bool              `json:"skip_invite_mentions,omitempty"` // не добавлять упомянутых в тред
}

// OutgoingMessage обёртка для POST/PUT запросов
type OutgoingMessage struct {
	Message     *Message `json:"message"`
	LinkPreview bool     `json:"link_preview,omitempty"` // показывать превью ссылки
}

// MessageFile вложение
type MessageFile struct {
	Key      string `json:"key"`              // путь от /uploads
	Name     string `json:"name"`             // имя файла с расширением
	FileType string `json:"file_type"`        // "file" или "image"
	Size     int    `json:"size"`             // размер в байтах
	Width    int    `json:"width,omitempty"`  // для изображений
	Height   int    `json:"height,omitempty"` // для изображений
}

// MessageButton кнопка
type MessageButton struct {
	Text string `json:"text"`
	Data string `json:"data,omitempty"` // данные для вебхука
	URL  string `json:"url,omitempty"`  // ссылка
}

// MessageResponse структура ответа API
type MessageResponse struct {
	ID               int        `json:"id"`
	EntityType       string     `json:"entity_type"`
	EntityID         int        `json:"entity_id"`
	ChatID           int        `json:"chat_id"`
	RootChatID       int        `json:"root_chat_id"`
	Content          string     `json:"content"`
	UserID           int        `json:"user_id"`
	CreatedAt        time.Time  `json:"created_at"`
	ChangedAt        *time.Time `json:"changed_at,omitempty"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
	URL              string     `json:"url"`
	DisplayAvatarURL *string    `json:"display_avatar_url,omitempty"`
	DisplayName      *string    `json:"display_name,omitempty"`
	Files            []struct {
		ID       int    `json:"id"`
		Key      string `json:"key"`
		Name     string `json:"name"`
		FileType string `json:"file_type"`
		URL      string `json:"url"`
		Width    int    `json:"width,omitempty"`
		Height   int    `json:"height,omitempty"`
	} `json:"files"`
	Buttons [][]MessageButton `json:"buttons,omitempty"`
	Thread  *struct {
		ID     int `json:"id"`
		ChatID int `json:"chat_id"`
	} `json:"thread,omitempty"`
	Forwarding *struct {
		OriginalMessageID          int       `json:"original_message_id"`
		OriginalChatID             int       `json:"original_chat_id"`
		AuthorID                   int       `json:"author_id"`
		OriginalCreatedAt          time.Time `json:"original_created_at"`
		OriginalThreadID           *int      `json:"original_thread_id,omitempty"`
		OriginalThreadMessageID    *int      `json:"original_thread_message_id,omitempty"`
		OriginalThreadParentChatID *int      `json:"original_thread_parent_chat_id,omitempty"`
	} `json:"forwarding,omitempty"`
	ParentMessageID *int `json:"parent_message_id,omitempty"`
}

// messageResponseRaw для одиночного ответа
type messageResponseRaw struct {
	Data MessageResponse `json:"data"`
}

// messagesResponseRaw для списка
type messagesResponseRaw struct {
	Data []MessageResponse `json:"data"`
	Meta struct {
		Paginate struct {
			NextPage string `json:"next_page"`
		} `json:"paginate"`
	} `json:"meta"`
}

// ---------- Клиент для сообщений ----------

type Messages struct {
	client *resty.Client
}

// NewMessages создаёт новый объект для работы с сообщениями
func NewMessages(client *resty.Client) *Messages {
	return &Messages{client: client}
}

// Get возвращает одно сообщение по ID
func (m *Messages) Get(ctx context.Context, messageID int) (*MessageResponse, *resty.Response, error) {
	if messageID <= 0 {
		return nil, nil, fmt.Errorf("%w: incorrect message ID %d", ErrInvalidInput, messageID)
	}

	url := fmt.Sprintf("%s/%d", messagesURL, messageID)
	resp, err := m.client.R().
		SetContext(ctx).
		Get(url)
	if err != nil {
		return nil, resp, err
	}
	if resp.StatusCode() != 200 {
		return nil, resp, fmt.Errorf("%w: %d", ErrResponseCode, resp.StatusCode())
	}

	var raw messageResponseRaw
	if err := json.Unmarshal(resp.Body(), &raw); err != nil {
		return nil, resp, fmt.Errorf("%w: %v", ErrResponseDecode, err)
	}
	return &raw.Data, resp, nil
}

// ListChatMessages возвращает список сообщений чата с пагинацией по курсору
func (m *Messages) ListChatMessages(ctx context.Context, opts *ChatMessagesOptions) ([]MessageResponse, string, *resty.Response, error) {
	if opts == nil {
		return nil, "", nil, fmt.Errorf("%w: options is nil", ErrInvalidInput)
	}
	if opts.ChatID <= 0 {
		return nil, "", nil, fmt.Errorf("%w: incorrect chat ID %d", ErrInvalidInput, opts.ChatID)
	}

	query := map[string]string{
		"chat_id": fmt.Sprint(opts.ChatID),
	}
	if opts.Limit > 0 {
		query["limit"] = fmt.Sprint(opts.Limit)
	}
	if opts.Cursor != "" {
		query["cursor"] = opts.Cursor
	}
	if opts.SortField != "" {
		query["sort["+opts.SortField+"]"] = opts.SortOrder // если order не задан, API использует desc
	} else if opts.SortOrder != "" {
		query["sort[id]"] = opts.SortOrder
	}

	resp, err := m.client.R().
		SetContext(ctx).
		SetQueryParams(query).
		SetHeader("Content-Type", "application/json").
		Get(messagesURL)
	if err != nil {
		return nil, "", resp, err
	}
	if resp.StatusCode() != 200 {
		return nil, "", resp, fmt.Errorf("%w: %d", ErrResponseCode, resp.StatusCode())
	}

	var raw messagesResponseRaw
	if err := json.Unmarshal(resp.Body(), &raw); err != nil {
		return nil, "", resp, fmt.Errorf("%w: %v", ErrResponseDecode, err)
	}
	return raw.Data, raw.Meta.Paginate.NextPage, resp, nil
}

// ListChatMessagesAll автоматически обходит все страницы и возвращает все сообщения чата
func (m *Messages) ListChatMessagesAll(ctx context.Context, chatID int) ([]MessageResponse, error) {
	if chatID <= 0 {
		return nil, fmt.Errorf("%w: incorrect chat ID %d", ErrInvalidInput, chatID)
	}
	limit := 50
	var all []MessageResponse
	cursor := ""
	for {
		opts := &ChatMessagesOptions{
			ChatID: chatID,
			CursorPagination: CursorPagination{
				Limit:  limit,
				Cursor: cursor,
			},
		}
		messages, next, _, err := m.ListChatMessages(ctx, opts)
		if err != nil {
			return nil, err
		}
		all = append(all, messages...)
		if next == "" {
			break
		}
		cursor = next
	}
	return all, nil
}

// Send отправляет новое сообщение
func (m *Messages) Send(ctx context.Context, msg *Message, linkPreview bool) (*MessageResponse, *resty.Response, error) {
	if msg == nil {
		return nil, nil, fmt.Errorf("%w: message is nil", ErrInvalidInput)
	}
	if msg.EntityID <= 0 {
		return nil, nil, fmt.Errorf("%w: incorrect entity ID %d", ErrInvalidInput, msg.EntityID)
	}
	if msg.Content == "" && len(msg.Files) == 0 {
		return nil, nil, fmt.Errorf("%w: message must have content or files", ErrInvalidInput)
	}

	body := OutgoingMessage{
		Message:     msg,
		LinkPreview: linkPreview,
	}
	resp, err := m.client.R().
		SetContext(ctx).
		SetBody(body).
		Post(messagesURL)
	if err != nil {
		return nil, resp, err
	}
	if resp.StatusCode() != 201 {
		return nil, resp, fmt.Errorf("%w: %d", ErrResponseCode, resp.StatusCode())
	}

	var raw messageResponseRaw
	if err := json.Unmarshal(resp.Body(), &raw); err != nil {
		return nil, resp, fmt.Errorf("%w: %v", ErrResponseDecode, err)
	}
	return &raw.Data, resp, nil
}

// Edit редактирует существующее сообщение
func (m *Messages) Edit(ctx context.Context, messageID int, msg *Message) (*MessageResponse, *resty.Response, error) {
	if messageID <= 0 {
		return nil, nil, fmt.Errorf("%w: incorrect message ID %d", ErrInvalidInput, messageID)
	}
	if msg == nil {
		return nil, nil, fmt.Errorf("%w: message is nil", ErrInvalidInput)
	}
	// при редактировании можно не передавать EntityID и Content (например, только кнопки)
	body := OutgoingMessage{Message: msg} // LinkPreview при редактировании не используется
	url := fmt.Sprintf("%s/%d", messagesURL, messageID)
	resp, err := m.client.R().
		SetContext(ctx).
		SetBody(body).
		Put(url)
	if err != nil {
		return nil, resp, err
	}
	if resp.StatusCode() != 200 {
		return nil, resp, fmt.Errorf("%w: %d", ErrResponseCode, resp.StatusCode())
	}
	var raw messageResponseRaw
	if err := json.Unmarshal(resp.Body(), &raw); err != nil {
		return nil, resp, fmt.Errorf("%w: %v", ErrResponseDecode, err)
	}
	return &raw.Data, resp, nil
}

// Delete удаляет сообщение
func (m *Messages) Delete(ctx context.Context, messageID int) (*resty.Response, error) {
	if messageID <= 0 {
		return nil, fmt.Errorf("%w: incorrect message ID %d", ErrInvalidInput, messageID)
	}
	url := fmt.Sprintf("%s/%d", messagesURL, messageID)
	resp, err := m.client.R().
		SetContext(ctx).
		Delete(url)
	if err != nil {
		return resp, err
	}
	if resp.StatusCode() != 204 {
		return resp, fmt.Errorf("%w: %d", ErrResponseCode, resp.StatusCode())
	}
	return resp, nil
}

// Pin закрепляет сообщение
func (m *Messages) Pin(ctx context.Context, messageID int) (*resty.Response, error) {
	if messageID <= 0 {
		return nil, fmt.Errorf("%w: incorrect message ID %d", ErrInvalidInput, messageID)
	}
	url := fmt.Sprintf("%s/%d/pin", messagesURL, messageID)
	resp, err := m.client.R().
		SetContext(ctx).
		Post(url)
	if err != nil {
		return resp, err
	}
	// 204 – успех, 409 – уже закреплено (не считаем ошибкой)
	if resp.StatusCode() != 204 && resp.StatusCode() != 409 {
		return resp, fmt.Errorf("%w: %d", ErrResponseCode, resp.StatusCode())
	}
	return resp, nil
}

// Unpin открепляет сообщение
func (m *Messages) Unpin(ctx context.Context, messageID int) (*resty.Response, error) {
	if messageID <= 0 {
		return nil, fmt.Errorf("%w: incorrect message ID %d", ErrInvalidInput, messageID)
	}
	url := fmt.Sprintf("%s/%d/pin", messagesURL, messageID)
	resp, err := m.client.R().
		SetContext(ctx).
		Delete(url)
	if err != nil {
		return resp, err
	}
	if resp.StatusCode() != 204 {
		return resp, fmt.Errorf("%w: %d", ErrResponseCode, resp.StatusCode())
	}
	return resp, nil
}
