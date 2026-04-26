package service

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/insmtx/SingerOS/backend/internal/api/contract"
	"github.com/insmtx/SingerOS/backend/internal/infra/db"
	"github.com/insmtx/SingerOS/backend/types"
)

// digitalAssistantService DigitalAssistant服务实现（未导出）
type digitalAssistantService struct {
	db *gorm.DB
}

// NewDigitalAssistantService 创建DigitalAssistant服务实例
func NewDigitalAssistantService(db *gorm.DB) contract.DigitalAssistantService {
	return &digitalAssistantService{
		db: db,
	}
}

// CreateDigitalAssistant 创建数字助手
func (s *digitalAssistantService) CreateDigitalAssistant(ctx context.Context, req *contract.CreateDigitalAssistantRequest) (*contract.DigitalAssistant, error) {
	// 验证必填字段
	if req.Code == "" {
		return nil, errors.New("code is required")
	}
	if req.Name == "" {
		return nil, errors.New("name is required")
	}

	// 检查code是否已存在
	exists, err := db.DigitalAssistantCodeExists(ctx, s.db, req.Code, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("digital assistant with this code already exists")
	}

	// 构建数据库实体
	now := time.Now()
	da := &types.DigitalAssistant{
		Code:        req.Code,
		OrgID:       req.OrgID,
		OwnerID:     req.OwnerID,
		Name:        req.Name,
		Description: req.Description,
		Avatar:      req.Avatar,
		Status:      string(contract.DigitalAssistantStatusDraft),
		Version:     0,
		Config: types.AssistantConfig{
			Runtime:   types.RuntimeConfig{Type: req.Config.Runtime.Type},
			LLM:       types.LLMConfig{Type: req.Config.LLM.Type},
			Skills:    convertSkillRefs(req.Config.Skills),
			Channels:  convertChannelRefs(req.Config.Channels),
			Knowledge: convertKnowledgeRefs(req.Config.Knowledge),
			Memory:    types.MemoryConfig{Type: req.Config.Memory.Type},
			Policies:  types.PolicyConfig{Type: req.Config.Policies.Type},
		},
	}

	// 设置时间戳
	da.CreatedAt = now
	da.UpdatedAt = now

	// 保存到数据库
	err = db.CreateDigitalAssistant(ctx, s.db, da)
	if err != nil {
		return nil, err
	}

	// 转换为响应结构
	return convertToContractDigitalAssistant(da), nil
}

// GetDigitalAssistantByID 根据ID获取数字助手详情（需验证同组织权限）
func (s *digitalAssistantService) GetDigitalAssistantByID(ctx context.Context, orgID uint, id uint) (*contract.DigitalAssistantDetail, error) {
	da, err := db.GetDigitalAssistantByID(ctx, s.db, id)
	if err != nil {
		return nil, err
	}
	if da == nil {
		return nil, errors.New("digital assistant not found")
	}

	// 验证组织权限
	if da.OrgID != orgID {
		return nil, errors.New("permission denied: digital assistant belongs to different organization")
	}

	return &contract.DigitalAssistantDetail{
		DigitalAssistant: *convertToContractDigitalAssistant(da),
	}, nil
}

// GetDigitalAssistantByCode 根据Code获取数字助手详情（需验证同组织权限）
func (s *digitalAssistantService) GetDigitalAssistantByCode(ctx context.Context, orgID uint, code string) (*contract.DigitalAssistantDetail, error) {
	da, err := db.GetDigitalAssistantByCode(ctx, s.db, code)
	if err != nil {
		return nil, err
	}
	if da == nil {
		return nil, errors.New("digital assistant not found")
	}

	// 验证组织权限
	if da.OrgID != orgID {
		return nil, errors.New("permission denied: digital assistant belongs to different organization")
	}

	return &contract.DigitalAssistantDetail{
		DigitalAssistant: *convertToContractDigitalAssistant(da),
	}, nil
}

// UpdateDigitalAssistant 更新数字助手
func (s *digitalAssistantService) UpdateDigitalAssistant(ctx context.Context, id uint, req *contract.UpdateDigitalAssistantRequest) (*contract.DigitalAssistant, error) {
	// TODO: 实现更新逻辑
	return nil, nil
}

// DeleteDigitalAssistant 删除数字助手
func (s *digitalAssistantService) DeleteDigitalAssistant(ctx context.Context, id uint) error {
	// TODO: 实现删除逻辑
	return nil
}

// ListDigitalAssistant 查询数字助手列表
func (s *digitalAssistantService) ListDigitalAssistant(ctx context.Context, req *contract.ListDigitalAssistantRequest) (*contract.DigitalAssistantList, error) {
	// TODO: 实现列表查询逻辑
	return nil, nil
}

// UpdateDigitalAssistantConfig 更新数字助手配置
func (s *digitalAssistantService) UpdateDigitalAssistantConfig(ctx context.Context, id uint, req *contract.UpdateDigitalAssistantConfigRequest) (*contract.DigitalAssistant, error) {
	// TODO: 实现配置更新逻辑
	return nil, nil
}

// UpdateDigitalAssistantStatus 更新数字助手状态
func (s *digitalAssistantService) UpdateDigitalAssistantStatus(ctx context.Context, id uint, req *contract.UpdateDigitalAssistantStatusRequest) error {
	// TODO: 实现状态更新逻辑
	return nil
}

// 确保 digitalAssistantService 实现了 contract.DigitalAssistantService 接口
var _ contract.DigitalAssistantService = (*digitalAssistantService)(nil)

// convertSkillRefs 转换技能引用
func convertSkillRefs(reqRefs []contract.SkillRef) []types.SkillRef {
	result := make([]types.SkillRef, 0, len(reqRefs))
	for _, ref := range reqRefs {
		result = append(result, types.SkillRef{
			SkillCode: ref.SkillCode,
			Version:   ref.Version,
		})
	}
	return result
}

// convertChannelRefs 转换渠道引用
func convertChannelRefs(reqRefs []contract.ChannelRef) []types.ChannelRef {
	result := make([]types.ChannelRef, 0, len(reqRefs))
	for _, ref := range reqRefs {
		result = append(result, types.ChannelRef{
			Type: ref.Type,
		})
	}
	return result
}

// convertKnowledgeRefs 转换知识库引用
func convertKnowledgeRefs(reqRefs []contract.KnowledgeRef) []types.KnowledgeRef {
	result := make([]types.KnowledgeRef, 0, len(reqRefs))
	for _, ref := range reqRefs {
		result = append(result, types.KnowledgeRef{
			Type:      ref.Type,
			DatasetID: ref.DatasetID,
			Repo:      ref.Repo,
		})
	}
	return result
}

// convertToContractDigitalAssistant 转换为合约层DigitalAssistant
func convertToContractDigitalAssistant(da *types.DigitalAssistant) *contract.DigitalAssistant {
	return &contract.DigitalAssistant{
		ID:          da.ID,
		Code:        da.Code,
		OrgID:       da.OrgID,
		OwnerID:     da.OwnerID,
		Name:        da.Name,
		Description: da.Description,
		Avatar:      da.Avatar,
		Status:      da.Status,
		Version:     da.Version,
		Config: contract.AssistantConfig{
			Runtime:   contract.RuntimeConfig{Type: da.Config.Runtime.Type},
			LLM:       contract.LLMConfig{Type: da.Config.LLM.Type},
			Skills:    convertSkillRefsToContract(da.Config.Skills),
			Channels:  convertChannelRefsToContract(da.Config.Channels),
			Knowledge: convertKnowledgeRefsToContract(da.Config.Knowledge),
			Memory:    contract.MemoryConfig{Type: da.Config.Memory.Type},
			Policies:  contract.PolicyConfig{Type: da.Config.Policies.Type},
		},
		CreatedAt: da.CreatedAt,
		UpdatedAt: da.UpdatedAt,
	}
}

// convertSkillRefsToContract 转换技能引用到合约层
func convertSkillRefsToContract(refs []types.SkillRef) []contract.SkillRef {
	result := make([]contract.SkillRef, 0, len(refs))
	for _, ref := range refs {
		result = append(result, contract.SkillRef{
			SkillCode: ref.SkillCode,
			Version:   ref.Version,
		})
	}
	return result
}

// convertChannelRefsToContract 转换渠道引用到合约层
func convertChannelRefsToContract(refs []types.ChannelRef) []contract.ChannelRef {
	result := make([]contract.ChannelRef, 0, len(refs))
	for _, ref := range refs {
		result = append(result, contract.ChannelRef{
			Type: ref.Type,
		})
	}
	return result
}

// convertKnowledgeRefsToContract 转换知识库引用到合约层
func convertKnowledgeRefsToContract(refs []types.KnowledgeRef) []contract.KnowledgeRef {
	result := make([]contract.KnowledgeRef, 0, len(refs))
	for _, ref := range refs {
		result = append(result, contract.KnowledgeRef{
			Type:      ref.Type,
			DatasetID: ref.DatasetID,
			Repo:      ref.Repo,
		})
	}
	return result
}
