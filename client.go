package pachca

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
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
	uploadsURL    = "/uploads"
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
	Files     *Files
	Views     *Views
	Profile   *Profile
}

type PaginationOptions struct {
	Per  int
	Page int
}

type PaginationOptionsUsers struct {
	Limit int
	Next  string
}

type RetryObserver func(meta RetryMeta)

type RetryMeta struct {
	Attempt      int
	ResponseCode int
	URL          string
	Method       string
	Context      context.Context
	Error        error
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
		//SetLogger(nil).
		SetTimeout(options.RequestsTimeout).
		SetBaseURL(options.ApiURL).
		SetHeader("Authorization", fmt.Sprintf("Bearer %v", options.AccessToken)).
		SetRetryCount(options.RetryCount).
		SetRetryWaitTime(options.RetryWait).
		SetRetryMaxWaitTime(options.RetryMaxWait).
		AddRetryCondition(
			func(r *resty.Response, err error) bool {
				//defer func() {
				//	if rec := recover(); rec != nil {
				//		fmt.Printf("panic in retry hook condition: %+v (Request=%+v, Response=%+v)\n", rec, r.Request, r)
				//	}
				//}()

				if err != nil {
					var netErr net.Error
					if errors.As(err, &netErr) && netErr.Timeout() {
						return true
					}
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
				//defer func() {
				//	if rec := recover(); rec != nil {
				//		fmt.Printf("panic in retry hook: %+v (Request=%+v, Response=%+v)\n", rec, r.Request, r)
				//	}
				//}()
				// Если есть наблюдатель, вызываем его с метаданными,
				// чтобы можно было отслеживать попытки ретрая со стороны приложения
				if options.RetryObserver != nil {
					meta := RetryMeta{
						Error: err,
					}
					if r != nil {
						if r.Request != nil {
							meta.Attempt = r.Request.Attempt
							meta.URL = r.Request.URL
							meta.Method = r.Request.Method
							meta.Context = r.Request.Context()
						}
						meta.ResponseCode = r.StatusCode()
					}

					// Вызываем наблюдателя с собранной информацией
					observer(meta)
				}
			})
	return &Client{
		client:   pachcaClient,
		Messages: &Messages{client: pachcaClient},
		Threads:  &Threads{client: pachcaClient},
		Users: &Users{
			client: pachcaClient,
			cache:  make(map[int]*UserResponse),
			mu:     &sync.RWMutex{},
		},
		Chats: &Chats{
			client: pachcaClient,
			cache:  make(map[int]*ChatResponse),
			mu:     &sync.RWMutex{},
		},
		Tags:      &Tags{client: pachcaClient},
		Reactions: &Reactions{client: pachcaClient},
		Files:     &Files{client: pachcaClient},
		Views:     &Views{client: pachcaClient},
		Profile:   &Profile{client: pachcaClient},
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
