package skilltools

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSkillUseToolListAndGet(t *testing.T) {
	catalog := newTestCatalog(t)
	tool := NewSkillUseTool(catalog)

	listResult, err := tool.Execute(context.Background(), map[string]interface{}{
		"action": actionList,
	})
	if err != nil {
		t.Fatalf("list skills failed: %v", err)
	}
	if listResult["ok"] != true {
		t.Fatalf("expected ok list result, got %#v", listResult)
	}
	if listResult["count"] != 1 {
		t.Fatalf("expected 1 skill, got %#v", listResult["count"])
	}

	getResult, err := tool.Execute(context.Background(), map[string]interface{}{
		"action": actionGet,
		"name":   "GITHUB-PR-REVIEW",
	})
	if err != nil {
		t.Fatalf("get skill failed: %v", err)
	}

	skill, ok := getResult["skill"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected skill object, got %#v", getResult["skill"])
	}
	if skill["name"] != "github-pr-review" {
		t.Fatalf("unexpected skill name: %#v", skill["name"])
	}
	if skill["body"] == "" {
		t.Fatalf("expected skill body")
	}

	if getResult["title"] != "Loaded skill: github-pr-review" {
		t.Fatalf("unexpected title: %#v", getResult["title"])
	}
	output, ok := getResult["output"].(string)
	if !ok || !strings.Contains(output, `<skill_content name="github-pr-review">`) {
		t.Fatalf("expected skill content output, got %#v", getResult["output"])
	}
	if !strings.Contains(output, "references/policy.md") {
		t.Fatalf("expected skill file list in output, got %s", output)
	}

	metadata, ok := getResult["metadata"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected metadata object, got %#v", getResult["metadata"])
	}
	files, ok := metadata["files"].([]string)
	if !ok || len(files) != 2 || files[0] != "references/large.md" || files[1] != "references/policy.md" {
		t.Fatalf("unexpected metadata files: %#v", metadata["files"])
	}
}

func TestSkillUseToolReadFile(t *testing.T) {
	catalog := newTestCatalog(t)
	tool := NewSkillUseTool(catalog)

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"action": actionReadFile,
		"name":   "github-pr-review",
		"path":   "references/policy.md",
	})
	if err != nil {
		t.Fatalf("read skill file failed: %v", err)
	}
	if result["ok"] != true {
		t.Fatalf("expected ok read result, got %#v", result)
	}
	if result["content"] != "policy content" {
		t.Fatalf("unexpected file content: %#v", result["content"])
	}
	if result["size"] != len("policy content") {
		t.Fatalf("unexpected file size: %#v", result["size"])
	}
	if result["truncated"] != false {
		t.Fatalf("expected untruncated file, got %#v", result["truncated"])
	}
}

func TestSkillUseToolReadFileTruncatesLargeContent(t *testing.T) {
	catalog := newTestCatalog(t)
	tool := NewSkillUseTool(catalog)

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"action": actionReadFile,
		"name":   "github-pr-review",
		"path":   "references/large.md",
	})
	if err != nil {
		t.Fatalf("read large skill file failed: %v", err)
	}
	if result["ok"] != true {
		t.Fatalf("expected ok read result, got %#v", result)
	}
	if result["truncated"] != true {
		t.Fatalf("expected truncated file, got %#v", result["truncated"])
	}
	content, ok := result["content"].(string)
	if !ok {
		t.Fatalf("expected string content, got %#v", result["content"])
	}
	if len(content) != maxSkillFileReadBytes {
		t.Fatalf("expected content length %d, got %d", maxSkillFileReadBytes, len(content))
	}
}

func TestSkillUseToolLoadsBundledWeatherSkillForWeatherQuery(t *testing.T) {
	catalog := newBundledSkillsCatalog(t)
	tool := NewSkillUseTool(catalog)

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"action": actionGet,
		"name":   "WEATHER",
	})
	if err != nil {
		t.Fatalf("get weather skill failed: %v", err)
	}
	if result["ok"] != true {
		t.Fatalf("expected ok weather skill result, got %#v", result)
	}

	skill, ok := result["skill"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected skill object, got %#v", result["skill"])
	}
	if skill["name"] != "weather" {
		t.Fatalf("unexpected skill name: %#v", skill["name"])
	}
	if skill["description"] != "Get current weather and forecasts (no API key required)." {
		t.Fatalf("unexpected weather skill description: %#v", skill["description"])
	}

	output, ok := result["output"].(string)
	if !ok {
		t.Fatalf("expected weather skill output string, got %#v", result["output"])
	}
	for _, expected := range []string{
		`<skill_content name="weather">`,
		`curl -s "wttr.in/London?format=3"`,
		"Open-Meteo",
		"current_weather=true",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected weather skill output to contain %q, got %s", expected, output)
		}
	}
}

func TestSkillUseToolMissingSkillReturnsAvailableNames(t *testing.T) {
	catalog := newTestCatalog(t)
	tool := NewSkillUseTool(catalog)

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"action": actionGet,
		"name":   "missing",
	})
	if err != nil {
		t.Fatalf("get missing skill should return structured result: %v", err)
	}
	if result["ok"] != false {
		t.Fatalf("expected not found result, got %#v", result)
	}

	available, ok := result["available"].([]string)
	if !ok {
		t.Fatalf("expected available skill names, got %#v", result["available"])
	}
	if len(available) != 1 || available[0] != "github-pr-review" {
		t.Fatalf("unexpected available skills: %#v", available)
	}
}

func TestSkillUseToolValidate(t *testing.T) {
	tool := NewSkillUseTool(nil)

	if err := tool.Validate(map[string]interface{}{}); err == nil {
		t.Fatalf("expected missing action to fail")
	}
	if err := tool.Validate(map[string]interface{}{"action": actionGet}); err == nil {
		t.Fatalf("expected missing name to fail")
	}
	if err := tool.Validate(map[string]interface{}{"action": "delete"}); err == nil {
		t.Fatalf("expected unsupported action to fail")
	}
}

func newTestCatalog(t *testing.T) *Catalog {
	t.Helper()

	rootDir := t.TempDir()
	skillDir := filepath.Join(rootDir, "github-pr-review")
	referencesDir := filepath.Join(skillDir, "references")
	if err := os.MkdirAll(referencesDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	skillDocument := `---
name: github-pr-review
description: Review GitHub pull requests.
version: 0.1.0
metadata:
  singeros:
    category: github
    tags: [github, pr, review]
    always: true
    requires_tools: [github.pr.get_files]
---
# Review

Read the pull request before reviewing.
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillDocument), 0o644); err != nil {
		t.Fatalf("write skill failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(referencesDir, "policy.md"), []byte("policy content"), 0o644); err != nil {
		t.Fatalf("write reference failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(referencesDir, "large.md"), []byte(strings.Repeat("a", maxSkillFileReadBytes+5)), 0o644); err != nil {
		t.Fatalf("write large reference failed: %v", err)
	}

	catalog, err := NewCatalog(os.DirFS(rootDir))
	if err != nil {
		t.Fatalf("load catalog failed: %v", err)
	}

	return catalog
}

func newBundledSkillsCatalog(t *testing.T) *Catalog {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("resolve current test file")
	}

	skillsDir := filepath.Join(filepath.Dir(currentFile), "..", "..", "skills")
	catalog, err := NewCatalog(os.DirFS(skillsDir))
	if err != nil {
		t.Fatalf("load bundled skills catalog: %v", err)
	}

	return catalog
}
