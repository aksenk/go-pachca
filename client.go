package pachca

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
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
	client    *resty.Client
	Messages  *Messages
	Threads   *Threads
	Users     *Users
	Chats     *Chats
	Tags      *Tags
	Reactions *Reactions
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
	Method       string
	Context      context.Context
}

type ClientOptions struct {
	ApiURL          string
	AccessToken     string
	RetryCount      int
	RetryWait       time.Duration
	RetryMaxWait    time.Duration
	RetryObserver   RetryObserver
	RequestsTimeout time.Duration
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
		options.RetryWait = 1 * time.Second
	}

	if options.RetryMaxWait <= 0 {
		options.RetryMaxWait = 30 * time.Second
	}

	if options.RequestsTimeout <= 0 {
		options.RequestsTimeout = 10 * time.Second
	}

	observer := options.RetryObserver

	pachcaClient := resty.New().
		SetLogger(nil).
		SetTimeout(options.RequestsTimeout).
		SetBaseURL(options.ApiURL).
		SetHeader("Authorization", fmt.Sprintf("Bearer %v", options.AccessToken)).
		SetRetryCount(options.RetryCount).
		SetRetryWaitTime(options.RetryWait).
		SetRetryMaxWaitTime(options.RetryMaxWait).
		SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
			if retryAfter := resp.Header().Get("Retry-After"); retryAfter != "" {
				if secs, err := strconv.Atoi(retryAfter); err == nil && secs > 0 {
					return time.Duration(secs) * time.Second, nil
				}
				// Обработка формата даты (если необходимо)
				if t, err := time.Parse(time.RFC1123, retryAfter); err == nil {
					return time.Until(t), nil
				}
			}
			return 0, nil // Использует стандартный backoff
		}).
		AddRetryCondition(
			func(r *resty.Response, err error) bool {
				if err != nil {
					var netErr net.Error
					if errors.As(err, &netErr) && netErr.Timeout() {
						return true
					}
					// Явно обрабатываем другие типы ошибок
					if errors.Is(err, io.EOF) || strings.Contains(err.Error(), "connection reset") {
						return true
					}
					return false
				}

				if r != nil {
					code := r.StatusCode()
					if code == 429 || code >= 500 {
						return true
					}
				}
				return false
			}).
		AddRetryHook(
			func(r *resty.Response, err error) {
				wait := 0 * time.Second

				// Если есть наблюдатель, вызываем его с метаданными,
				// чтобы можно было отслеживать попытки ретрая со стороны приложения
				if options.RetryObserver != nil {
					observer(RetryMeta{
						Attempt:      r.Request.Attempt,
						Wait:         wait,
						ResponseCode: r.StatusCode(),
						URL:          r.Request.URL,
						Context:      r.Request.Context(),
						Method:       r.Request.Method,
					})
				}
			})
	return &Client{
		client:    pachcaClient,
		Messages:  &Messages{client: pachcaClient},
		Threads:   &Threads{client: pachcaClient},
		Users:     &Users{client: pachcaClient},
		Chats:     &Chats{client: pachcaClient},
		Tags:      &Tags{client: pachcaClient},
		Reactions: &Reactions{client: pachcaClient},
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
