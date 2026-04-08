package code_review_skill

import (
	"context"
	"testing"

	"github.com/insmtx/SingerOS/backend/skills"
)

func TestCodeReviewSkill(t *testing.T) {
	skill := NewCodeReviewSkill()

	// Verify skill implements Skill interface
	_, ok := skill.(skills.Skill)
	if !ok {
		t.Error("CodeReviewSkill does not implement skills.Skill interface")
	}

	// Test basic info functions
	info := skill.Info()
	if info.ID != "code.review" {
		t.Errorf("Expected ID 'code.review', got '%s'", info.ID)
	}

	if info.Name != "Code Review Skill" {
		t.Errorf("Expected Name 'Code Review Skill', got '%s'", info.Name)
	}

	if info.Description != "Reviews code and provides feedback" {
		t.Errorf("Expected Description 'Reviews code and provides feedback', got '%s'", info.Description)
	}
}

func TestCodeReviewSkillExecute(t *testing.T) {
	skill := NewCodeReviewSkill()

	// Test with valid inputs
	input := map[string]interface{}{
		"code":     "func main() { println('hello') }",
		"language": "go",
		"context":  "sample function",
	}

	result, err := skill.Execute(context.TODO(), input)
	if err != nil {
		t.Errorf("Unexpected error when executing skill: %v", err)
	}

	if result == nil {
		t.Error("Expected non-nil result")
	}

	// Verify result contains expected keys
	if _, exists := result["feedback"]; !exists {
		t.Error("Result should contain 'feedback' key")
	}

	if _, exists := result["summary"]; !exists {
		t.Error("Result should contain 'summary' key")
	}

	if _, exists := result["issues"]; !exists {
		t.Error("Result should contain 'issues' key")
	}
}

func TestCodeReviewSkillExecuteWithMissingCode(t *testing.T) {
	skill := NewCodeReviewSkill()

	// Test with missing required parameter
	input := map[string]interface{}{
		"language": "go",
	}

	_, err := skill.Execute(context.TODO(), input)
	if err == nil {
		t.Error("Expected error when 'code' parameter is missing")
	}
}
