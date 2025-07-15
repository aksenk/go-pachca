package pachca

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestUsers_Get_Success(t *testing.T) {
	client := resty.New()
	client.SetTransport(&mockTransport{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       mockBody(`{"data": {"id": 1, "first_name": "John", "last_name": "Doe"}}`),
		},
	})

	users := &Users{client: client}
	ctx := context.Background()

	resp, httpResp, err := users.Get(ctx, 1)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, resp.Data.ID)
	assert.Equal(t, "John", resp.Data.FirstName)
}

func TestUsers_Get_NotFound(t *testing.T) {
	client := resty.New()
	client.SetTransport(&mockTransport{
		response: &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       mockBody(`{"error": "user not found"}`),
		},
	})

	users := &Users{client: client}
	ctx := context.Background()

	resp, httpResp, err := users.Get(ctx, 1)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusNotFound, httpResp.StatusCode())
}

func TestUsers_GetAll_Success(t *testing.T) {
	client := resty.New()
	client.SetTransport(&mockTransport{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       mockBody(`{"data": [{"id": 1, "first_name": "John"}, {"id": 2, "first_name": "Jane"}]}`),
		},
	})

	users := &Users{client: client}
	ctx := context.Background()

	result, httpResp, err := users.List(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Len(t, result, 2)
	assert.Equal(t, "John", result[0].FirstName)
	assert.Equal(t, "Jane", result[1].FirstName)
}

func TestUsers_Update_Success(t *testing.T) {
	client := resty.New()
	client.SetTransport(&mockTransport{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       mockBody(`{"data": {"id": 1, "first_name": "Updated"}}`),
		},
	})

	users := &Users{client: client}
	ctx := context.Background()

	user := User{ID: 1, FirstName: "Updated"}
	resp, httpResp, err := users.Update(ctx, user)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.NotNil(t, resp)
	assert.Equal(t, "Updated", resp.Data.FirstName)
}

func TestMessages_Edit_Success(t *testing.T) {
	client := resty.New()
	client.SetTransport(&mockTransport{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       mockBody(`{"data": {"id": 1, "entity_type": "chat", "entity_id": 123, "content": "Updated message"}}`),
		},
	})

	messages := &Messages{client: client}
	ctx := context.Background()

	newMessage := &Message{
		Message: struct {
			EntityType string `json:"entity_type,omitempty"`
			EntityID   int    `json:"entity_id,omitempty"`
			Content    string `json:"content,omitempty"`
		}{
			EntityID: 123,
			Content:  "Updated message",
		},
	}

	respMsg, resp, err := messages.Edit(ctx, 1, newMessage)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, respMsg)
	assert.Equal(t, 1, respMsg.Data.ID)
	assert.Equal(t, "Updated message", respMsg.Data.Content)
}

func TestMessages_Edit_InvalidInput(t *testing.T) {
	client := resty.New()
	messages := &Messages{client: client}
	ctx := context.Background()

	// Test invalid message ID
	respMsg, resp, err := messages.Edit(ctx, 0, nil)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Nil(t, respMsg)
	assert.Contains(t, err.Error(), "incorrect message ID")

	// Test nil new message
	respMsg, resp, err = messages.Edit(ctx, 1, nil)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Nil(t, respMsg)
	assert.Contains(t, err.Error(), "new message is nil")
}

func TestMessages_Edit_ErrorResponse(t *testing.T) {
	client := resty.New()
	client.SetTransport(&mockTransport{
		response: &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       mockBody(`{"error": "bad request"}`),
		},
	})

	messages := &Messages{client: client}
	ctx := context.Background()

	newMessage := &Message{
		Message: struct {
			EntityType string `json:"entity_type,omitempty"`
			EntityID   int    `json:"entity_id,omitempty"`
			Content    string `json:"content,omitempty"`
		}{
			EntityID: 123,
			Content:  "Updated message",
		},
	}

	respMsg, resp, err := messages.Edit(ctx, 1, newMessage)

	assert.Error(t, err)
	assert.Nil(t, respMsg)
	assert.NotNil(t, resp)
	assert.Contains(t, err.Error(), "400")
}

func TestMessages_Send_Success(t *testing.T) {
	client := resty.New()
	client.SetTransport(&mockTransport{
		response: &http.Response{
			StatusCode: http.StatusCreated,
			Body:       mockBody(`{"data": {"id": 1, "entity_type": "chat", "entity_id": 123, "content": "Test message"}}`),
		},
	})

	messages := &Messages{client: client}
	ctx := context.Background()

	msg := &Message{
		Message: struct {
			EntityType string `json:"entity_type,omitempty"`
			EntityID   int    `json:"entity_id,omitempty"`
			Content    string `json:"content,omitempty"`
		}{
			EntityID: 123,
			Content:  "Test message",
		},
	}

	respMsg, resp, err := messages.Send(ctx, msg)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, respMsg)
	assert.Equal(t, 1, respMsg.Data.ID)
	assert.Equal(t, "Test message", respMsg.Data.Content)
}

func TestMessages_Send_InvalidInput(t *testing.T) {
	client := resty.New()
	messages := &Messages{client: client}
	ctx := context.Background()

	// Test nil message
	respMsg, resp, err := messages.Send(ctx, nil)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Nil(t, respMsg)
	assert.Contains(t, err.Error(), "message is nil")

	// Test invalid entity ID
	msg := &Message{
		Message: struct {
			EntityType string `json:"entity_type,omitempty"`
			EntityID   int    `json:"entity_id,omitempty"`
			Content    string `json:"content,omitempty"`
		}{
			EntityID: 0,
			Content:  "Test message",
		},
	}
	respMsg, resp, err = messages.Send(ctx, msg)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Nil(t, respMsg)
	assert.Contains(t, err.Error(), "incorrect entity ID")
}

func TestMessages_Send_ErrorResponse(t *testing.T) {
	client := resty.New()
	client.SetTransport(&mockTransport{
		response: &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       mockBody(`{"error": "internal server error"}`),
		},
	})

	messages := &Messages{client: client}
	ctx := context.Background()

	msg := &Message{
		Message: struct {
			EntityType string `json:"entity_type,omitempty"`
			EntityID   int    `json:"entity_id,omitempty"`
			Content    string `json:"content,omitempty"`
		}{
			EntityID: 123,
			Content:  "Test message",
		},
	}

	respMsg, resp, err := messages.Send(ctx, msg)

	assert.Error(t, err)
	assert.Nil(t, respMsg)
	assert.NotNil(t, resp)
	assert.Contains(t, err.Error(), "500")
}
