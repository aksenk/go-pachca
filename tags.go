package pachca

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
)

// Tags
// Объект для работы с тегами
type Tags struct {
	client *resty.Client
}

// TagResponseRaw
// Объект для хранения ответа от API которые возвращают один тег
type TagResponseRaw struct {
	Data TagResponse `json:"data"`
}

// TagsResponseRaw
// Объект для хранения ответа от API которые возвращают список тегов
type TagsResponseRaw struct {
	Data []TagResponse `json:"data"`
}

// TagResponse
// Объект тега обогащенный информацией из API
type TagResponse struct {
	ID         int `json:"id"`
	UsersCount int `json:"users_count"`
	Tag
}

// Tag
// Объект тега
type Tag struct {
	Name string `json:"name"`
}

// ListTagsOptions
// Опции для запросов тегов с пагинацией
type ListTagsOptions struct {
	PaginationOptions
}

// New
// Метод для создания нового тега.
// Данный метод доступен для работы только с access_token администратора пространства
func (t *Tags) New(ctx context.Context, object *Tag) (*TagResponse, *resty.Response, error) {
	if object.Name == "" {
		return nil, nil, fmt.Errorf("%w: incorrect tag name %d", ErrInvalidInput, object.Name)
	}

	body := struct {
		GroupTag *Tag `json:"group_tag"`
	}{
		GroupTag: object,
	}

	resp, err := t.client.R().
		SetContext(ctx).
		SetBody(body).
		Post(tagsURL)
	if err != nil {
		return nil, resp, err
	}

	if resp.StatusCode() != 200 {
		return nil, resp, fmt.Errorf("%w: %d, body: %s", ErrResponseCode, resp.StatusCode())
	}

	var tag TagResponseRaw
	err = json.Unmarshal(resp.Body(), &tag)
	if err != nil {
		return nil, resp, fmt.Errorf("%w: %w", ErrResponseDecode, err)
	}

	return &tag.Data, resp, nil
}

// Edit
// Метод для редактирования тега.
// Данный метод доступен для работы только с access_token администратора пространства.
func (t *Tags) Edit(ctx context.Context, tagID int, object *Tag) (*TagResponse, *resty.Response, error) {
	if object.Name == "" {
		return nil, nil, fmt.Errorf("%w: incorrect tag name %d", ErrInvalidInput, object.Name)
	}
	if tagID <= 0 {
		return nil, nil, fmt.Errorf("%w: incorrect tag ID %d", ErrInvalidInput, tagID)
	}

	body := struct {
		GroupTag *Tag `json:"group_tag"`
	}{
		GroupTag: object,
	}

	resp, err := t.client.R().
		SetContext(ctx).
		SetBody(body).
		Put(tagsURL)
	if err != nil {
		return nil, resp, err
	}

	if resp.StatusCode() != 200 {
		return nil, resp, fmt.Errorf("%w: %d, body: %s", ErrResponseCode, resp.StatusCode())
	}

	var tag TagResponseRaw
	err = json.Unmarshal(resp.Body(), &tag)
	if err != nil {
		return nil, resp, fmt.Errorf("%w: %w", ErrResponseDecode, err)
	}

	return &tag.Data, resp, nil
}

// Delete
// Метод для удаления тега.
// Данный метод доступен для работы только с access_token администратора пространства.
func (t *Tags) Delete(ctx context.Context, tagID int) (*resty.Response, error) {
	if tagID <= 0 {
		return nil, fmt.Errorf("%w: incorrect tag ID %d", ErrInvalidInput, tagID)
	}

	resp, err := t.client.R().
		SetContext(ctx).
		Delete(tagsURL)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode() != 200 {
		return resp, fmt.Errorf("%w: %d, body: %s", ErrResponseCode, resp.StatusCode())
	}

	return resp, nil
}

// getTagsPaginated
// Запрос списка тегов с пагинацией
func (t *Tags) getTagsPaginated(ctx context.Context, options *ListTagsOptions) ([]TagResponse, *resty.Response, error) {
	if options.page <= 0 {
		return nil, nil, fmt.Errorf("%w: incorrect page %d", ErrInvalidInput, options.page)
	}
	if options.per <= 0 {
		return nil, nil, fmt.Errorf("%w: incorrect per %d", ErrInvalidInput, &options.per)

	}
	resp, err := t.client.R().
		SetQueryParams(map[string]string{
			"page": fmt.Sprint(options.page),
			"per":  fmt.Sprint(options.per),
		}).
		SetContext(ctx).
		Get(tagsURL)
	if err != nil {
		return nil, resp, err
	}
	if resp.StatusCode() != 200 {
		return nil, resp, fmt.Errorf("%w: %d, body: %s", ErrResponseCode, resp.StatusCode())
	}

	var tags TagsResponseRaw
	err = json.Unmarshal(resp.Body(), &tags)
	if err != nil {
		return nil, resp, fmt.Errorf("%w: %w", ErrResponseDecode, err)
	}

	return tags.Data, resp, nil
}

// Get
// Метод для получения информации о теге. Названия тегов являются уникальными в компании.
func (t *Tags) Get(ctx context.Context, tagID int) (TagResponse, *resty.Response, error) {
	if tagID <= 0 {
		return TagResponse{}, nil, fmt.Errorf("invalid tag ID: %d", tagID)
	}

	resp, err := t.client.R().
		SetContext(ctx).
		Get(fmt.Sprintf("%s/%d", tagsURL, tagID))
	if err != nil {
		return TagResponse{}, resp, err
	}

	if resp.StatusCode() != 200 {
		return TagResponse{}, resp, fmt.Errorf("%w: %d, body: %s", ErrResponseCode, resp.StatusCode())
	}

	var tag TagResponseRaw
	err = json.Unmarshal(resp.Body(), &tag)
	if err != nil {
		return TagResponse{}, resp, fmt.Errorf("%w: %w", ErrResponseDecode, err)
	}

	return tag.Data, resp, nil
}

// List
// Метод для получения актуального списка тегов сотрудников. Названия тегов являются уникальными в компании.
// Опции не обязательны
func (t *Tags) List(ctx context.Context, options *ListTagsOptions) (allTags []TagResponse, resp *resty.Response, err error) {
	if options == nil {
		options = &ListTagsOptions{
			PaginationOptions{
				page: 1,
				per:  50,
			},
		}
	}

	var tags []TagResponse

	for {
		tags, resp, err = t.getTagsPaginated(ctx, options)
		if err != nil {
			return nil, resp, err
		}
		if len(tags) == 0 {
			break
		}

		allTags = append(allTags, tags...)

		if len(tags) < options.per {
			break
		}
		options.page++
	}

	return allTags, resp, nil
}

// Find
// Метод для поиска тега по его имени
func (t *Tags) Find(ctx context.Context, name string) (tag *TagResponse, resp *resty.Response, err error) {
	options := &ListTagsOptions{
		PaginationOptions{
			page: 1,
			per:  50,
		},
	}

	var tags []TagResponse

	for {
		tags, resp, err = t.getTagsPaginated(ctx, options)
		if err != nil {
			return nil, resp, err
		}
		if len(tags) == 0 {
			break
		}

		for _, t := range tags {
			if t.Name == name {
				return &t, resp, nil
			}
		}

		if len(tags) < options.per {
			break
		}

		options.page++
	}

	return nil, resp, fmt.Errorf("tag with name %s not found", name)
}

type TagUsersOptions struct {
	PaginationOptions
}

// Users
// Метод для получения списка сотрудников принадлежащего этому тегу.
func (t *Tags) Users(ctx context.Context, tagID int) (allUsers []UserResponse, resp *resty.Response, err error) {
	if tagID <= 0 {
		return nil, nil, fmt.Errorf("invalid tag ID: %d", tagID)
	}

	options := &TagUsersOptions{
		PaginationOptions{
			page: 1,
			per:  50,
		},
	}

	var users []UserResponse

	for {
		users, resp, err = t.getTagUsersPaginated(ctx, tagID, options)
		if err != nil {
			return users, resp, err
		}
		if len(users) == 0 {
			break
		}

		allUsers = append(allUsers, users...)

		if len(users) < options.per {
			break
		}

		options.page++
	}

	return allUsers, resp, nil

}

// getTagUsersPaginated
// Получение списка пользователей для тега с пагинацией.
func (t *Tags) getTagUsersPaginated(ctx context.Context, tagID int, options *TagUsersOptions) ([]UserResponse, *resty.Response, error) {
	url := fmt.Sprintf("%s/%d/users", tagsURL, tagID)
	resp, err := t.client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"page": fmt.Sprint(options.page),
			"per":  fmt.Sprint(options.per),
		}).
		Get(url)
	if err != nil {
		return nil, resp, err
	}

	if resp.StatusCode() != 200 {
		body := resp.Body()
		return nil, resp, fmt.Errorf("%w: %d, body: %s", ErrResponseCode, resp.StatusCode(), string(body))
	}

	var users UsersResponseRaw
	err = json.Unmarshal(resp.Body(), &users)
	if err != nil {
		return nil, resp, fmt.Errorf("%w: %w", ErrResponseDecode, err)
	}

	return users.Data, resp, nil
}
