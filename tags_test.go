package pachca

import (
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Мокируем клиента Resty для симуляции ответов API
type MockRestyClient struct {
	mock.Mock
}

func (m *MockRestyClient) R() *resty.Request {
	args := m.Called()
	return args.Get(0).(*resty.Request)
}

func (m *MockRestyClient) SetQueryParams(params map[string]string) *resty.Request {
	args := m.Called(params)
	return args.Get(0).(*resty.Request)
}

func (m *MockRestyClient) Get(url string) (*resty.Response, error) {
	args := m.Called(url)
	return args.Get(0).(*resty.Response), args.Error(1)
}

// Тест для метода getTagsPaginated
func TestGetTagsPaginated(t *testing.T) {
	client := new(MockRestyClient)
	tags := &Tags{client: client}

	// Подготовка мока
	client.On("R").Return(&resty.Request{})
	client.On("SetQueryParams", mock.Anything).Return(&resty.Request{})
	client.On("Get", "tags_url?page=1&per=10").Return(&resty.Response{
		StatusCode: 200,
		Body:       []byte(`{"data":[{"id":1,"name":"tag1","users_count":10}]}`),
	}, nil)

	// Тестируем успешный вызов
	result, err := tags.getTagsPaginated(1, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, result[0].Name, "tag1")

	// Тестируем ошибку при неверной странице
	_, err = tags.getTagsPaginated(-1, 10)
	assert.Error(t, err)
}

// Тест для метода Get
func TestGet(t *testing.T) {
	client := new(MockRestyClient)
	tags := &Tags{client: client}

	// Подготовка мока
	client.On("R").Return(&resty.Request{})
	client.On("Get", "tags_url/1").Return(&resty.Response{
		StatusCode: 200,
		Body:       []byte(`{"data":{"id":1,"name":"tag1","users_count":10}}`),
	}, nil)

	// Тестируем успешный вызов
	tag, resp, err := tags.Get(1)
	assert.NoError(t, err)
	assert.Equal(t, tag.Name, "tag1")
	assert.Equal(t, resp.StatusCode(), 200)

	// Тестируем ошибку при неверном ID
	_, _, err = tags.Get(0)
	assert.Error(t, err)
}

// Тест для метода List
func TestGetAll(t *testing.T) {
	client := new(MockRestyClient)
	tags := &Tags{client: client}

	// Подготовка мока
	client.On("R").Return(&resty.Request{})
	client.On("SetQueryParams", mock.Anything).Return(&resty.Request{})
	client.On("Get", "tags_url?page=1&per=50").Return(&resty.Response{
		StatusCode: 200,
		Body:       []byte(`{"data":[{"id":1,"name":"tag1","users_count":10}]}`),
	}, nil)
	client.On("Get", "tags_url?page=2&per=50").Return(&resty.Response{
		StatusCode: 200,
		Body:       []byte(`{"data":[]}`),
	}, nil)

	// Тестируем успешный вызов
	result, err := tags.List()
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, result[0].Name, "tag1")
}

// Тест для метода Find
func TestFind(t *testing.T) {
	client := new(MockRestyClient)
	tags := &Tags{client: client}

	// Подготовка мока
	client.On("R").Return(&resty.Request{})
	client.On("SetQueryParams", mock.Anything).Return(&resty.Request{})
	client.On("Get", "tags_url?page=1&per=50").Return(&resty.Response{
		StatusCode: 200,
		Body:       []byte(`{"data":[{"id":1,"name":"tag1","users_count":10}]}`),
	}, nil)

	// Тестируем успешный поиск
	tag, err := tags.Find("tag1")
	assert.NoError(t, err)
	assert.Equal(t, tag.Name, "tag1")

	// Тестируем ошибку при отсутствии тега
	_, err = tags.Find("nonexistent")
	assert.Error(t, err)
}

// Тест для метода Users
func TestUsers(t *testing.T) {
	client := new(MockRestyClient)
	tags := &Tags{client: client}

	// Подготовка мока
	client.On("R").Return(&resty.Request{})
	client.On("SetQueryParams", mock.Anything).Return(&resty.Request{})
	client.On("Get", "tags_url/1/users?page=1&perPage=50").Return(&resty.Response{
		StatusCode: 200,
		Body:       []byte(`{"data":[{"id":1,"name":"user1"}]}`),
	}, nil)

	// Тестируем успешный вызов
	users, err := tags.Users(1)
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, users[0].Name, "user1")

	// Тестируем ошибку при неверном ID
	_, err = tags.Users(0)
	assert.Error(t, err)
}
