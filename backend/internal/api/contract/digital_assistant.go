package contract

import "context"

// DigitalAssistantService 定义数字助手服务接口
type DigitalAssistantService interface {
	// 创建数字助手
	CreateDigitalAssistant(ctx context.Context, req *CreateDigitalAssistantRequest) (*DigitalAssistant, error)

	// 根据 ID 获取数字助手详情（需验证组织权限）
	GetDigitalAssistantByID(ctx context.Context, orgID uint, id uint) (*DigitalAssistantDetail, error)

	// 根据 Code 获取数字助手详情（需验证组织权限）
	GetDigitalAssistantByCode(ctx context.Context, orgID uint, code string) (*DigitalAssistantDetail, error)

	// 更新数字助手信息
	UpdateDigitalAssistant(ctx context.Context, id uint, req *UpdateDigitalAssistantRequest) (*DigitalAssistant, error)

	// 删除数字助手
	DeleteDigitalAssistant(ctx context.Context, id uint) error

	// 查询数字助手列表
	ListDigitalAssistant(ctx context.Context, req *ListDigitalAssistantRequest) (*DigitalAssistantList, error)

	// 更新数字助手配置
	UpdateDigitalAssistantConfig(ctx context.Context, id uint, req *UpdateDigitalAssistantConfigRequest) (*DigitalAssistant, error)

	// 更新数字助手状态
	UpdateDigitalAssistantStatus(ctx context.Context, id uint, req *UpdateDigitalAssistantStatusRequest) error
}
