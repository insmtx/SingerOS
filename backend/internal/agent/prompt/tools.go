package prompt

import (
	"strings"

	"github.com/insmtx/SingerOS/backend/tools"
)

// ToolsContext is the prompt-ready projection of the runtime tool registry.
type ToolsContext struct {
	SummarySection string
}

// BuildToolsContext converts a tool registry into a compact summary for runtime prompts.
func BuildToolsContext(registry *tools.Registry) *ToolsContext {
	if registry == nil {
		return &ToolsContext{}
	}

	infos := registry.ListInfos()
	if len(infos) == 0 {
		return &ToolsContext{}
	}

	return &ToolsContext{
		SummarySection: buildToolsSummary(infos),
	}
}

func buildToolsSummary(infos []tools.ToolInfo) string {
	var builder strings.Builder

	builder.WriteString("Available tools:\n")
	for _, info := range infos {
		builder.WriteString("- ")
		builder.WriteString(info.Name)
		builder.WriteString(": ")
		builder.WriteString(info.Description)
		if info.Provider != "" {
			builder.WriteString(" [provider=")
			builder.WriteString(info.Provider)
			builder.WriteString("]")
		}
		if info.ReadOnly {
			builder.WriteString(" [mode=read]")
		} else {
			builder.WriteString(" [mode=write]")
		}
		if info.InputSchema != nil && len(info.InputSchema.Required) > 0 {
			builder.WriteString(" [required=")
			builder.WriteString(strings.Join(info.InputSchema.Required, ","))
			builder.WriteString("]")
		}
		builder.WriteString("\n")
	}

	builder.WriteString("\nUse read tools first to gather context before calling write tools.")
	return strings.TrimSpace(builder.String())
}
