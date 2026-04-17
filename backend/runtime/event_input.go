package runtime

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/insmtx/SingerOS/backend/interaction"
)

func buildQueryFromEvent(event *interaction.Event) string {
	if event == nil {
		return ""
	}

	sections := []string{
		"You are handling an external event inside SingerOS.",
		buildEventEnvelope(event),
		buildEventTask(event),
	}

	if contextSection := buildJSONSection("Event context", event.Context); contextSection != "" {
		sections = append(sections, contextSection)
	}
	if payloadSection := buildJSONSection("Raw event payload", event.Payload); payloadSection != "" {
		sections = append(sections, payloadSection)
	}

	return strings.Join(filterEmptyStrings(sections), "\n\n")
}

func buildEventEnvelope(event *interaction.Event) string {
	lines := []string{"Event envelope:"}
	if event.Channel != "" {
		lines = append(lines, "- channel: "+event.Channel)
	}
	if event.EventType != "" {
		lines = append(lines, "- event_type: "+event.EventType)
	}
	if event.Actor != "" {
		lines = append(lines, "- actor: "+event.Actor)
	}
	if event.Repository != "" {
		lines = append(lines, "- repository: "+event.Repository)
	}
	if event.EventID != "" {
		lines = append(lines, "- event_id: "+event.EventID)
	}
	if event.TraceID != "" {
		lines = append(lines, "- trace_id: "+event.TraceID)
	}
	return strings.Join(lines, "\n")
}

func buildEventTask(event *interaction.Event) string {
	base := "Task:\n- Understand what happened from the event payload.\n- Use available skills and tools to gather authoritative details before making claims.\n- If the event requires an external response, decide whether to publish one and keep it evidence-based."

	switch event.EventType {
	case "pull_request", "github.pull_request", "github.pull_request.opened":
		return base + "\n- This appears to be a GitHub pull request event. Review the change carefully before publishing any GitHub review."
	case "push", "github.push":
		return base + "\n- This appears to be a GitHub push event. Use the commit list and repository context to understand what changed before deciding whether any follow-up is needed."
	case "issue_comment", "github.issue_comment":
		return base + "\n- This appears to be a GitHub issue or pull request comment event. Decide whether a reply is needed."
	default:
		return base
	}
}

func buildJSONSection(title string, value interface{}) string {
	if value == nil {
		return ""
	}

	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return ""
	}

	text := string(encoded)
	if len(text) > 6000 {
		text = text[:6000] + "\n... (truncated)"
	}

	return fmt.Sprintf("%s:\n```json\n%s\n```", title, text)
}

func filterEmptyStrings(values []string) []string {
	filtered := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		filtered = append(filtered, value)
	}
	return filtered
}
