package pachca

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

// Users
// Объект для работы с пользователями
type Users struct {
	client *resty.Client
	cache  map[int]*UserResponse
	mu     *sync.RWMutex
}

// User модель сотрудника / профиля
type User struct {
	ID               int              `json:"id"`
	FirstName        string           `json:"first_name"`
	LastName         string           `json:"last_name"`
	Nickname         string           `json:"nickname"`
	Email            string           `json:"email"`
	PhoneNumber      string           `json:"phone_number"`
	Department       string           `json:"department"`
	Title            string           `json:"title"`
	Role             string           `json:"role"` // admin, user, multi_guest, guest
	Suspended        bool             `json:"suspended"`
	InviteStatus     string           `json:"invite_status"` // confirmed, sent
	ListTags         []string         `json:"list_tags"`
	CustomProperties []CustomProperty `json:"custom_properties"`
	UserStatus       *UserStatus      `json:"user_status"` // может быть null
	Bot              bool             `json:"bot"`
	Sso              bool             `json:"sso"`
	CreatedAt        time.Time        `json:"created_at"`
	LastActivityAt   time.Time        `json:"last_activity_at"`
	TimeZone         string           `json:"time_zone"`
	ImageURL         *string          `json:"image_url"` // ссылка на аватарку
}

// CustomProperty дополнительное поле сотрудника
type CustomProperty struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	DataType string `json:"data_type"` // string, number, date, link
	Value    string `json:"value"`
}

// UserStatus статус пользователя
type UserStatus struct {
	Emoji       string                 `json:"emoji"`
	Title       string                 `json:"title"`
	ExpiresAt   *time.Time             `json:"expires_at"`
	IsAway      bool                   `json:"is_away"`
	AwayMessage *UserStatusAwayMessage `json:"away_message"`
}

// UserStatusAwayMessage сообщение при режиме «Нет на месте»
type UserStatusAwayMessage struct {
	Text string `json:"text"`
}

type UserResponse struct {
	ID             int         `json:"id"`
	InviteStatus   string      `json:"invite_status"`
	UserStatus     *UserStatus `json:"user_status"`
	Bot            bool        `json:"bot"`
	CreatedAt      string      `json:"created_at"`
	LastActivityAt string      `json:"last_activity_at"`
	TimeZone       string      `json:"time_zone"`
	ImageURL       string      `json:"image_url"`
	User
}

type UserResponseRaw struct {
	Data UserResponse `json:"data"`
}

type PaginateUsersRaw struct {
	NextPage string `json:"next_page"`
}

type MetaRaw struct {
	Paginate PaginateUsersRaw `json:"paginate"`
}

type UsersResponseRaw struct {
	Data []UserResponse `json:"data"`
	Meta MetaRaw        `json:"meta"`
}

// Get
// Метод для получения информации о сотруднике.
func (u *Users) Get(ctx context.Context, userID int) (*UserResponse, *resty.Response, error) {
	if userID <= 0 {
		return nil, nil, fmt.Errorf("invalid user ID: %d", userID)
	}

	url := fmt.Sprintf("%v/%v", usersURL, userID)
	resp, err := u.client.R().
		SetContext(ctx).
		Get(url)
	if err != nil {
		return nil, resp, err
	}
	if resp.StatusCode() != 200 {
		return nil, resp, fmt.Errorf("%w: %d", ErrResponseCode, resp.StatusCode())
	}

	var r UserResponseRaw
	err = json.Unmarshal(resp.Body(), &r)
	if err != nil {
		return nil, resp, fmt.Errorf("%w: %w", ErrResponseDecode, err)
	}

	return &r.Data, resp, nil
}

// GetWithCache
// Метод для получения информации о сотруднике используя кэш. Если пользователя нету в кэше, то будет выполнен запрос к API.
func (u *Users) GetWithCache(ctx context.Context, userID int) (*UserResponse, *resty.Response, error) {
	u.mu.RLock()
	cachedUser, found := u.cache[userID]
	u.mu.RUnlock()

	if found {
		return cachedUser, nil, nil
	}

	user, resp, err := u.Get(ctx, userID)
	if err != nil {
		return nil, resp, err
	}

	u.mu.Lock()
	u.cache[userID] = user
	u.mu.Unlock()

	return user, resp, nil
}

// getUsersPaginated Deprecated per page не работает. Актуальный метод getUsersPaginatedV2
// Получение списка пользователей с пагинацией.
func (u *Users) getUsersPaginated(ctx context.Context, options *ListUsersOptions) ([]UserResponse, *resty.Response, error) {
	resp, err := u.client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"per":   fmt.Sprint(options.Per),
			"page":  fmt.Sprint(options.Page),
			"query": options.Query,
		}).
		Get(usersURL)
	if err != nil {
		return nil, resp, err
	}

	if resp.StatusCode() != 200 {
		return nil, resp, fmt.Errorf("%w: %d", ErrResponseCode, resp.StatusCode())
	}

	var users *UsersResponseRaw
	err = json.Unmarshal(resp.Body(), &users)
	if err != nil {
		return nil, resp, fmt.Errorf("%w: %w", ErrResponseDecode, err)
	}

	return users.Data, resp, nil
}

func (u *Users) getUsersPaginatedV2(ctx context.Context, options *ListUsersOptionsV2) ([]UserResponse, string, *resty.Response, error) {
	resp, err := u.client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"limit":  fmt.Sprint(options.PaginationOptions.Limit),
			"cursor": options.PaginationOptions.Next,
			"query":  options.Query,
		}).
		Get(usersURL)
	if err != nil {
		return nil, "", resp, err
	}

	if resp.StatusCode() != 200 {
		return nil, "", resp, fmt.Errorf("%w: %d", ErrResponseCode, resp.StatusCode())
	}

	var users *UsersResponseRaw
	err = json.Unmarshal(resp.Body(), &users)
	if err != nil {
		return nil, "", resp, fmt.Errorf("%w: %w", ErrResponseDecode, err)
	}

	return users.Data, users.Meta.Paginate.NextPage, resp, nil
}

type ListUsersOptions struct {
	Query string
	PaginationOptions
}

type ListUsersOptionsV2 struct {
	Query             string
	PaginationOptions PaginationOptionsUsers
}

// List Deprecated Пагинация per page не работает
// Метод для получения актуального списка сотрудников вашей компании.
// Все параметры передаются через опции. В случае отсутствия опций будут возвращены все сотрудники.
func (u *Users) List(ctx context.Context, options *ListUsersOptions) (users []UserResponse, resp *resty.Response, err error) {
	if options == nil {
		options = &ListUsersOptions{
			PaginationOptions: PaginationOptions{
				Page: 1,
				Per:  50,
			},
		}
	}

	var usersResponse []UserResponse

	for {
		usersResponse, resp, err = u.getUsersPaginated(ctx, options)
		if err != nil {
			return nil, resp, err
		}

		if len(usersResponse) == 0 {
			break
		}

		users = append(users, usersResponse...)

		if len(usersResponse) == options.Per {
			options.Page++
		} else {
			break
		}
	}

	return users, resp, nil
}

// ListV2 Метод для получения актуального списка сотрудников вашей компании.
func (u *Users) ListV2(ctx context.Context, options *ListUsersOptionsV2) (users []UserResponse, next string, resp *resty.Response, err error) {
	if options == nil {
		options = &ListUsersOptionsV2{
			PaginationOptions: PaginationOptionsUsers{
				Limit: 50,
			},
		}
	}

	usersResponse, next, resp, err := u.getUsersPaginatedV2(ctx, options)
	if err != nil {
		return nil, "", resp, err
	}

	if len(usersResponse) == 0 {
		return usersResponse, "", resp, nil
	}

	return usersResponse, next, resp, nil
}

// Find
// Метод для поиска пользователей. Упрощенная версия метода List. Позволяет передать фильтрующий запрос и получить результаты.
// Поисковая фраза для фильтрации результатов (поиск идет по полям first_name (имя), last_name (фамилия), email (электронная почта), phone_number (телефон) и nickname (никнейм))
// resp - в случае успеха - последний успешний ответ
func (u *Users) Find(ctx context.Context, query string) (users []UserResponse, resp *resty.Response, err error) {
	options := &ListUsersOptionsV2{
		PaginationOptions: PaginationOptionsUsers{
			Limit: 50,
		},
		Query: query,
	}

	var rawResp *resty.Response

	for {
		usersResponse, next, resp, err := u.getUsersPaginatedV2(ctx, options)
		if err != nil {
			return nil, resp, err
		}

		rawResp = resp

		if len(usersResponse) == 0 {
			break
		}

		users = append(users, usersResponse...)

		// Сейчас API возврашает next даже на пустой странице.
		// Это условие для защиты на случай изменеия API, если на последней странице возвращается пустой next
		if next == "" {
			break
		}

		options.PaginationOptions.Next = next
	}

	return users, rawResp, nil
}

// Update
// Метод для редактирования сотрудника.
// Данный метод доступен для работы только с access_token администратора пространства.
func (u *Users) Update(ctx context.Context, userID int, user *User) (*UserResponse, *resty.Response, error) {
	url := fmt.Sprintf("%v/%v", usersURL, userID)

	body := struct {
		User *User `json:"user"`
	}{
		User: user,
	}

	resp, err := u.client.R().
		SetContext(ctx).
		SetBody(body).
		Put(url)
	if err != nil {
		return nil, resp, err
	}
	if resp.StatusCode() != 200 {
		return nil, resp, fmt.Errorf("%w: %d", ErrResponseCode, resp.StatusCode())
	}

	var r UserResponse
	err = json.Unmarshal(resp.Body(), &r)
	if err != nil {
		return nil, resp, fmt.Errorf("%w: %w", ErrResponseDecode, err)
	}

	return &r, resp, nil
}

// New
// Метод для создания нового сотрудника в вашей компании.
// Данный метод доступен для работы только с access_token администратора пространства.
func (u *Users) New(ctx context.Context, user *User) (*UserResponse, *resty.Response, error) {
	// TODO добавить валидацию

	body := struct {
		User *User `json:"user"`
	}{
		User: user,
	}

	resp, err := u.client.R().
		SetContext(ctx).
		SetBody(body).
		Post(usersURL)
	if err != nil {
		return nil, resp, err
	}

	if resp.StatusCode() != 201 {
		return nil, resp, fmt.Errorf("%w: %d", ErrResponseCode, resp.StatusCode())
	}

	var r UserResponseRaw
	err = json.Unmarshal(resp.Body(), &r)
	if err != nil {
		return nil, resp, fmt.Errorf("%w: %w", ErrResponseDecode, err)
	}

	return &r.Data, resp, nil
}
