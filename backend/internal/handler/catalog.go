package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"indoor-navigation/backend/internal/dto"
	"indoor-navigation/backend/internal/service"
	"indoor-navigation/backend/pkg/errcode"
	"indoor-navigation/backend/pkg/response"
)

// CatalogHandler 处理楼栋、楼层与教室目录查询。
type CatalogHandler struct {
	service *service.CatalogService
}

func NewCatalogHandler(service *service.CatalogService) *CatalogHandler {
	return &CatalogHandler{service: service}
}

// Buildings 返回楼栋列表。
// @Summary 楼栋列表
// @Tags catalog
// @Produce json
// @Success 200 {object} dto.BuildingsResponseEnvelope
// @Failure 500 {object} dto.ErrorEnvelope
// @Router /buildings [get]
func (h *CatalogHandler) Buildings(c *gin.Context) {
	response.OK(c, dto.BuildingsResponse{Items: h.service.Buildings()})
}

// Floors 返回指定楼栋的楼层列表。
// @Summary 楼层列表
// @Tags catalog
// @Produce json
// @Param code path string true "楼栋编码" example(C5)
// @Success 200 {object} dto.FloorsResponseEnvelope
// @Failure 404 {object} dto.ErrorEnvelope
// @Failure 500 {object} dto.ErrorEnvelope
// @Router /buildings/{code}/floors [get]
func (h *CatalogHandler) Floors(c *gin.Context) {
	code := strings.ToUpper(strings.TrimSpace(c.Param("code")))
	if code == "" {
		response.Fail(c, errcode.CodeInvalidParam, "楼栋编码不能为空", nil)
		return
	}
	items, ok := h.service.Floors(code)
	if !ok {
		response.Fail(c, errcode.CodeCatalogNotFound, "楼栋不存在", nil)
		return
	}
	response.OK(c, dto.FloorsResponse{Building: code, Items: items})
}

// Classrooms 按关键词与楼层搜索教室。
// @Summary 教室搜索
// @Tags catalog
// @Produce json
// @Param building query string false "楼栋编码" example(C5)
// @Param floor query string false "楼层" Enums(F1,F2,F3,F4,F5)
// @Param keyword query string false "房号或名称关键词" example(10)
// @Param limit query int false "每页数量，默认 20，最大 100" example(20)
// @Param offset query int false "偏移量" example(0)
// @Success 200 {object} dto.ClassroomsResponseEnvelope
// @Failure 400 {object} dto.ErrorEnvelope
// @Failure 500 {object} dto.ErrorEnvelope
// @Router /classrooms [get]
func (h *CatalogHandler) Classrooms(c *gin.Context) {
	var query dto.ClassroomQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		respondValidationError(c, err)
		return
	}

	total, items := h.service.Classrooms(query.Building, query.Floor, query.Keyword, query.Limit, query.Offset)
	response.OK(c, dto.ClassroomsResponse{Total: total, Items: items})
}
