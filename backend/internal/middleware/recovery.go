package middleware

import (
	"fmt"
	"log/slog"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"indoor-navigation/backend/pkg/errcode"
	"indoor-navigation/backend/pkg/response"
)

// Recovery 捕获 panic，把堆栈写入日志并返回统一 500 响应。
func Recovery(logger *slog.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = slog.Default()
	}
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("panic_recovered",
					"request_id", GetRequestID(c),
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"panic", fmt.Sprint(recovered),
					"stack", string(debug.Stack()),
				)
				if !c.Writer.Written() {
					response.Fail(c, errcode.CodeInternal, "内部错误", nil)
				}
				c.Abort()
			}
		}()
		c.Next()
	}
}
