package dto

// ErrorDetail 单条参数校验错误。
type ErrorDetail struct {
	Field   string `json:"field" example:"from.room_no"`
	Message string `json:"message" example:"required"`
}

// ErrorData 错误响应附加数据。
type ErrorData struct {
	Details []ErrorDetail `json:"details,omitempty"`
}

// ErrorEnvelope Swagger 错误响应结构。
type ErrorEnvelope struct {
	Code      int        `json:"code" example:"1001"`
	Message   string     `json:"message" example:"参数校验失败"`
	Data      *ErrorData `json:"data,omitempty"`
	RequestID string     `json:"request_id" example:"8f3c2a7e"`
}
