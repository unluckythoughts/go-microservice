package web

type HTTPResponse struct {
	Success   bool        `json:"success"`
	RequestID string      `json:"requestId"`
	Error     string      `json:"error,omitempty"`
	Result    interface{} `json:"result,omitempty"`
}
