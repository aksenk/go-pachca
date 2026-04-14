// profile.go
package pachca

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

// ---------- URL-ы ----------
const (
	tokenInfoURL   = "/oauth/token/info"
	profileInfoURL = "/profile"
)

// ---------- Модели данных ----------

// AccessTokenInfo информация о токене доступа
type AccessTokenInfo struct {
	ID         int64      `json:"id"`           // идентификатор токена
	Token      string     `json:"token"`        // маскированный токен (первые 8 и последние 4 символа)
	Name       *string    `json:"name"`         // пользовательское имя токена
	UserID     int64      `json:"user_id"`      // идентификатор владельца токена
	Scopes     []string   `json:"scopes"`       // список скоупов токена
	CreatedAt  time.Time  `json:"created_at"`   // дата создания токена
	RevokedAt  *time.Time `json:"revoked_at"`   // дата отзыва токена (null, если не отозван)
	ExpiresIn  *int32     `json:"expires_in"`   // время жизни токена в секундах (null для бессрочных)
	LastUsedAt *time.Time `json:"last_used_at"` // дата последнего использования
}

// ---------- Обёртки ответов ----------

type tokenInfoResponse struct {
	Data AccessTokenInfo `json:"data"`
}

type profileResponse struct {
	Data User `json:"data"`
}

// ---------- Клиент для профиля и токена ----------

type Profile struct {
	client *resty.Client
}

// NewProfile создаёт новый объект для работы с профилем и токеном
func NewProfile(client *resty.Client) *Profile {
	return &Profile{client: client}
}

// GetTokenInfo возвращает информацию о текущем OAuth токене
func (p *Profile) GetTokenInfo(ctx context.Context) (*AccessTokenInfo, *resty.Response, error) {
	resp, err := p.client.R().
		SetContext(ctx).
		Get(tokenInfoURL)
	if err != nil {
		return nil, resp, err
	}
	if resp.StatusCode() != 200 {
		return nil, resp, fmt.Errorf("%w: %d", ErrResponseCode, resp.StatusCode())
	}

	var raw tokenInfoResponse
	if err := json.Unmarshal(resp.Body(), &raw); err != nil {
		return nil, resp, fmt.Errorf("%w: %v", ErrResponseDecode, err)
	}
	return &raw.Data, resp, nil
}

// GetProfile возвращает информацию о профиле пользователя, которому принадлежит токен
func (p *Profile) GetProfile(ctx context.Context) (*User, *resty.Response, error) {
	resp, err := p.client.R().
		SetContext(ctx).
		Get(profileInfoURL)
	if err != nil {
		return nil, resp, err
	}
	if resp.StatusCode() != 200 {
		return nil, resp, fmt.Errorf("%w: %d", ErrResponseCode, resp.StatusCode())
	}

	var raw profileResponse
	if err := json.Unmarshal(resp.Body(), &raw); err != nil {
		return nil, resp, fmt.Errorf("%w: %v", ErrResponseDecode, err)
	}
	return &raw.Data, resp, nil
}
