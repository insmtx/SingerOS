package main

import (
	"fmt"

	"github.com/insmtx/SingerOS/backend/agent/react"
	"github.com/insmtx/SingerOS/backend/skills"
	code_review_skill "github.com/insmtx/SingerOS/backend/skills/tool_skills/code_review_skill"
	"github.com/ygpkg/yg-go/logs"
)

// RegisterProgrammingSkills registers skills specific for the programmer assistant
func RegisterProgrammingSkills(skillManager skills.SkillManager) error {
	// Create and register the code review skill
	codeReviewSkill := code_review_skill.NewCodeReviewSkill()
	if err := skillManager.Register(codeReviewSkill); err != nil {
		return fmt.Errorf("failed to register code review skill: %w", err)
	}

	// Create and register the PR analysis skill
	prAnalysisSkill := react.NewPRAnalysisSkill()
	if err := skillManager.Register(prAnalysisSkill); err != nil {
		return fmt.Errorf("failed to register PR analysis skill: %w", err)
	}

	// Additional programming-related skills could be registered here
	// Such as code suggestion, documentation search, etc.

	logs.Info("Programming skills registered successfully")
	return nil
}
