package prompt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	skilltools "github.com/insmtx/SingerOS/backend/tools/skill"
)

func TestBuildSkillsContext(t *testing.T) {
	rootDir := t.TempDir()

	alwaysSkill := `---
name: github-pr-review
description: Review pull requests.
metadata:
  singeros:
    category: github
    always: true
    requires_tools: [github.pr.get_files]
---
# Review Process

Always load this.
`
	optionalSkill := `---
name: issue-reply
description: Reply to issues.
---
# Issue Reply

Load on demand.
`

	if err := os.MkdirAll(filepath.Join(rootDir, "github-pr-review"), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(rootDir, "issue-reply"), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rootDir, "github-pr-review", "SKILL.md"), []byte(alwaysSkill), 0o644); err != nil {
		t.Fatalf("write skill failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rootDir, "issue-reply", "SKILL.md"), []byte(optionalSkill), 0o644); err != nil {
		t.Fatalf("write skill failed: %v", err)
	}

	catalog, err := skilltools.NewCatalog(os.DirFS(rootDir))
	if err != nil {
		t.Fatalf("load catalog failed: %v", err)
	}

	context, err := BuildSkillsContext(catalog)
	if err != nil {
		t.Fatalf("build skills context failed: %v", err)
	}

	if !strings.Contains(context.SummarySection, "github-pr-review: Review pull requests.") {
		t.Fatalf("summary does not include github-pr-review: %s", context.SummarySection)
	}
	if !strings.Contains(context.SummarySection, "issue-reply: Reply to issues.") {
		t.Fatalf("summary does not include issue-reply: %s", context.SummarySection)
	}
	if len(context.AlwaysSections) != 1 {
		t.Fatalf("expected 1 always section, got %d", len(context.AlwaysSections))
	}
	if !strings.Contains(context.AlwaysSections[0], "Always load this.") {
		t.Fatalf("always section missing body: %s", context.AlwaysSections[0])
	}
}
