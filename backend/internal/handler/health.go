package handler

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"

	"indoor-navigation/backend/internal/dto"
	"indoor-navigation/backend/internal/graph"
)

// Pinger 抽象 MySQL Ping，便于接口测试注入桩实现。
type Pinger interface {
	PingContext(ctx context.Context) error
}

// HealthHandler 处理健康检查。
type HealthHandler struct {
	db      Pinger
	store   *graph.GraphStore
	version string
}

func NewHealthHandler(db Pinger, store *graph.GraphStore, version string) *HealthHandler {
	return &HealthHandler{db: db, store: store, version: version}
}

// Get 返回服务与依赖状态。
// @Summary 健康检查
// @Tags health
// @Produce json
// @Success 200 {object} dto.HealthResponse
// @Failure 503 {object} dto.HealthResponse
// @Router /health [get]
func (h *HealthHandler) Get(c *gin.Context) {
	checks := map[string]string{
		"mysql": "ok",
		"graph": "ok",
	}
	status := "ok"
	httpStatus := 200

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	if h.db == nil || h.db.PingContext(ctx) != nil {
		checks["mysql"] = "error"
		status = "degraded"
		httpStatus = 503
	}
	if h.store == nil || h.store.Stats().Nodes == 0 {
		checks["graph"] = "error"
		status = "degraded"
		httpStatus = 503
	}

	c.JSON(httpStatus, dto.HealthResponse{
		Status:  status,
		Service: "c5-indoor-navigation-backend",
		Version: h.version,
		Checks:  checks,
		Time:    time.Now().UTC().Format(time.RFC3339),
	})
}
