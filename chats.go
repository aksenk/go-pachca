package pachca

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

// Chats
// Объект для работы с чатами
type Chats struct {
	client *resty.Client
}

// Chat
// Объект описывающий чат
type Chat struct {
	Name        string `json:"name"`          // Название
	MemberIDs   []int  `json:"member_ids"`    // Массив идентификаторов пользователей, которые станут участниками
	GroupTagIDs []int  `json:"group_tag_ids"` // Массив идентификаторов тегов, которые станут участниками
	Channel     bool   `json:"channel"`       // Тип: беседа (по умолчанию, false) или канал (true)
	Public      bool   `json:"public"`        // Доступ: закрытый (по умолчанию, false) или открытый (true)
}

// ChatResponse
// Объект получаемый в ответах API описывающий чат с дополнительными неизменяемыми параметрами
type ChatResponse struct {
	ID            int       `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	OwnerID       int       `json:"owner_id"`
	LastMessageAt time.Time `json:"last_message_at"`
	MeetRoomURL   string    `json:"meet_room_url"`
	Chat
}

// ChatResponseRaw
// Объект для хранения сырого ответа с чатом из pachca API
type ChatResponseRaw struct {
	Data ChatResponse `json:"data"`
}

// ChatsResponseRaw
// Объект для хранения сырого ответа со списком чатов из pachca API
type ChatsResponseRaw struct {
	Data []ChatResponse `json:"data"`
}

// Get
// Метод для получения информации о беседе или канале.
func (c *Chats) Get(ctx context.Context, chatID int) (*ChatResponse, *resty.Response, error) {
	if chatID <= 0 {
		return nil, nil, fmt.Errorf("%w: incorrect chat ID %d", ErrInvalidInput, chatID)
	}

	url := fmt.Sprintf("%v/%v", chatsURL, chatID)
	resp, err := c.client.R().
		SetContext(ctx).
		Get(url)
	if err != nil {
		return nil, resp, err
	}

	if resp.StatusCode() != 200 {
		return nil, resp, fmt.Errorf("%w: %v", ErrResponseCode, resp.StatusCode())
	}

	var r *ChatResponseRaw
	err = json.Unmarshal(resp.Body(), &r)
	if err != nil {
		return nil, resp, fmt.Errorf("%w: %v", ErrResponseDecode, err)
	}

	return &r.Data, resp, nil
}

// New
// Метод для создания новой беседы или нового канала.
// Для создания личной переписки 1 на 1 с пользователем воспользуйтесь методом Новое сообщение.
// При создании беседы или канала вы автоматически становитесь участником.
func (c *Chats) New(ctx context.Context, chat Chat) (*ChatResponse, *resty.Response, error) {
	if chat.Name == "" {
		return nil, nil, fmt.Errorf("%w: empty name", ErrInvalidInput)
	}

	url := fmt.Sprintf("%v", chatsURL)
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(chat).
		Post(url)
	if err != nil {
		return nil, resp, err
	}

	if resp.StatusCode() != 201 {
		return nil, resp, fmt.Errorf("%w: %v", ErrResponseCode, resp.StatusCode())
	}

	var r *ChatResponseRaw
	err = json.Unmarshal(resp.Body(), &r)
	if err != nil {
		return nil, resp, fmt.Errorf("%w: %v", ErrResponseDecode, err)
	}

	return &r.Data, resp, nil
}

// Update
// Метод для обновления параметров беседы или канала.
// Менять можно только параметры name и public
func (c *Chats) Update(ctx context.Context, chatID int, chat Chat) (*ChatResponse, *resty.Response, error) {
	if chatID <= 0 {
		return nil, nil, fmt.Errorf("%w: incorrect chat ID %d", ErrInvalidInput, chatID)
	}

	url := fmt.Sprintf("%v/%v", chatsURL, chatID)
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(chat).
		Put(url)
	if err != nil {
		return nil, resp, err
	}

	if resp.StatusCode() != 200 {
		return nil, resp, fmt.Errorf("%w: %v", ErrResponseCode, resp.StatusCode())
	}

	var r *ChatResponseRaw
	err = json.Unmarshal(resp.Body(), &r)
	if err != nil {
		return nil, resp, fmt.Errorf("%w: %v", ErrResponseDecode, err)
	}

	return &r.Data, resp, nil
}

// Archive
// Метод для отправки беседы или канала в архив.
func (c *Chats) Archive(ctx context.Context, chatID int) (*resty.Response, error) {
	if chatID <= 0 {
		return nil, fmt.Errorf("%w: incorrect chat ID %d", ErrInvalidInput, chatID)
	}

	url := fmt.Sprintf("%v/%v/archive", chatsURL, chatID)
	resp, err := c.client.R().
		SetContext(ctx).
		Post(url)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode() != 200 {
		return resp, fmt.Errorf("%w: %v", ErrResponseCode, resp.StatusCode())
	}

	return resp, nil
}

// Unarchive
// Метод для возвращения беседы или канала из архива.
func (c *Chats) Unarchive(ctx context.Context, chatID int) (*resty.Response, error) {
	if chatID <= 0 {
		return nil, fmt.Errorf("%w: incorrect chat ID %d", ErrInvalidInput, chatID)
	}

	url := fmt.Sprintf("%v/%v/unarchive", chatsURL, chatID)
	resp, err := c.client.R().
		SetContext(ctx).
		Post(url)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode() != 200 {
		return resp, fmt.Errorf("%w: %v", ErrResponseCode, resp.StatusCode())
	}

	return resp, nil
}

// ListChatsOptions
// Опции для поиска чатов
type ListChatsOptions struct {
	Sort                string
	Availability        string
	LastMessageAtAfter  string
	LastMessageAtBefore string
	PaginationOptions
}

// List
// Получение списка всех бесед и каналов
func (c *Chats) List(ctx context.Context, options *ListChatsOptions) (allChats []ChatResponse, resp *resty.Response, err error) {
	if options.Sort != "asc" && options.Sort != "desc" {
		options.Sort = "desc"
	}
	if options.Per <= 0 {
		options.Per = 25
	}
	if options.Page <= 0 {
		options.Page = 1
	}
	if options.Availability != "is_member" && options.Availability != "public" {
		options.Availability = "is_member"
	}
	// TODO last_message_at_after last_message_at_before

	var chats []ChatResponse

	for {
		chats, resp, err = c.getChatsPaginated(ctx, options)
		if err != nil {
			return allChats, resp, err
		}
		if len(chats) == 0 {
			break
		}

		allChats = append(allChats, chats...)

		if len(chats) < options.Per {
			break
		}

		options.Page++
	}

	return allChats, resp, nil

}

// getChatsPaginated
// Получение списка чатов с пагинацией
func (c *Chats) getChatsPaginated(ctx context.Context, options *ListChatsOptions) ([]ChatResponse, *resty.Response, error) {
	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"Page":         fmt.Sprint(options.Page),
			"Per":          fmt.Sprint(options.Per),
			"Availability": options.Availability,
			"Sort":         options.Sort,
			//"last_message_at_after":  options.LastMessageAtAfter,
			//"last_message_at_before": options.LastMessageAtBefore,
		}).
		SetContext(ctx).
		Get(chatsURL)
	if err != nil {
		return nil, resp, err
	}
	if resp.StatusCode() != 200 {
		return nil, resp, fmt.Errorf("%w: %d", ErrResponseCode, resp.StatusCode())
	}

	var chats ChatsResponseRaw
	err = json.Unmarshal(resp.Body(), &chats)
	if err != nil {
		return nil, resp, fmt.Errorf("%w: %w", ErrResponseDecode, err)
	}

	return chats.Data, resp, nil
}

// Find
// Поиск чата по имени
func (c *Chats) Find(ctx context.Context, name string) (chat ChatResponse, resp *resty.Response, err error) {
	options := &ListChatsOptions{
		Availability: "public",
		PaginationOptions: PaginationOptions{
			Page: 1,
			Per:  50,
		},
	}
	page := 1
	perPage := 50
	var chats []ChatResponse

	for {
		chats, resp, err = c.getChatsPaginated(ctx, options)
		if err != nil {
			return chat, resp, err
		}
		if len(chats) == 0 {
			break
		}

		for _, c := range chats {
			if c.Name == name {
				return c, resp, nil
			}
		}

		if len(chats) < perPage {
			break
		}

		page++
	}

	return chat, resp, nil

}
