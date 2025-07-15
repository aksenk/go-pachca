package pachca

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// AddUsers
// Метод для добавления пользователей в состав участников беседы или канала.
func (c *Chats) AddUsers(ctx context.Context, chatID int, membersIDs []int, silent bool) (*resty.Response, error) {
	if chatID <= 0 {
		return nil, fmt.Errorf("%w: incorrect chat ID %d", ErrInvalidInput, chatID)
	}

	if len(membersIDs) == 0 {
		return nil, fmt.Errorf("%w: empty user IDs", ErrInvalidInput)
	}

	body := struct {
		MemberIDs []int `json:"member_ids"`
		Silent    bool  `json:"silent"`
	}{
		MemberIDs: membersIDs,
		Silent:    silent,
	}

	url := fmt.Sprintf("%v/%v/members", chatsURL, chatID)
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(body).
		Post(url)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode() != 200 {
		return resp, fmt.Errorf("%w: %v", ErrResponseCode, resp.StatusCode())
	}

	return resp, nil
}

// RemoveUsers
// Метод для исключения пользователя из состава участников беседы или канала.
// Если пользователь является владельцем чата, то исключить его нельзя. Он может только самостоятельно выйти из чата, воспользовавшись методом Выход из беседы или канала.
func (c *Chats) RemoveUsers(ctx context.Context, chatID, userID int) (*resty.Response, error) {
	if chatID <= 0 {
		return nil, fmt.Errorf("%w: incorrect chat ID %d", ErrInvalidInput, chatID)
	}

	if userID <= 0 {
		return nil, fmt.Errorf("%w: incorrect user ID %d", ErrInvalidInput, userID)
	}

	url := fmt.Sprintf("%v/%v/members/%v", chatsURL, chatID, userID)
	resp, err := c.client.R().
		SetContext(ctx).
		Delete(url)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode() != 200 {
		return resp, fmt.Errorf("%w: %v", ErrResponseCode, resp.StatusCode())
	}

	return resp, nil
}

// AddTags
// Метод для добавления тегов в состав участников беседы или канала.
func (c *Chats) AddTags(ctx context.Context, chatID int, groupTagsIDs []int) (*resty.Response, error) {
	if chatID <= 0 {
		return nil, fmt.Errorf("%w: incorrect chat ID %d", ErrInvalidInput, chatID)
	}

	if len(groupTagsIDs) == 0 {
		return nil, fmt.Errorf("%w: empty group tags IDs", ErrInvalidInput)
	}

	body := struct {
		GroupTagIDs []int `json:"group_tag_ids"`
	}{
		GroupTagIDs: groupTagsIDs,
	}

	url := fmt.Sprintf("%v/%v/group_tags", chatsURL, chatID)
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(body).
		Post(url)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode() != 200 {
		return resp, fmt.Errorf("%w: %v", ErrResponseCode, resp.StatusCode())
	}

	return resp, nil
}

// RemoveTags
// Метод для исключения тега из состава участников беседы или канала.
func (c *Chats) RemoveTags(ctx context.Context, chatID int, groupTagID int) (*resty.Response, error) {
	if chatID <= 0 {
		return nil, fmt.Errorf("%w: incorrect chat ID %d", ErrInvalidInput, chatID)
	}

	if groupTagID <= 0 {
		return nil, fmt.Errorf("%w: incorrect group tag IDs", ErrInvalidInput)
	}

	url := fmt.Sprintf("%v/%v/group_tags/%v", chatsURL, chatID, groupTagID)
	resp, err := c.client.R().
		SetContext(ctx).
		Delete(url)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode() != 200 {
		return resp, fmt.Errorf("%w: %v", ErrResponseCode, resp.StatusCode())
	}

	return resp, nil
}

// EditRole
// Метод для редактирования роли пользователя или бота в беседе или канале.
// Владельцу чата роль изменить нельзя. Он всегда имеет права Админа в чате.
func (c *Chats) EditRole(ctx context.Context, chatID, memberID int, role string) (*resty.Response, error) {
	if chatID <= 0 {
		return nil, fmt.Errorf("%w: incorrect chat ID %d", ErrInvalidInput, chatID)
	}
	if memberID <= 0 {
		return nil, fmt.Errorf("%w: incorrect member ID %d", ErrInvalidInput, memberID)
	}
	if role == "" {
		return nil, fmt.Errorf("%w: empty role", ErrInvalidInput)
	}

	body := struct {
		Role string `json:"role"`
	}{
		Role: role,
	}

	url := fmt.Sprintf("%v/%v/members/%v", chatsURL, chatID, memberID)
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(body).
		Put(url)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode() != 200 {
		return resp, fmt.Errorf("%w: %v", ErrResponseCode, resp.StatusCode())
	}

	return resp, nil
}

// Leave
// Метод для самостоятельного выхода из беседы или канала.
func (c *Chats) Leave(ctx context.Context, chatID int) (*resty.Response, error) {
	if chatID <= 0 {
		return nil, fmt.Errorf("%w: incorrect chat ID %d", ErrInvalidInput, chatID)
	}

	url := fmt.Sprintf("%v/%v/group_tags", chatsURL, chatID)
	resp, err := c.client.R().
		SetContext(ctx).
		Delete(url)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode() != 200 {
		return resp, fmt.Errorf("%w: %v", ErrResponseCode, resp.StatusCode())
	}

	return resp, nil
}
