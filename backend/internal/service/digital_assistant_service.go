package service

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/insmtx/SingerOS/backend/internal/api/auth"
	"github.com/insmtx/SingerOS/backend/internal/api/contract"
	"github.com/insmtx/SingerOS/backend/internal/infra/db"
	"github.com/insmtx/SingerOS/backend/types"
)

var _ contract.DigitalAssistantService = (*digitalAssistantService)(nil)

type digitalAssistantService struct {
	db *gorm.DB
}

func NewDigitalAssistantService(db *gorm.DB) contract.DigitalAssistantService {
	return &digitalAssistantService{
		db: db,
	}
}

func (s *digitalAssistantService) CreateDigitalAssistant(ctx context.Context, req *contract.CreateDigitalAssistantRequest) (*contract.DigitalAssistant, error) {
	orgID, err := getOrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Code == "" {
		return nil, errors.New("code is required")
	}
	if req.Name == "" {
		return nil, errors.New("name is required")
	}

	exists, err := db.DigitalAssistantCodeExists(ctx, s.db, req.Code, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("digital assistant with this code already exists")
	}

	caller, _ := auth.FromContext(ctx)
	da := &types.DigitalAssistant{
		Code:        req.Code,
		OrgID:       orgID,
		OwnerID:     caller.Uin,
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

	if err := db.CreateDigitalAssistant(ctx, s.db, da); err != nil {
		return nil, err
	}

	return convertToContractDigitalAssistant(da), nil
}

func (s *digitalAssistantService) GetDigitalAssistantByID(ctx context.Context, id uint) (*contract.DigitalAssistantDetail, error) {
	orgID, err := getOrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	da, err := db.GetDigitalAssistantByID(ctx, s.db, id)
	if err != nil {
		return nil, err
	}
	if da == nil {
		return nil, errors.New("digital assistant not found")
	}

	if err := verifyOrgPermission(da.OrgID, orgID); err != nil {
		return nil, err
	}

	return &contract.DigitalAssistantDetail{
		DigitalAssistant: *convertToContractDigitalAssistant(da),
	}, nil
}

func (s *digitalAssistantService) GetDigitalAssistantByCode(ctx context.Context, code string) (*contract.DigitalAssistantDetail, error) {
	orgID, err := getOrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	da, err := db.GetDigitalAssistantByCode(ctx, s.db, code)
	if err != nil {
		return nil, err
	}
	if da == nil {
		return nil, errors.New("digital assistant not found")
	}

	if err := verifyOrgPermission(da.OrgID, orgID); err != nil {
		return nil, err
	}

	return &contract.DigitalAssistantDetail{
		DigitalAssistant: *convertToContractDigitalAssistant(da),
	}, nil
}

func (s *digitalAssistantService) UpdateDigitalAssistant(ctx context.Context, id uint, req *contract.UpdateDigitalAssistantRequest) (*contract.DigitalAssistant, error) {
	orgID, err := getOrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	da, err := db.GetDigitalAssistantByID(ctx, s.db, id)
	if err != nil {
		return nil, err
	}
	if da == nil {
		return nil, errors.New("digital assistant not found")
	}

	if err := verifyOrgPermission(da.OrgID, orgID); err != nil {
		return nil, err
	}

	if req.Name != "" {
		da.Name = req.Name
	}
	if req.Description != "" {
		da.Description = req.Description
	}
	if req.Avatar != "" {
		da.Avatar = req.Avatar
	}
	if req.Config != nil {
		da.Config = types.AssistantConfig{
			Runtime:   types.RuntimeConfig{Type: req.Config.Runtime.Type},
			LLM:       types.LLMConfig{Type: req.Config.LLM.Type},
			Skills:    convertSkillRefs(req.Config.Skills),
			Channels:  convertChannelRefs(req.Config.Channels),
			Knowledge: convertKnowledgeRefs(req.Config.Knowledge),
			Memory:    types.MemoryConfig{Type: req.Config.Memory.Type},
			Policies:  types.PolicyConfig{Type: req.Config.Policies.Type},
		}
	}
	da.UpdatedAt = time.Now()

	if err := db.UpdateDigitalAssistant(ctx, s.db, da); err != nil {
		return nil, err
	}

	return convertToContractDigitalAssistant(da), nil
}

func (s *digitalAssistantService) DeleteDigitalAssistant(ctx context.Context, id uint) error {
	orgID, err := getOrgIDFromContext(ctx)
	if err != nil {
		return err
	}

	da, err := db.GetDigitalAssistantByID(ctx, s.db, id)
	if err != nil {
		return err
	}
	if da == nil {
		return errors.New("digital assistant not found")
	}

	if err := verifyOrgPermission(da.OrgID, orgID); err != nil {
		return err
	}

	return db.DeleteDigitalAssistant(ctx, s.db, id)
}

func (s *digitalAssistantService) ListDigitalAssistant(ctx context.Context, req *contract.ListDigitalAssistantRequest) (*contract.DigitalAssistantList, error) {
	orgID, err := getOrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	entities, total, err := db.ListDigitalAssistant(ctx, s.db, &orgID, nil, req.Status, req.Keyword, req.Page, req.PerPage)
	if err != nil {
		return nil, err
	}

	items := make([]contract.DigitalAssistant, 0, len(entities))
	for _, entity := range entities {
		items = append(items, *convertToContractDigitalAssistant(entity))
	}

	return &contract.DigitalAssistantList{
		Total: total,
		Page:  req.Page,
		Items: items,
	}, nil
}

func (s *digitalAssistantService) UpdateDigitalAssistantConfig(ctx context.Context, id uint, req *contract.UpdateDigitalAssistantConfigRequest) (*contract.DigitalAssistant, error) {
	orgID, err := getOrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	da, err := db.GetDigitalAssistantByID(ctx, s.db, id)
	if err != nil {
		return nil, err
	}
	if da == nil {
		return nil, errors.New("digital assistant not found")
	}

	if err := verifyOrgPermission(da.OrgID, orgID); err != nil {
		return nil, err
	}

	da.Config = types.AssistantConfig{
		Runtime:   types.RuntimeConfig{Type: req.Config.Runtime.Type},
		LLM:       types.LLMConfig{Type: req.Config.LLM.Type},
		Skills:    convertSkillRefs(req.Config.Skills),
		Channels:  convertChannelRefs(req.Config.Channels),
		Knowledge: convertKnowledgeRefs(req.Config.Knowledge),
		Memory:    types.MemoryConfig{Type: req.Config.Memory.Type},
		Policies:  types.PolicyConfig{Type: req.Config.Policies.Type},
	}
	da.UpdatedAt = time.Now()

	if err := db.UpdateDigitalAssistant(ctx, s.db, da); err != nil {
		return nil, err
	}

	return convertToContractDigitalAssistant(da), nil
}

func (s *digitalAssistantService) UpdateDigitalAssistantStatus(ctx context.Context, id uint, req *contract.UpdateDigitalAssistantStatusRequest) error {
	orgID, err := getOrgIDFromContext(ctx)
	if err != nil {
		return err
	}

	da, err := db.GetDigitalAssistantByID(ctx, s.db, id)
	if err != nil {
		return err
	}
	if da == nil {
		return errors.New("digital assistant not found")
	}

	if err := verifyOrgPermission(da.OrgID, orgID); err != nil {
		return err
	}

	da.Status = req.Status
	da.UpdatedAt = time.Now()

	return db.UpdateDigitalAssistant(ctx, s.db, da)
}

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

func convertChannelRefs(reqRefs []contract.ChannelRef) []types.ChannelRef {
	result := make([]types.ChannelRef, 0, len(reqRefs))
	for _, ref := range reqRefs {
		result = append(result, types.ChannelRef{
			Type: ref.Type,
		})
	}
	return result
}

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
	}
}

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

func convertChannelRefsToContract(refs []types.ChannelRef) []contract.ChannelRef {
	result := make([]contract.ChannelRef, 0, len(refs))
	for _, ref := range refs {
		result = append(result, contract.ChannelRef{
			Type: ref.Type,
		})
	}
	return result
}

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
