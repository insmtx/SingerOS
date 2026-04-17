package runtime

import (
	"strings"
	"testing"

	"github.com/insmtx/SingerOS/backend/interaction"
)

func TestBuildQueryFromEvent(t *testing.T) {
	query := buildQueryFromEvent(&interaction.Event{
		Channel:    "github",
		EventID:    "evt-1",
		TraceID:    "trace-1",
		EventType:  "github.pull_request.opened",
		Actor:      "octocat",
		Repository: "insmtx/SingerOS",
		Context: map[string]interface{}{
			"delivery": "123",
		},
		Payload: map[string]interface{}{
			"title":       "Fix runtime wiring",
			"description": "Refactor the runtime layer",
			"comment":     "please review",
			"pull_request": map[string]interface{}{
				"title": "Fix runtime wiring",
			},
		},
	})

	expectedFragments := []string{
		"You are handling an external event inside SingerOS.",
		"- channel: github",
		"- event_type: github.pull_request.opened",
		"- actor: octocat",
		"- repository: insmtx/SingerOS",
		"- event_id: evt-1",
		"- trace_id: trace-1",
		"This appears to be a GitHub pull request event.",
		"Event context:",
		`"delivery": "123"`,
		"Raw event payload:",
		"Fix runtime wiring",
		"Refactor the runtime layer",
		"please review",
	}
	for _, fragment := range expectedFragments {
		if !strings.Contains(query, fragment) {
			t.Fatalf("expected query to contain %q, got %q", fragment, query)
		}
	}
}

func TestBuildQueryFromPushEvent(t *testing.T) {
	query := buildQueryFromEvent(&interaction.Event{
		Channel:    "github",
		EventType:  "push",
		Actor:      "octocat",
		Repository: "insmtx/SingerOS",
		Context:    map[string]interface{}{"ref": "refs/heads/main"},
		Payload:    map[string]interface{}{"commits": []interface{}{map[string]interface{}{"message": "feat: add runtime"}}},
	})

	expectedFragments := []string{
		"- event_type: push",
		"This appears to be a GitHub push event.",
		`"message": "feat: add runtime"`,
	}
	for _, fragment := range expectedFragments {
		if !strings.Contains(query, fragment) {
			t.Fatalf("expected query to contain %q, got %q", fragment, query)
		}
	}
}
