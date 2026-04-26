package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/insmtx/SingerOS/backend/internal/api/contract"
	"github.com/insmtx/SingerOS/backend/internal/api/dto"
)

// DigitalAssistantHandler DigitalAssistant RPC风格Handler
type DigitalAssistantHandler struct {
	service contract.DigitalAssistantService
}

// NewDigitalAssistantHandler 创建Handler实例
func NewDigitalAssistantHandler(service contract.DigitalAssistantService) *DigitalAssistantHandler {
	return &DigitalAssistantHandler{
		service: service,
	}
}

// RegisterRoutes 注册RPC风格路由
func (h *DigitalAssistantHandler) RegisterRoutes(r gin.IRouter) {
	r.POST("/CreateDigitalAssistant", h.CreateDigitalAssistant)
}

// RegisterDigitalAssistantRoutes 注册DigitalAssistant路由(便捷函数)
func RegisterDigitalAssistantRoutes(r gin.IRouter, service contract.DigitalAssistantService) {
	h := NewDigitalAssistantHandler(service)
	h.RegisterRoutes(r)
}

// CreateDigitalAssistant 创建数字助手
// @Summary 创建数字助手
// @Description 创建一个新的数字助手实例
// @Tags DigitalAssistant
// @Accept json
// @Produce json
// @Param body body contract.CreateDigitalAssistantRequest true "创建数字助手请求"
// @Success 200 {object} dto.CreateDigitalAssistantResponse "成功响应"
// @Failure 400 {object} dto.ErrorResponse "请求参数错误"
// @Failure 500 {object} dto.ErrorResponse "内部服务器错误"
// @Router /v1/CreateDigitalAssistant [post]
func (h *DigitalAssistantHandler) CreateDigitalAssistant(ctx *gin.Context) {
	var req contract.CreateDigitalAssistantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	result, err := h.service.CreateDigitalAssistant(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, dto.NewCreateDigitalAssistantResponse(result))
}
