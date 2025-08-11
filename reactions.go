package pachca

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

type WebhookReaction struct {
	Type      string    `json:"type"`
	Event     string    `json:"event"`
	MessageID int       `json:"message_id"`
	Code      string    `json:"code"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
}

type Reactions struct {
	client *resty.Client
}

type ReactionOutgoing struct {
	Code string `json:"code"`
}

type ReactionResponse struct {
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	Code      string    `json:"code"`
}

type ReactionResponseRaw struct {
	Data []ReactionResponse `json:"data"`
}

func (r *Reactions) Add(ctx context.Context, messageID int, reactionCode string) (*resty.Response, error) {
	if messageID <= 0 {
		return nil, fmt.Errorf("%w: incorrect message ID %d", ErrInvalidInput, messageID)
	}
	if reactionCode == "" {
		return nil, fmt.Errorf("%w: reaction code cannot be empty", ErrInvalidInput)
	}

	url := fmt.Sprintf("%v/%v/reactions", messagesURL, messageID)
	resp, err := r.client.R().
		SetContext(ctx).
		SetBody(ReactionOutgoing{Code: reactionCode}).
		Post(url)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode() != 200 {
		return resp, fmt.Errorf("%w: %d", ErrResponseCode, resp.StatusCode())
	}

	return resp, err
}

func (r *Reactions) Del(ctx context.Context, messageID int, reactionCode string) (*resty.Response, error) {
	if messageID <= 0 {
		return nil, fmt.Errorf("%w: incorrect message ID %d", ErrInvalidInput, messageID)
	}
	if reactionCode == "" {
		return nil, fmt.Errorf("%w: reaction code cannot be empty", ErrInvalidInput)
	}

	url := fmt.Sprintf("%v/%v/reactions", messagesURL, messageID)
	resp, err := r.client.R().
		SetContext(ctx).
		SetBody(ReactionOutgoing{Code: reactionCode}).
		Delete(url)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode() != 200 {
		return resp, fmt.Errorf("%w: %d", ErrResponseCode, resp.StatusCode())
	}

	return resp, err
}

func (r *Reactions) List(ctx context.Context, messageID int, options *PaginationOptions) ([]ReactionResponse, *resty.Response, error) {
	if messageID <= 0 {
		return nil, nil, fmt.Errorf("%w: incorrect message ID %d", ErrInvalidInput, messageID)
	}

	if options == nil {
		options = &PaginationOptions{
			Page: 1,
			Per:  50,
		}
	}

	url := fmt.Sprintf("%v/%v/reactions", messagesURL, messageID)
	resp, err := r.client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"per":  fmt.Sprint(options.Per),
			"page": fmt.Sprint(options.Page),
		}).
		Get(url)
	if err != nil {
		return nil, resp, err
	}

	if resp.StatusCode() != 200 {
		return nil, resp, fmt.Errorf("%w: %d", ErrResponseCode, resp.StatusCode())
	}

	var rData ReactionResponseRaw
	err = json.Unmarshal(resp.Body(), &rData)
	if err != nil {
		return nil, resp, fmt.Errorf("%w: %w", ErrResponseDecode, err)
	}

	return rData.Data, resp, nil
}
