package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// AccessLog 记录结构化访问日志；request_id 由 RequestID 中间件提前写入。
func AccessLog(logger *slog.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = slog.Default()
	}
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		status := c.Writer.Status()
		attrs := []any{
			"request_id", GetRequestID(c),
			"method", c.Request.Method,
			"path", path,
			"query", query,
			"status", status,
			"latency_ms", float64(time.Since(start).Microseconds()) / 1000.0,
			"client_ip", c.ClientIP(),
			"bytes", c.Writer.Size(),
		}
		if status >= 500 {
			logger.Error("http_request", attrs...)
			return
		}
		logger.Info("http_request", attrs...)
	}
}
