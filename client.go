package pachca

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
)

const (
	apiURL      = "https://api.pachca.com/api/shared/v1"
	profileURL  = "/profile/status"
	messagesURL = "/messages"
	usersURL    = "/users"
	chatsURL    = "/chats"
	threadsURL  = "/threads"
	tagsURL     = "/group_tags"
)

var (
	ErrResponseCode   = fmt.Errorf("unexpected response code")
	ErrResponseDecode = fmt.Errorf("error json decoding body")
	ErrInvalidInput   = fmt.Errorf("invalid input data")
)

// Client
// Клиент для работы с мессенджером pachca
type Client struct {
	client   *resty.Client
	Messages *Messages
	Threads  *Threads
	Users    *Users
	Chats    *Chats
	Tags     *Tags
}

type PaginationOptions struct {
	Per  int
	Page int
}

// NewClient
// Конструктор клиента для мессенджера pachca
func NewClient(accessToken string) *Client {
	pachcaClient := resty.New()
	pachcaClient.SetBaseURL(apiURL)
	pachcaClient.SetHeader("Authorization", fmt.Sprintf("Bearer %v", accessToken))
	return &Client{
		client:   pachcaClient,
		Messages: &Messages{client: pachcaClient},
		Threads:  &Threads{client: pachcaClient},
		Users:    &Users{client: pachcaClient},
		Chats:    &Chats{client: pachcaClient},
		Tags:     &Tags{client: pachcaClient},
	}
}

// CheckConnection
// Проверка соединения в клиенте
func (c *Client) CheckConnection(ctx context.Context) error {
	resp, err := c.client.R().
		SetContext(ctx).
		Get(profileURL)
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("%v: %v", ErrResponseCode, resp.StatusCode())
	}

	var r UserResponse
	err = json.Unmarshal(resp.Body(), &r)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrResponseDecode, err)
	}
	return nil
}
