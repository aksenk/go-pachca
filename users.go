package pachca

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
)

// Users
// Объект для работы с пользователями
type Users struct {
	client *resty.Client
}

type UserCustomProperty struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	DataType string `json:"data_type"`
	Value    string `json:"value"`
}

type UserStatus struct {
	Emoji     string `json:"emoji"`
	Title     string `json:"title"`
	ExpiresAt string `json:"expires_at"`
}

type User struct {
	FirstName        string               `json:"first_name"`
	LastName         string               `json:"last_name"`
	Nickname         string               `json:"nickname"`
	Email            string               `json:"email"`
	PhoneNumber      string               `json:"phone_number"`
	Department       string               `json:"department"`
	Title            string               `json:"title"`
	Role             string               `json:"role"`
	Suspended        bool                 `json:"suspended"`
	ListTags         []string             `json:"list_tags"`
	CustomProperties []UserCustomProperty `json:"custom_properties"`
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

type UsersResponseRaw struct {
	Data []UserResponse `json:"data"`
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

// getUsersPaginated
// Получение списка пользователей с пагинацией.
func (u *Users) getUsersPaginated(ctx context.Context, options *ListUsersOptions) ([]UserResponse, *resty.Response, error) {
	resp, err := u.client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"per":   fmt.Sprint(options.per),
			"page":  fmt.Sprint(options.page),
			"query": options.query,
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

type ListUsersOptions struct {
	query string
	PaginationOptions
}

// List
// Метод для получения актуального списка сотрудников вашей компании.
// Все параметры передаются через опции. В случае отсутствия опций будут возвращены все сотрудники.
func (u *Users) List(ctx context.Context, options *ListUsersOptions) (users []UserResponse, resp *resty.Response, err error) {
	if options == nil {
		options = &ListUsersOptions{
			page: 1,
			per:  50,
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

		if len(usersResponse) < options.per {
			break
		}

		options.page++
	}

	return users, resp, nil
}

// Find
// Метод для поиска пользователей. Упрощенная версия метода List. Позволяет передать фильтрующий запрос и получить результаты.
// Поисковая фраза для фильтрации результатов (поиск идет по полям first_name (имя), last_name (фамилия), email (электронная почта), phone_number (телефон) и nickname (никнейм))
func (u *Users) Find(ctx context.Context, query string) (users []UserResponse, resp *resty.Response, err error) {
	options := &ListUsersOptions{
		per:   50,
		page:  1,
		query: query,
	}

	for {
		usersResponse, resp, err := u.getUsersPaginated(ctx, options)
		if err != nil {
			return nil, resp, err
		}

		if len(usersResponse) == 0 {
			break
		}

		users = append(users, usersResponse...)

		if len(usersResponse) < options.per {
			break
		}

		options.page++
	}
	return users, resp, nil
}

// Update
// Метод для редактирования сотрудника.
// Данный метод доступен для работы только с access_token администратора пространства.
func (u *Users) Update(ctx context.Context, userID int, user User) (*UserResponse, *resty.Response, error) {
	url := fmt.Sprintf("%v/%v", usersURL, userID)
	resp, err := u.client.R().
		SetContext(ctx).
		SetBody(user).
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
