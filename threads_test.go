package pachca

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestThreads_New_Success(t *testing.T) {
	client := resty.New()
	client.SetTransport(&mockTransport{
		response: &http.Response{
			StatusCode: http.StatusCreated,
			Body:       mockBody(`{"data": {"id": 1, "chat_id": 123, "message_id": 456, "message_chat_id": 789, "updated_at": "2023-01-01T00:00:00Z"}}`),
		},
	})

	threads := &Threads{client: client}
	ctx := context.Background()

	resp, err := threads.New(ctx, 456)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, resp.Data.ID)
	assert.Equal(t, 123, resp.Data.ChatID)
}

func TestThreads_New_InvalidMessageID(t *testing.T) {
	client := resty.New()
	threads := &Threads{client: client}
	ctx := context.Background()

	resp, err := threads.New(ctx, 0)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid message ID")
}

func TestThreads_New_ErrorResponse(t *testing.T) {
	client := resty.New()
	client.SetTransport(&mockTransport{
		response: &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       mockBody(`{"error": "bad request"}`),
		},
	})

	threads := &Threads{client: client}
	ctx := context.Background()

	resp, err := threads.New(ctx, 456)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "400")
}

func TestThreads_Get_Success(t *testing.T) {
	client := resty.New()
	client.SetTransport(&mockTransport{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       mockBody(`{"data": {"id": 1, "chat_id": 123, "message_id": 456, "message_chat_id": 789, "updated_at": "2023-01-01T00:00:00Z"}}`),
		},
	})

	threads := &Threads{client: client}
	ctx := context.Background()

	resp, err := threads.Get(ctx, 1)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, resp.Data.ID)
	assert.Equal(t, 123, resp.Data.ChatID)
}

func TestThreads_Get_InvalidThreadID(t *testing.T) {
	client := resty.New()
	threads := &Threads{client: client}
	ctx := context.Background()

	resp, err := threads.Get(ctx, 0)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid thread ID")
}

func TestThreads_Get_ErrorResponse(t *testing.T) {
	client := resty.New()
	client.SetTransport(&mockTransport{
		response: &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       mockBody(`{"error": "not found"}`),
		},
	})

	threads := &Threads{client: client}
	ctx := context.Background()

	resp, err := threads.Get(ctx, 1)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "404")
}
