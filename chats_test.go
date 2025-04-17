package pachca

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

// mockTransport используется для эмуляции HTTP-запросов и ответов
type mockTransport struct {
	response *http.Response
	err      error
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.response, m.err
}

// mockBody создает тело ��твета для тестов
func mockBody(content string) io.ReadCloser {
	return io.NopCloser(io.Reader(&mockReadCloser{data: []byte(content)}))
}

type mockReadCloser struct {
	data []byte
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	copy(p, m.data)
	return len(m.data), io.EOF
}

func (m *mockReadCloser) Close() error {
	return nil
}

func TestChats_New_Success(t *testing.T) {
	client := resty.New()
	client.SetTransport(&mockTransport{
		response: &http.Response{
			StatusCode: http.StatusCreated,
			Body:       mockBody(`{"data": {"id": 1, "name": "New Chat", "created_at": "2023-01-01T00:00:00Z"}}`),
		},
	})

	chats := &Chats{client: client}
	ctx := context.Background()

	chat := Chat{
		Name:      "New Chat",
		MemberIDs: []int{1, 2, 3},
		Public:    true,
	}

	resp, httpResp, err := chats.New(ctx, chat)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, resp.ID)
	assert.Equal(t, "New Chat", resp.Name)
}

func TestChats_New_EmptyName(t *testing.T) {
	client := resty.New()
	chats := &Chats{client: client}
	ctx := context.Background()

	chat := Chat{
		Name: "",
	}

	resp, httpResp, err := chats.New(ctx, chat)

	assert.Error(t, err)
	assert.Nil(t, httpResp)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "empty name")
}

func TestChats_New_ErrorResponse(t *testing.T) {
	client := resty.New()
	client.SetTransport(&mockTransport{
		response: &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       mockBody(`{"error": "bad request"}`),
		},
	})

	chats := &Chats{client: client}
	ctx := context.Background()

	chat := Chat{
		Name:      "New Chat",
		MemberIDs: []int{1, 2, 3},
	}

	resp, httpResp, err := chats.New(ctx, chat)

	assert.Error(t, err)
	assert.NotNil(t, httpResp)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "400")
}
