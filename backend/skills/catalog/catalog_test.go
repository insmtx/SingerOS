package catalog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCatalogLoadsSkillDocuments(t *testing.T) {
	rootDir := t.TempDir()
	skillDir := filepath.Join(rootDir, "github-pr-review")
	referencesDir := filepath.Join(skillDir, "references")

	if err := os.MkdirAll(referencesDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	skillDocument := `---
name: github-pr-review
description: Review GitHub pull requests with SingerOS conventions.
version: 1.0.0
metadata:
  singeros:
    category: github
    tags: [github, pr, review]
    always: true
    requires_tools: [github.pr.get_files, github.pr.publish_review]
---
# GitHub PR Review

Review steps here.
`
	if err := os.WriteFile(filepath.Join(skillDir, skillFileName), []byte(skillDocument), 0o644); err != nil {
		t.Fatalf("write skill failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(referencesDir, "policy.md"), []byte("policy content"), 0o644); err != nil {
		t.Fatalf("write reference failed: %v", err)
	}

	catalog, err := New(os.DirFS(rootDir))
	if err != nil {
		t.Fatalf("load catalog failed: %v", err)
	}

	summaries := catalog.List()
	if len(summaries) != 1 {
		t.Fatalf("expected 1 skill summary, got %d", len(summaries))
	}

	summary := summaries[0]
	if summary.Name != "github-pr-review" {
		t.Fatalf("expected skill name github-pr-review, got %s", summary.Name)
	}
	if !summary.Always {
		t.Fatalf("expected skill always flag to be true")
	}

	entry, err := catalog.Get("github-pr-review")
	if err != nil {
		t.Fatalf("get skill failed: %v", err)
	}
	if entry.Manifest.Metadata.SingerOS.Category != "github" {
		t.Fatalf("expected category github, got %s", entry.Manifest.Metadata.SingerOS.Category)
	}
	if entry.Body == "" {
		t.Fatalf("expected non-empty skill body")
	}

	referenceBody, err := catalog.ReadFile("github-pr-review", "references/policy.md")
	if err != nil {
		t.Fatalf("read skill file failed: %v", err)
	}
	if string(referenceBody) != "policy content" {
		t.Fatalf("unexpected reference body: %s", string(referenceBody))
	}
}

func TestCatalogDerivesNameWithoutFrontmatter(t *testing.T) {
	rootDir := t.TempDir()
	skillDir := filepath.Join(rootDir, "plain-skill")

	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, skillFileName), []byte("# Plain Skill"), 0o644); err != nil {
		t.Fatalf("write skill failed: %v", err)
	}

	catalog, err := New(os.DirFS(rootDir))
	if err != nil {
		t.Fatalf("load catalog failed: %v", err)
	}

	entry, err := catalog.Get("plain-skill")
	if err != nil {
		t.Fatalf("get skill failed: %v", err)
	}
	if entry.Manifest.Description != "plain-skill" {
		t.Fatalf("expected derived description plain-skill, got %s", entry.Manifest.Description)
	}
}

func TestCatalogRejectsPathTraversal(t *testing.T) {
	rootDir := t.TempDir()
	skillDir := filepath.Join(rootDir, "safe-skill")

	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, skillFileName), []byte("# Safe Skill"), 0o644); err != nil {
		t.Fatalf("write skill failed: %v", err)
	}

	catalog, err := New(os.DirFS(rootDir))
	if err != nil {
		t.Fatalf("load catalog failed: %v", err)
	}

	if _, err := catalog.ReadFile("safe-skill", "../secret.txt"); err == nil {
		t.Fatalf("expected traversal path to be rejected")
	}
}
