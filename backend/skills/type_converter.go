// skills 包提供 SingerOS 的技能系统定义和实现
//
// 技能是 SingerOS 的核心能力单元，可以是本地实现的功能，
// 也可以是远程 API 服务。该包定义了技能接口、基础实现和相关类型。
package skills

import (
	"github.com/insmtx/SingerOS/backend/types"
)

// ConvertToDBModel 将Skill接口实例转换为可用于数据库存储的types.Skill模型
func ConvertToDBModel(skill Skill) *types.Skill {
	info := skill.Info()

	return &types.Skill{
		Code:         info.ID,
		Name:         info.Name,
		Description:  info.Description,
		Author:       info.Author,
		Version:      info.Version,
		Category:     info.Category,
		SkillType:    string(info.SkillType),
		InputSchema:  inputSchemaToMap(info.InputSchema),
		OutputSchema: outputSchemaToMap(info.OutputSchema),
		Permissions:  permissionsToInterfaceSlice(info.Permissions),
		Config:       map[string]interface{}{},
		Status:       "active",
		IsSystem:     false,
	}
}

// ConvertFromDBModel 将数据库中的types.Skill模型转换为Skill引用信息
func ConvertFromDBModel(model *types.Skill) *SkillInfo {
	return &SkillInfo{
		ID:           model.Code,
		Name:         model.Name,
		Description:  model.Description,
		Author:       model.Author,
		Version:      model.Version,
		Category:     model.Category,
		SkillType:    SkillType(model.SkillType),
		Icon:         model.Icon,
		InputSchema:  convertToInputSchema(model.InputSchema),
		OutputSchema: convertToOutputSchema(model.OutputSchema),
		Permissions:  convertToPermissions(model.Permissions),
	}
}

// interfaceMap 将interface{}类型的map转换通用map
func interfaceMap(schema interface{}) map[string]interface{} {
	if m, ok := schema.(map[string]interface{}); ok {
		return m
	}
	return make(map[string]interface{})
}

// permissionsToInterfaceSlice 将Permission切片转换为interface{}切片
func permissionsToInterfaceSlice(perms []Permission) []interface{} {
	result := make([]interface{}, len(perms))
	for i, perm := range perms {
		result[i] = perm
	}
	return result
}

// inputSchemaToMap 将InputSchema结构转换为map
func inputSchemaToMap(schema InputSchema) map[string]interface{} {
	result := make(map[string]interface{})

	result["type"] = schema.Type

	if len(schema.Required) > 0 {
		result["required"] = schema.Required
	}

	if len(schema.Properties) > 0 {
		props := make(map[string]interface{})
		for k, v := range schema.Properties {
			props[k] = propertyToMap(v)
		}
		result["properties"] = props
	}

	return result
}

// outputSchemaToMap 将OutputSchema结构转换为map
func outputSchemaToMap(schema OutputSchema) map[string]interface{} {
	result := make(map[string]interface{})

	result["type"] = schema.Type

	if len(schema.Required) > 0 {
		result["required"] = schema.Required
	}

	if len(schema.Properties) > 0 {
		props := make(map[string]interface{})
		for k, v := range schema.Properties {
			props[k] = propertyToMap(v)
		}
		result["properties"] = props
	}

	return result
}

// propertyToMap 将Property结构转换为map
func propertyToMap(prop *Property) map[string]interface{} {
	result := make(map[string]interface{})

	if prop.Type != "" {
		result["type"] = prop.Type
	}

	if prop.Title != "" {
		result["title"] = prop.Title
	}

	if prop.Description != "" {
		result["description"] = prop.Description
	}

	if prop.Default != nil {
		result["default"] = prop.Default
	}

	if prop.Items != nil {
		result["items"] = propertyToMap(prop.Items)
	}

	if len(prop.Enum) > 0 {
		result["enum"] = prop.Enum
	}

	return result
}

// convertToPermissions 转换权限切片
func convertToPermissions(interfs []interface{}) []Permission {
	perms := make([]Permission, 0)

	for _, interf := range interfs {
		if permMap, ok := interf.(map[string]interface{}); ok {
			var perm Permission

			if resource, exists := permMap["resource"]; exists {
				if resourceStr, ok := resource.(string); ok {
					perm.Resource = resourceStr
				}
			}

			if action, exists := permMap["action"]; exists {
				if actionStr, ok := action.(string); ok {
					perm.Action = actionStr
				}
			}

			perms = append(perms, perm)
		}
	}

	return perms
}

// convertToInputSchema 将interface{}类型的map转换为InputSchema
func convertToInputSchema(interf map[string]interface{}) InputSchema {
	if interf == nil {
		return InputSchema{}
	}

	schema := InputSchema{}
	if v, ok := interf["type"]; ok {
		if s, ok := v.(string); ok {
			schema.Type = s
		}
	}

	if v, ok := interf["required"]; ok {
		switch val := v.(type) {
		case []interface{}:
			for _, item := range val {
				if s, ok := item.(string); ok {
					schema.Required = append(schema.Required, s)
				}
			}
		case []string:
			schema.Required = val
		}
	}

	if v, ok := interf["properties"]; ok {
		properties := make(map[string]*Property)
		if propMap, ok := v.(map[string]interface{}); ok {
			for k, v := range propMap {
				properties[k] = mapToProperty(v)
			}
			schema.Properties = properties
		}
	}

	return schema
}

// convertToOutputSchema 将interface{}类型的map转换为OutputSchema
func convertToOutputSchema(interf map[string]interface{}) OutputSchema {
	if interf == nil {
		return OutputSchema{}
	}

	schema := OutputSchema{}
	if v, ok := interf["type"]; ok {
		if s, ok := v.(string); ok {
			schema.Type = s
		}
	}

	if v, ok := interf["required"]; ok {
		switch val := v.(type) {
		case []interface{}:
			for _, item := range val {
				if s, ok := item.(string); ok {
					schema.Required = append(schema.Required, s)
				}
			}
		case []string:
			schema.Required = val
		}
	}

	if v, ok := interf["properties"]; ok {
		properties := make(map[string]*Property)
		if propMap, ok := v.(map[string]interface{}); ok {
			for k, v := range propMap {
				properties[k] = mapToProperty(v)
			}
			schema.Properties = properties
		}
	}

	return schema
}

// mapToProperty 将interface{}转为Property
func mapToProperty(interf interface{}) *Property {
	prop := &Property{}

	if m, ok := interf.(map[string]interface{}); ok {
		if v, ok := m["type"]; ok {
			if s, ok := v.(string); ok {
				prop.Type = s
			}
		}

		if v, ok := m["title"]; ok {
			if s, ok := v.(string); ok {
				prop.Title = s
			}
		}

		if v, ok := m["description"]; ok {
			if s, ok := v.(string); ok {
				prop.Description = s
			}
		}

		if v, ok := m["default"]; ok {
			prop.Default = v
		}

		if v, ok := m["items"]; ok {
			prop.Items = mapToProperty(v)
		}

		if v, ok := m["enum"]; ok {
			if enumList, ok := v.([]interface{}); ok {
				enumStrs := make([]string, len(enumList))
				for i, enumItem := range enumList {
					if s, ok := enumItem.(string); ok {
						enumStrs[i] = s
					}
				}
				prop.Enum = enumStrs
			}
		}
	}

	return prop
}
