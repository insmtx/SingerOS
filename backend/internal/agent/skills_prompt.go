package agent

import (
	"fmt"
	"strings"

	skilltools "github.com/insmtx/SingerOS/backend/tools/skill"
)

type skillsContext struct {
	SummarySection string
	AlwaysSections []string
}

func buildSkillsContext(catalog *skilltools.Catalog) (*skillsContext, error) {
	if catalog == nil {
		return &skillsContext{}, nil
	}

	summaries := catalog.List()
	if len(summaries) == 0 {
		return &skillsContext{}, nil
	}

	context := &skillsContext{
		SummarySection: buildSkillSummarySection(summaries),
		AlwaysSections: make([]string, 0),
	}

	for _, summary := range summaries {
		if !summary.Always {
			continue
		}

		entry, err := catalog.Get(summary.Name)
		if err != nil {
			return nil, fmt.Errorf("load always skill %s: %w", summary.Name, err)
		}

		var section strings.Builder
		section.WriteString("## Skill: ")
		section.WriteString(entry.Manifest.Name)
		section.WriteString("\n")
		section.WriteString(entry.Body)
		context.AlwaysSections = append(context.AlwaysSections, strings.TrimSpace(section.String()))
	}

	return context, nil
}

func buildSkillSummarySection(summaries []skilltools.Summary) string {
	var builder strings.Builder

	builder.WriteString("Available skills:\n")
	for _, summary := range summaries {
		builder.WriteString("- ")
		builder.WriteString(summary.Name)
		builder.WriteString(": ")
		builder.WriteString(summary.Description)
		if summary.Category != "" {
			builder.WriteString(" [category=")
			builder.WriteString(summary.Category)
			builder.WriteString("]")
		}
		if len(summary.RequiresTools) > 0 {
			builder.WriteString(" [requires_tools=")
			builder.WriteString(strings.Join(summary.RequiresTools, ","))
			builder.WriteString("]")
		}
		builder.WriteString("\n")
	}

	builder.WriteString("\nLoad a skill body only when it is relevant to the current task.")
	return strings.TrimSpace(builder.String())
}
