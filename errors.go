package pachca

type ErrorResponse struct {
	Key     string      `json:"key"`
	Value   string      `json:"value"`
	Message string      `json:"message"`
	Code    string      `json:"code"`
	Payload interface{} `json:"payload"`
}
