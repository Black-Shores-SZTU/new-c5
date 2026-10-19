package handler

import (
	"github.com/gin-gonic/gin"

	"indoor-navigation/backend/internal/dto"
	"indoor-navigation/backend/internal/service"
	"indoor-navigation/backend/pkg/response"
)

// RouteHandler 处理路线规划接口。
type RouteHandler struct {
	service *service.RouteService
}

func NewRouteHandler(service *service.RouteService) *RouteHandler {
	return &RouteHandler{service: service}
}

// Create 规划两个教室之间的路线。
// @Summary 规划室内路线
// @Tags route
// @Accept json
// @Produce json
// @Param request body dto.RouteRequest true "路线请求"
// @Success 200 {object} dto.RouteResponseEnvelope
// @Failure 400 {object} dto.ErrorEnvelope
// @Failure 404 {object} dto.ErrorEnvelope
// @Failure 422 {object} dto.ErrorEnvelope
// @Failure 500 {object} dto.ErrorEnvelope
// @Router /api/v1/route [post]
func (h *RouteHandler) Create(c *gin.Context) {
	var request dto.RouteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondValidationError(c, err)
		return
	}

	result, err := h.service.Plan(c.Request.Context(), request)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.OK(c, result)
}
