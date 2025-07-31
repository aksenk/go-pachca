package pachca

type APIError struct {
	Key     string      `json:"key"`
	Value   string      `json:"value"`
	Message string      `json:"message"`
	Code    string      `json:"code"`
	Payload interface{} `json:"payload"`
}

type ErrorResponse struct {
	Errors []APIError `json:"errors"`
}
