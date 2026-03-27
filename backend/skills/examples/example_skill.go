// examples 包提供 SingerOS 技能系统的示例实现
//
// 该包包含示例技能，演示如何实现 Skill 接口，
// 供开发者参考和学习。
package examples

import (
	"context"
	"fmt"
	"github.com/insmtx/SingerOS/backend/skills"
)

// ExampleSkill 是一个示例技能，演示技能接口的实现方式
type ExampleSkill struct {
	skills.BaseSkill
}

// NewExampleSkill 创建一个新的示例技能实例
func NewExampleSkill() *ExampleSkill {
	return &ExampleSkill{
		BaseSkill: skills.BaseSkill{
			InfoData: &skills.SkillInfo{
				ID:          "example.hello_world",
				Name:        "Hello World Skill",
				Description: "一个简单的示例技能，输出问候信息",
				Version:     "1.0.0",
				Category:    "tool",
				Author:      "SingerOS Team",
				SkillType:   skills.LocalSkill,
				Permissions: []skills.Permission{
					{Resource: "greetings", Action: "execute"},
				},
				InputSchema: skills.InputSchema{
					Type:     "object",
					Required: []string{"name"},
					Properties: map[string]*skills.Property{
						"name": {
							Type:        "string",
							Title:       "姓名",
							Description: "要问候的人的名字",
						},
						"greeting": {
							Type:        "string",
							Title:       "问候语",
							Description: "自定义的问候语",
							Default:     "Hello",
						},
					},
				},
				OutputSchema: skills.OutputSchema{
					Type:     "object",
					Required: []string{"message"},
					Properties: map[string]*skills.Property{
						"message": {
							Type:        "string",
							Title:       "消息",
							Description: "生成的问候消息",
						},
					},
				},
			},
		},
	}
}

// Execute 执行示例技能的业务逻辑
func (s *ExampleSkill) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// 检查必要的参数
	nameVal, exists := input["name"]
	if !exists {
		return nil, fmt.Errorf("必要参数 'name' 未提供")
	}

	name, ok := nameVal.(string)
	if !ok {
		return nil, fmt.Errorf("'name' 参数必须为字符串类型")
	}

	// 获取可选参数
	greetingVal, exists := input["greeting"]
	greeting := "Hello" // 默认值
	if exists {
		if greetingStr, ok := greetingVal.(string); ok {
			greeting = greetingStr
		} else {
			return nil, fmt.Errorf("'greeting' 参数必须为字符串类型")
		}
	}

	// 构建结果
	message := fmt.Sprintf("%s, %s!", greeting, name)

	result := map[string]interface{}{
		"message": message,
	}

	return result, nil
}
