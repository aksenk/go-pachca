package pachca

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
)

// Files
// Объект для работы с файлами
type Files struct {
	client *resty.Client
}

// uploadParams
// Структура параметров для загрузки файла
type uploadParams struct {
	ContentDisposition string `json:"Content-Disposition"`
	ACL                string `json:"acl"`
	Policy             string `json:"policy"`
	XAmzCredential     string `json:"x-amz-credential"`
	XAmzAlgorithm      string `json:"x-amz-algorithm"`
	XAmzDate           string `json:"x-amz-date"`
	XAmzSignature      string `json:"x-amz-signature"`
	Key                string `json:"key"`
	DirectURL          string `json:"direct_url"`
}

// getUploadParams
// Метод для получения параметров загрузки файла в pachca
func (f *Files) getUploadParams(ctx context.Context) (*uploadParams, *resty.Response, error) {
	url := uploadsURL
	resp, err := f.client.R().
		SetContext(ctx).
		Post(url)
	if err != nil {
		return nil, resp, err
	}
	if resp.StatusCode() != 201 {
		return nil, resp, fmt.Errorf("%w: %d", ErrResponseCode, resp.StatusCode())
	}

	var r uploadParams
	err = json.Unmarshal(resp.Body(), &r)
	if err != nil {
		return nil, resp, fmt.Errorf("%w: %w", ErrResponseDecode, err)
	}

	return &r, resp, nil
}

// UploadFile
// Метод для загрузки файла в pachca
// Возвращает key загруженного файла, ответ сервера и ошибку
func (f *Files) UploadFile(ctx context.Context, fileName string, fileContent []byte) (string, *resty.Response, error) {
	params, resp, err := f.getUploadParams(ctx)
	if err != nil {
		return "", resp, err
	}

	reader := bytes.NewReader(fileContent)

	uploadResp, err := f.client.R().
		SetContext(ctx).
		SetFormData(map[string]string{
			"Content-Disposition": params.ContentDisposition,
			"acl":                 params.ACL,
			"policy":              params.Policy,
			"x-amz-credential":    params.XAmzCredential,
			"x-amz-algorithm":     params.XAmzAlgorithm,
			"x-amz-date":          params.XAmzDate,
			"x-amz-signature":     params.XAmzSignature,
			"key":                 params.Key,
		}).
		SetFileReader("file", fileName, reader).
		Post(params.DirectURL)

	if err != nil {
		return "", uploadResp, err
	}
	// Ответ 204 при успешной загрузке файла
	if uploadResp.StatusCode() != 204 {
		return "", uploadResp, fmt.Errorf("%w: %d", ErrResponseCode, uploadResp.StatusCode())
	}

	return strings.Replace(params.Key, "${filename}", fileName, 1), uploadResp, nil
}
