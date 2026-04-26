package db

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/insmtx/SingerOS/backend/types"
)

// CreateDigitalAssistant 创建数字助手
func CreateDigitalAssistant(ctx context.Context, db *gorm.DB, da *types.DigitalAssistant) error {
	return db.WithContext(ctx).Create(da).Error
}

// GetDigitalAssistantByID 根据ID获取数字助手
func GetDigitalAssistantByID(ctx context.Context, db *gorm.DB, id uint) (*types.DigitalAssistant, error) {
	var entity types.DigitalAssistant
	err := db.WithContext(ctx).First(&entity, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// GetDigitalAssistantByCode 根据Code获取数字助手
func GetDigitalAssistantByCode(ctx context.Context, db *gorm.DB, code string) (*types.DigitalAssistant, error) {
	var entity types.DigitalAssistant
	err := db.WithContext(ctx).Where("code = ?", code).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// UpdateDigitalAssistant 更新数字助手
func UpdateDigitalAssistant(ctx context.Context, db *gorm.DB, da *types.DigitalAssistant) error {
	return db.WithContext(ctx).Save(da).Error
}

// DeleteDigitalAssistant 删除数字助手
func DeleteDigitalAssistant(ctx context.Context, db *gorm.DB, id uint) error {
	return db.WithContext(ctx).Delete(&types.DigitalAssistant{}, id).Error
}

// DigitalAssistantCodeExists 检查code是否存在（排除指定ID）
func DigitalAssistantCodeExists(ctx context.Context, db *gorm.DB, code string, excludeID uint) (bool, error) {
	var count int64
	query := db.WithContext(ctx).Model(&types.DigitalAssistant{}).Where("code = ?", code)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListDigitalAssistant 分页查询数字助手列表
func ListDigitalAssistant(ctx context.Context, db *gorm.DB, orgID *uint, ownerID *uint, status *string, keyword *string, page, perPage int) ([]*types.DigitalAssistant, int64, error) {
	var entities []*types.DigitalAssistant
	var total int64

	query := db.WithContext(ctx).Model(&types.DigitalAssistant{})

	if orgID != nil {
		query = query.Where("org_id = ?", *orgID)
	}
	if ownerID != nil {
		query = query.Where("owner_id = ?", *ownerID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if keyword != nil && *keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ? OR description LIKE ?", "%"+*keyword+"%", "%"+*keyword+"%", "%"+*keyword+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	err = query.Offset(offset).Limit(perPage).Order("created_at DESC").Find(&entities).Error
	if err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}
