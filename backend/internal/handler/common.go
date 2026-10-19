package handler

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"indoor-navigation/backend/internal/dto"
	"indoor-navigation/backend/internal/route"
	"indoor-navigation/backend/pkg/errcode"
	"indoor-navigation/backend/pkg/response"
)

func respondValidationError(c *gin.Context, err error) {
	response.Fail(c, errcode.CodeInvalidParam, "参数校验失败", dto.ErrorData{
		Details: validationDetails(err),
	})
}

func validationDetails(err error) []dto.ErrorDetail {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		details := make([]dto.ErrorDetail, 0, len(validationErrors))
		for _, fieldErr := range validationErrors {
			details = append(details, dto.ErrorDetail{
				Field:   fieldErr.Namespace(),
				Message: fieldErr.Tag(),
			})
		}
		return details
	}
	return []dto.ErrorDetail{{Field: "request", Message: err.Error()}}
}

func handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, route.ErrInvalidRequest), errors.Is(err, route.ErrFloorUnavailable):
		response.Fail(c, errcode.CodeInvalidParam, err.Error(), nil)
	case errors.Is(err, route.ErrEndpointNotFound):
		response.Fail(c, errcode.CodeRoomNotFound, err.Error(), nil)
	case errors.Is(err, route.ErrNoRoute):
		response.Fail(c, errcode.CodeNoRoute, "未找到可行路线", nil)
	default:
		log.Printf("api error: %v", err)
		response.Fail(c, errcode.CodeInternal, "内部错误", nil)
	}
}
