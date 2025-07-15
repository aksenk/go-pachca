package pachca

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"strconv"
	"time"
)

const (
	defaultApiURL = "https://api.pachca.com/api/shared/v1"
	profileURL    = "/profile/status"
	messagesURL   = "/messages"
	usersURL      = "/users"
	chatsURL      = "/chats"
	threadsURL    = "/threads"
	tagsURL       = "/group_tags"
)

var (
	ErrResponseCode    = fmt.Errorf("unexpected response code")
	ErrResponseDecode  = fmt.Errorf("error json decoding body")
	ErrInvalidInput    = fmt.Errorf("invalid input data")
	ErrNoClientOptions = fmt.Errorf("client options cannot be nil")
	ErrNoAccessToken   = fmt.Errorf("access token cannot be empty")
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

type RetryObserver func(meta RetryMeta)

type RetryMeta struct {
	Attempt      int
	Wait         time.Duration
	ResponseCode int
	URL          string
}

type ClientOptions struct {
	ApiURL            string
	AccessToken       string
	RetryCount        int
	RetryWait         time.Duration
	RetryMaxWait      time.Duration
	DisableRetryOn5XX bool
	RetryObserver     RetryObserver
}

// NewClient
// Конструктор клиента для мессенджера pachca
func NewClient(options *ClientOptions) (*Client, error) {
	if options == nil {
		return nil, ErrNoClientOptions
	}

	if options.AccessToken == "" {
		return nil, ErrNoAccessToken
	}

	if options.ApiURL == "" {
		options.ApiURL = defaultApiURL
	}

	if options.RetryCount <= 0 {
		options.RetryCount = 4
	}

	if options.RetryWait <= 0 {
		options.RetryWait = 1
	}

	if options.RetryMaxWait <= 0 {
		options.RetryMaxWait = 30
	}

	observer := options.RetryObserver

	pachcaClient := resty.New().
		SetBaseURL(options.ApiURL).
		SetHeader("Authorization", fmt.Sprintf("Bearer %v", options.AccessToken)).
		SetRetryCount(options.RetryCount).
		SetRetryWaitTime(options.RetryWait).
		SetRetryMaxWaitTime(options.RetryMaxWait).
		AddRetryCondition(
			func(r *resty.Response, err error) bool {
				// ретрай на ошибки сети, таймауты и т.д.
				if err != nil {
					return true
				}
				// ретрай на 429 (Too Many Requests)
				if r.StatusCode() == 429 {
					return true
				}
				// ретрай на 5xx ошибки, если не отключено
				if !options.DisableRetryOn5XX {
					if r.StatusCode() >= 500 && r.StatusCode() < 600 { // Server errors
						return true
					}
				}
				return false
			}).
		AddRetryHook(
			func(r *resty.Response, err error) {
				wait := 0 * time.Second
				if retryAfter := r.Header().Get("Retry-After"); retryAfter != "" {
					if secs, err := strconv.Atoi(retryAfter); err == nil && secs > 0 {
						wait = time.Duration(secs) * time.Second
						time.Sleep(wait)
					}
				}
				// Если есть наблюдатель, вызываем его с метаданными,
				// чтобы можно было отслеживать попытки ретрая со стороны приложения
				if options.RetryObserver != nil {
					observer(RetryMeta{
						Attempt:      r.Request.Attempt,
						Wait:         wait,
						ResponseCode: r.StatusCode(),
						URL:          r.Request.URL,
					})
				}
			})
	return &Client{
		client:   pachcaClient,
		Messages: &Messages{client: pachcaClient},
		Threads:  &Threads{client: pachcaClient},
		Users:    &Users{client: pachcaClient},
		Chats:    &Chats{client: pachcaClient},
		Tags:     &Tags{client: pachcaClient},
	}, nil
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
