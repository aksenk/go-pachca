package pachca

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

// sequenceTransport эмулирует последовательность HTTP-ответов и запоминает запросы,
// чтобы можно было проверять параметры пагинации между вызовами
type sequenceTransport struct {
	responses []*http.Response
	requests  []*http.Request
	idx       int
}

func (s *sequenceTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	s.requests = append(s.requests, req)
	if s.idx >= len(s.responses) {
		return nil, fmt.Errorf("unexpected request #%d", s.idx+1)
	}
	resp := s.responses[s.idx]
	s.idx++
	return resp, nil
}

func usersPageBody(nextPage string, ids ...int) string {
	data := ""
	for i, id := range ids {
		if i > 0 {
			data += ","
		}
		data += fmt.Sprintf(`{"id": %d, "first_name": "User%d", "nickname": "user%d"}`, id, id, id)
	}
	return fmt.Sprintf(`{"data": [%s], "meta": {"paginate": {"next_page": "%s"}}}`, data, nextPage)
}

func newUsersWithTransport(transport http.RoundTripper) *Users {
	client := resty.New()
	client.SetTransport(transport)
	return &Users{client: client}
}

func TestUsers_ListV2_Success(t *testing.T) {
	transport := &sequenceTransport{
		responses: []*http.Response{
			{
				StatusCode: http.StatusOK,
				Body:       mockBody(usersPageBody("cursor123", 1, 2)),
			},
		},
	}
	users := newUsersWithTransport(transport)

	options := &ListUsersOptionsV2{
		Query: "ivan",
		PaginationOptions: PaginationOptionsUsers{
			Limit: 25,
			Next:  "prev_cursor",
		},
	}

	result, next, httpResp, err := users.ListV2(context.Background(), options)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Len(t, result, 2)
	assert.Equal(t, 1, result[0].ID)
	assert.Equal(t, "User1", result[0].FirstName)
	assert.Equal(t, 2, result[1].ID)
	assert.Equal(t, "cursor123", next)

	// проверяем, что опции корректно превратились в query-параметры
	assert.Len(t, transport.requests, 1)
	query := transport.requests[0].URL.Query()
	assert.Equal(t, "25", query.Get("limit"))
	assert.Equal(t, "prev_cursor", query.Get("cursor"))
	assert.Equal(t, "ivan", query.Get("query"))
}

func TestUsers_ListV2_NilOptions(t *testing.T) {
	transport := &sequenceTransport{
		responses: []*http.Response{
			{
				StatusCode: http.StatusOK,
				Body:       mockBody(usersPageBody("", 1)),
			},
		},
	}
	users := newUsersWithTransport(transport)

	result, next, httpResp, err := users.ListV2(context.Background(), nil)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Len(t, result, 1)
	assert.Equal(t, "", next)

	// при nil-опциях должен применяться лимит по умолчанию
	query := transport.requests[0].URL.Query()
	assert.Equal(t, "50", query.Get("limit"))
	assert.Equal(t, "", query.Get("cursor"))
	assert.Equal(t, "", query.Get("query"))
}

func TestUsers_ListV2_EmptyData(t *testing.T) {
	transport := &sequenceTransport{
		responses: []*http.Response{
			{
				StatusCode: http.StatusOK,
				Body:       mockBody(usersPageBody("dangling_cursor")),
			},
		},
	}
	users := newUsersWithTransport(transport)

	result, next, httpResp, err := users.ListV2(context.Background(), nil)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Empty(t, result)
	// при пустой странице курсор не возвращается, даже если API его прислал
	assert.Equal(t, "", next)
}

func TestUsers_ListV2_ErrorResponse(t *testing.T) {
	transport := &sequenceTransport{
		responses: []*http.Response{
			{
				StatusCode: http.StatusBadRequest,
				Body:       mockBody(`{"error": "bad request"}`),
			},
		},
	}
	users := newUsersWithTransport(transport)

	result, next, httpResp, err := users.ListV2(context.Background(), nil)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrResponseCode)
	assert.Contains(t, err.Error(), "400")
	assert.NotNil(t, httpResp)
	assert.Nil(t, result)
	assert.Equal(t, "", next)
}

func TestUsers_ListV2_DecodeError(t *testing.T) {
	transport := &sequenceTransport{
		responses: []*http.Response{
			{
				StatusCode: http.StatusOK,
				Body:       mockBody(`not a json`),
			},
		},
	}
	users := newUsersWithTransport(transport)

	result, next, httpResp, err := users.ListV2(context.Background(), nil)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrResponseDecode)
	assert.NotNil(t, httpResp)
	assert.Nil(t, result)
	assert.Equal(t, "", next)
}

func TestUsers_Find_SinglePage(t *testing.T) {
	transport := &sequenceTransport{
		responses: []*http.Response{
			{
				StatusCode: http.StatusOK,
				Body:       mockBody(usersPageBody("cursor_p2", 1, 2)),
			},
			{
				StatusCode: http.StatusOK,
				Body:       mockBody(usersPageBody("")),
			},
		},
	}
	users := newUsersWithTransport(transport)

	result, httpResp, err := users.Find(context.Background(), "ivan")

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Len(t, result, 2)
	assert.Equal(t, 1, result[0].ID)
	assert.Equal(t, 2, result[1].ID)

	// поисковая фраза должна передаваться в каждом запросе,
	// а курсор второго запроса — браться из next_page первого ответа
	assert.Len(t, transport.requests, 2)
	firstQuery := transport.requests[0].URL.Query()
	assert.Equal(t, "ivan", firstQuery.Get("query"))
	assert.Equal(t, "", firstQuery.Get("cursor"))
	assert.Equal(t, "50", firstQuery.Get("limit"))

	secondQuery := transport.requests[1].URL.Query()
	assert.Equal(t, "ivan", secondQuery.Get("query"))
	assert.Equal(t, "cursor_p2", secondQuery.Get("cursor"))
}

func TestUsers_Find_MultiPage(t *testing.T) {
	transport := &sequenceTransport{
		responses: []*http.Response{
			{
				StatusCode: http.StatusOK,
				Body:       mockBody(usersPageBody("cursor_p2", 1, 2)),
			},
			{
				StatusCode: http.StatusOK,
				Body:       mockBody(usersPageBody("cursor_p3", 3, 4)),
			},
			{
				StatusCode: http.StatusOK,
				Body:       mockBody(usersPageBody("")),
			},
		},
	}
	users := newUsersWithTransport(transport)

	result, httpResp, err := users.Find(context.Background(), "petrov")

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Len(t, result, 4)
	assert.Equal(t, []int{1, 2, 3, 4}, []int{result[0].ID, result[1].ID, result[2].ID, result[3].ID})

	// курсоры должны передаваться по цепочке: "" -> cursor_p2 -> cursor_p3
	assert.Len(t, transport.requests, 3)
	assert.Equal(t, "", transport.requests[0].URL.Query().Get("cursor"))
	assert.Equal(t, "cursor_p2", transport.requests[1].URL.Query().Get("cursor"))
	assert.Equal(t, "cursor_p3", transport.requests[2].URL.Query().Get("cursor"))
}

func TestUsers_Find_EmptyResult(t *testing.T) {
	transport := &sequenceTransport{
		responses: []*http.Response{
			{
				StatusCode: http.StatusOK,
				Body:       mockBody(usersPageBody("")),
			},
		},
	}
	users := newUsersWithTransport(transport)

	result, httpResp, err := users.Find(context.Background(), "nobody")

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Empty(t, result)
	assert.Len(t, transport.requests, 1)
}

func TestUsers_Find_StopsOnEmptyNextCursor(t *testing.T) {
	// последняя непустая страница с пустым next_page должна завершать пагинацию
	// без дополнительного запроса (защита от бесконечного цикла)
	transport := &sequenceTransport{
		responses: []*http.Response{
			{
				StatusCode: http.StatusOK,
				Body:       mockBody(usersPageBody("cursor_p2", 1, 2)),
			},
			{
				StatusCode: http.StatusOK,
				Body:       mockBody(usersPageBody("", 3)),
			},
		},
	}
	users := newUsersWithTransport(transport)

	result, httpResp, err := users.Find(context.Background(), "ivan")

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Len(t, result, 3)
	assert.Equal(t, 3, result[2].ID)
	assert.Len(t, transport.requests, 2)
}

func TestUsers_Find_ErrorResponse(t *testing.T) {
	transport := &sequenceTransport{
		responses: []*http.Response{
			{
				StatusCode: http.StatusUnauthorized,
				Body:       mockBody(`{"error": "unauthorized"}`),
			},
		},
	}
	users := newUsersWithTransport(transport)

	result, httpResp, err := users.Find(context.Background(), "ivan")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrResponseCode)
	assert.Contains(t, err.Error(), "401")
	assert.NotNil(t, httpResp)
	assert.Nil(t, result)
}

func TestUsers_Find_ErrorOnSecondPage(t *testing.T) {
	transport := &sequenceTransport{
		responses: []*http.Response{
			{
				StatusCode: http.StatusOK,
				Body:       mockBody(usersPageBody("cursor_p2", 1, 2)),
			},
			{
				StatusCode: http.StatusInternalServerError,
				Body:       mockBody(`{"error": "internal"}`),
			},
		},
	}
	users := newUsersWithTransport(transport)

	result, _, err := users.Find(context.Background(), "ivan")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrResponseCode)
	assert.Contains(t, err.Error(), "500")
	// при ошибке на любой из страниц результат отбрасывается целиком
	assert.Nil(t, result)
}
