package externalcli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/insmtx/SingerOS/backend/internal/agent"
)

func buildPrompt(req *agent.RequestContext) string {
	if req == nil {
		return ""
	}

	sections := []string{
		"# SingerOS Runtime Request",
		"你是 SingerOS 智能助手。请基于下面的结构化上下文完成任务，并在最终输出中给出清晰、可执行的结果。",
	}

	if req.Assistant.ID != "" || req.Assistant.Name != "" || req.Assistant.Role != "" || req.Assistant.SystemPrompt != "" {
		sections = append(sections, formatJSONSection("Assistant", req.Assistant))
	}
	if req.Actor.UserID != "" || req.Actor.Channel != "" || req.Actor.ExternalID != "" || req.Actor.AccountID != "" {
		sections = append(sections, formatJSONSection("Actor", req.Actor))
	}
	if req.Conversation.ID != "" || len(req.Conversation.Messages) > 0 {
		sections = append(sections, formatJSONSection("Conversation", req.Conversation))
	}
	sections = append(sections, formatJSONSection("Input", req.Input))
	if len(req.Capability.AllowedTools) > 0 {
		sections = append(sections, formatJSONSection("Capability", req.Capability))
	}
	if req.Policy.RequireApproval {
		sections = append(sections, formatJSONSection("Policy", req.Policy))
	}
	if len(req.Metadata) > 0 {
		sections = append(sections, formatJSONSection("Metadata", req.Metadata))
	}

	sections = append(sections, `## Output Contract
- 使用中文输出最终结果。
- 不要编造未实际执行的命令、文件、链接、ID 或状态。
- 如果需要执行真实环境操作，请使用 runtime 已配置的工具或 MCP 能力。`)

	return strings.Join(sections, "\n\n")
}

func formatJSONSection(title string, value any) string {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("## %s\n%v", title, value)
	}
	return fmt.Sprintf("## %s\n```json\n%s\n```", title, string(encoded))
}
