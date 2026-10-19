package handler

import (
	"github.com/gin-gonic/gin"

	"indoor-navigation/backend/internal/dto"
	"indoor-navigation/backend/internal/service"
	"indoor-navigation/backend/pkg/response"
)

// NearbyHandler 处理最近设施查询。
type NearbyHandler struct {
	service *service.NearbyService
}

func NewNearbyHandler(service *service.NearbyService) *NearbyHandler {
	return &NearbyHandler{service: service}
}

// Get 查询最近设施。
// @Summary 查询最近设施
// @Tags nearby
// @Produce json
// @Param building query string true "楼栋编码" example(C5)
// @Param floor query string true "楼层" Enums(F1,F2,F3,F4,F5)
// @Param room_no query string true "起点教室房号" example(101)
// @Param types query string false "设施类型，逗号分隔" example(restroom,elevator,stairs,exit,water)
// @Param allowed_floors query string false "允许楼层，逗号分隔" example(F1,F2)
// @Success 200 {object} dto.NearbyResponseEnvelope
// @Failure 400 {object} dto.ErrorEnvelope
// @Failure 404 {object} dto.ErrorEnvelope
// @Failure 500 {object} dto.ErrorEnvelope
// @Router /api/v1/nearby [get]
func (h *NearbyHandler) Get(c *gin.Context) {
	var query dto.NearbyQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		respondValidationError(c, err)
		return
	}

	result, err := h.service.Nearby(c.Request.Context(), query)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.OK(c, result)
}
