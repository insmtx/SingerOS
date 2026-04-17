package tools

import (
	"context"
	"testing"
)

type mockTool struct {
	info *ToolInfo
}

func (m *mockTool) Info() *ToolInfo {
	return m.info
}

func (m *mockTool) Validate(input map[string]interface{}) error {
	return nil
}

func (m *mockTool) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{"ok": true}, nil
}

func TestRegistryListInfosByProvider(t *testing.T) {
	registry := NewRegistry()

	if err := registry.Register(&mockTool{
		info: &ToolInfo{
			Name:        "github.pr.get_metadata",
			Description: "Read pull request metadata",
			Provider:    "github",
			ReadOnly:    true,
		},
	}); err != nil {
		t.Fatalf("register github tool: %v", err)
	}

	if err := registry.Register(&mockTool{
		info: &ToolInfo{
			Name:        "feishu.message.send",
			Description: "Send a message",
			Provider:    "feishu",
			ReadOnly:    false,
		},
	}); err != nil {
		t.Fatalf("register feishu tool: %v", err)
	}

	infos := registry.ListInfos()
	if len(infos) != 2 {
		t.Fatalf("expected 2 tool infos, got %d", len(infos))
	}
	if infos[0].Name != "feishu.message.send" || infos[1].Name != "github.pr.get_metadata" {
		t.Fatalf("expected sorted tool infos, got %+v", infos)
	}

	githubInfos := registry.ListInfosByProvider("GitHub")
	if len(githubInfos) != 1 {
		t.Fatalf("expected 1 github tool, got %d", len(githubInfos))
	}
	if githubInfos[0].Name != "github.pr.get_metadata" {
		t.Fatalf("unexpected github tool info: %+v", githubInfos[0])
	}
}
