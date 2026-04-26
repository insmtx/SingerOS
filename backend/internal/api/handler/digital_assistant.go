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
	r.POST("/GetDigitalAssistant", h.GetDigitalAssistant)
	r.POST("/UpdateDigitalAssistant", h.UpdateDigitalAssistant)
	r.POST("/DeleteDigitalAssistant", h.DeleteDigitalAssistant)
	r.POST("/ListDigitalAssistant", h.ListDigitalAssistant)
	r.POST("/UpdateDigitalAssistantConfig", h.UpdateDigitalAssistantConfig)
	r.POST("/UpdateDigitalAssistantStatus", h.UpdateDigitalAssistantStatus)
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
// @Router /CreateDigitalAssistant [post]
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

// GetDigitalAssistantRequest 获取数字助手请求
type GetDigitalAssistantRequest struct {
	ID    *uint   `json:"id,omitempty"`
	Code  *string `json:"code,omitempty"`
	OrgID uint    `json:"org_id" binding:"required"`
}

// GetDigitalAssistant 获取数字助手详情
// @Summary 获取数字助手详情
// @Description 根据ID或Code获取数字助手详情
// @Tags DigitalAssistant
// @Accept json
// @Produce json
// @Param body body GetDigitalAssistantRequest true "获取数字助手请求"
// @Success 200 {object} dto.CreateDigitalAssistantResponse "成功响应"
// @Failure 400 {object} dto.ErrorResponse "请求参数错误"
// @Failure 403 {object} dto.ErrorResponse "权限不足"
// @Failure 404 {object} dto.ErrorResponse "资源不存在"
// @Failure 500 {object} dto.ErrorResponse "内部服务器错误"
// @Router /GetDigitalAssistant [post]
func (h *DigitalAssistantHandler) GetDigitalAssistant(ctx *gin.Context) {
	var req GetDigitalAssistantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	if req.ID == nil && req.Code == nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, "id or code is required"))
		return
	}

	if req.OrgID == 0 {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, "org_id is required"))
		return
	}

	var result *contract.DigitalAssistantDetail
	var err error

	if req.ID != nil {
		result, err = h.service.GetDigitalAssistantByID(ctx, req.OrgID, *req.ID)
	} else {
		result, err = h.service.GetDigitalAssistantByCode(ctx, req.OrgID, *req.Code)
	}

	if err != nil {
		if err.Error() == "permission denied: digital assistant belongs to different organization" {
			ctx.JSON(http.StatusForbidden, dto.Error(dto.CodeInternalError, err.Error()))
			return
		}
		if err.Error() == "digital assistant not found" {
			ctx.JSON(http.StatusNotFound, dto.Error(dto.CodeNotFound, err.Error()))
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, dto.Success(result))
}

// UpdateDigitalAssistantRequest 更新数字助手请求
type UpdateDigitalAssistantRequest struct {
	ID uint `json:"id" binding:"required"`
	contract.UpdateDigitalAssistantRequest
}

// UpdateDigitalAssistant 更新数字助手
// @Summary 更新数字助手
// @Description 更新数字助手基本信息
// @Tags DigitalAssistant
// @Accept json
// @Produce json
// @Param body body UpdateDigitalAssistantRequest true "更新数字助手请求"
// @Success 200 {object} dto.CreateDigitalAssistantResponse "成功响应"
// @Failure 400 {object} dto.ErrorResponse "请求参数错误"
// @Failure 500 {object} dto.ErrorResponse "内部服务器错误"
// @Router /UpdateDigitalAssistant [post]
func (h *DigitalAssistantHandler) UpdateDigitalAssistant(ctx *gin.Context) {
	var req UpdateDigitalAssistantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	result, err := h.service.UpdateDigitalAssistant(ctx, req.ID, &req.UpdateDigitalAssistantRequest)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, dto.Success(result))
}

// DeleteDigitalAssistantRequest 删除数字助手请求
type DeleteDigitalAssistantRequest struct {
	ID uint `json:"id" binding:"required"`
}

// DeleteDigitalAssistant 删除数字助手
// @Summary 删除数字助手
// @Description 根据ID删除数字助手
// @Tags DigitalAssistant
// @Accept json
// @Produce json
// @Param body body DeleteDigitalAssistantRequest true "删除数字助手请求"
// @Success 200 {object} dto.BaseResponse "成功响应"
// @Failure 400 {object} dto.ErrorResponse "请求参数错误"
// @Failure 500 {object} dto.ErrorResponse "内部服务器错误"
// @Router /DeleteDigitalAssistant [post]
func (h *DigitalAssistantHandler) DeleteDigitalAssistant(ctx *gin.Context) {
	var req DeleteDigitalAssistantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	err := h.service.DeleteDigitalAssistant(ctx, req.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, dto.Success(nil))
}

// ListDigitalAssistant 查询数字助手列表
// @Summary 查询数字助手列表
// @Description 分页查询数字助手列表
// @Tags DigitalAssistant
// @Accept json
// @Produce json
// @Param body body contract.ListDigitalAssistantRequest true "查询列表请求"
// @Success 200 {object} dto.CreateDigitalAssistantResponse "成功响应"
// @Failure 400 {object} dto.ErrorResponse "请求参数错误"
// @Failure 500 {object} dto.ErrorResponse "内部服务器错误"
// @Router /ListDigitalAssistant [post]
func (h *DigitalAssistantHandler) ListDigitalAssistant(ctx *gin.Context) {
	var req contract.ListDigitalAssistantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	result, err := h.service.ListDigitalAssistant(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, dto.Success(result))
}

// UpdateDigitalAssistantConfigRequest 更新数字助手配置请求
type UpdateDigitalAssistantConfigRequest struct {
	ID uint `json:"id" binding:"required"`
	contract.UpdateDigitalAssistantConfigRequest
}

// UpdateDigitalAssistantConfig 更新数字助手配置
// @Summary 更新数字助手配置
// @Description 更新数字助手的配置信息
// @Tags DigitalAssistant
// @Accept json
// @Produce json
// @Param body body UpdateDigitalAssistantConfigRequest true "更新配置请求"
// @Success 200 {object} dto.CreateDigitalAssistantResponse "成功响应"
// @Failure 400 {object} dto.ErrorResponse "请求参数错误"
// @Failure 500 {object} dto.ErrorResponse "内部服务器错误"
// @Router /UpdateDigitalAssistantConfig [post]
func (h *DigitalAssistantHandler) UpdateDigitalAssistantConfig(ctx *gin.Context) {
	var req UpdateDigitalAssistantConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	result, err := h.service.UpdateDigitalAssistantConfig(ctx, req.ID, &req.UpdateDigitalAssistantConfigRequest)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, dto.Success(result))
}

// UpdateDigitalAssistantStatusRequest 更新数字助手状态请求
type UpdateDigitalAssistantStatusRequest struct {
	ID uint `json:"id" binding:"required"`
	contract.UpdateDigitalAssistantStatusRequest
}

// UpdateDigitalAssistantStatus 更新数字助手状态
// @Summary 更新数字助手状态
// @Description 更新数字助手的运行状态
// @Tags DigitalAssistant
// @Accept json
// @Produce json
// @Param body body UpdateDigitalAssistantStatusRequest true "更新状态请求"
// @Success 200 {object} dto.BaseResponse "成功响应"
// @Failure 400 {object} dto.ErrorResponse "请求参数错误"
// @Failure 500 {object} dto.ErrorResponse "内部服务器错误"
// @Router /UpdateDigitalAssistantStatus [post]
func (h *DigitalAssistantHandler) UpdateDigitalAssistantStatus(ctx *gin.Context) {
	var req UpdateDigitalAssistantStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
		return
	}

	err := h.service.UpdateDigitalAssistantStatus(ctx, req.ID, &req.UpdateDigitalAssistantStatusRequest)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, dto.Success(nil))
}
