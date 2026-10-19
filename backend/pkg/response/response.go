package response

import (
	"github.com/gin-gonic/gin"

	"indoor-navigation/backend/pkg/errcode"
)

// Envelope 统一响应结构。
type Envelope struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
	RequestID string `json:"request_id"`
}

// OK 返回成功响应。
func OK(c *gin.Context, data any) {
	c.JSON(200, Envelope{
		Code:      int(errcode.CodeOK),
		Message:   errcode.Message(errcode.CodeOK),
		Data:      data,
		RequestID: requestID(c),
	})
}

// Fail 返回错误响应。
func Fail(c *gin.Context, code errcode.Code, message string, data any) {
	if message == "" {
		message = errcode.Message(code)
	}
	c.JSON(errcode.HTTPStatus(code), Envelope{
		Code:      int(code),
		Message:   message,
		Data:      data,
		RequestID: requestID(c),
	})
}

func requestID(c *gin.Context) string {
	if value := c.GetString("request_id"); value != "" {
		return value
	}
	return c.GetHeader("X-Request-ID")
}
