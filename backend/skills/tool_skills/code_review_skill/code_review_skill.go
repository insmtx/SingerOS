package code_review_skill

import (
	"context"
	"fmt"

	"github.com/insmtx/SingerOS/backend/skills"
)

// CodeReviewSkill 实现代码审阅技能
type CodeReviewSkill struct {
	skills.BaseSkill
}

// NewCodeReviewSkill 创建一个新的代码审阅技能实例
func NewCodeReviewSkill() skills.Skill {
	return &CodeReviewSkill{
		BaseSkill: skills.BaseSkill{
			InfoData: &skills.SkillInfo{
				ID:          "code.review",
				Name:        "Code Review Skill",
				Description: "Reviews code and provides feedback",
				Version:     "1.0.0",
				Category:    "programming",
				SkillType:   skills.LocalSkill,
				InputSchema: skills.InputSchema{
					Type:     "object",
					Required: []string{"code", "language"},
					Properties: map[string]*skills.Property{
						"code": {
							Type:        "string",
							Description: "The code to review",
						},
						"language": {
							Type:        "string",
							Description: "Programming language of the code",
						},
						"context": {
							Type:        "string",
							Description: "Additional context for the review",
						},
					},
				},
				OutputSchema: skills.OutputSchema{
					Type: "object",
					Properties: map[string]*skills.Property{
						"feedback": {
							Type:        "string",
							Description: "Detailed code review feedback",
						},
						"issues": {
							Type: "array",
							Items: &skills.Property{
								Type:        "object",
								Description: "Issue in code review",
							},
						},
					},
				},
			},
		},
	}
}

// Execute 执行代码审阅
func (s *CodeReviewSkill) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	code, ok := input["code"].(string)
	if !ok || code == "" {
		return nil, fmt.Errorf("missing or invalid 'code' parameter")
	}

	language, ok := input["language"].(string)
	if !ok {
		language = "unknown"
	}

	contextInfo, _ := input["context"].(string)

	// 执行代码审阅分析（模拟）
	feedback := generateCodeReviewFeedback(code, language, contextInfo)

	result := map[string]interface{}{
		"feedback": feedback,
		"summary":  fmt.Sprintf("Reviewed %d lines of %s code", len(code)/10, language), // approximate the number of lines
		"issues":   []map[string]interface{}{},
	}

	// 在真实实现中，我们将分析代码并找出问题
	return result, nil
}

// generateCodeReviewFeedback 生成代码审阅反馈
func generateCodeReviewFeedback(code, language, contextInfo string) string {
	// 简单的模拟实现 - 在实际情况下，这可能会使用更复杂的分析逻辑
	feedback := fmt.Sprintf("Code review for %s:\n", language)
	feedback += fmt.Sprintf("- The code appears to be well-structured.\n")
	feedback += fmt.Sprintf("- Language-specific suggestions for %s would appear here.\n", language)
	feedback += fmt.Sprintf("- Additional context: %s.\n", contextInfo)

	return feedback
}
