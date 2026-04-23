// Package skilltools provides runtime tools for discovering and loading SingerOS skills.
package skilltools

import (
	"context"
	"fmt"
	"html"
	"strings"
	"unicode/utf8"

	"github.com/insmtx/SingerOS/backend/tools"
)

const (
	// ProviderSkill identifies tools that operate on SingerOS skill documents.
	ProviderSkill = "skill"

	// ToolNameSkillUse is the runtime tool used to discover and load skill documents.
	ToolNameSkillUse = "skill_use"
)

const (
	actionList     = "list"
	actionGet      = "get"
	actionReadFile = "read_file"
)

const (
	defaultSkillFileListLimit = 10
	maxSkillFileReadBytes     = 128 * 1024
)

// SkillCatalog is the read-only catalog surface required by SkillUseTool.
type SkillCatalog interface {
	List() []Summary
	Get(name string) (*Entry, error)
	ReadFile(name string, relativePath string) ([]byte, error)
	ListFiles(name string, limit int) ([]string, error)
}

// SkillUseTool lets an agent query and load skills from the runtime skill catalog.
type SkillUseTool struct {
	catalog SkillCatalog
}

// NewSkillUseTool creates a catalog-backed skill use tool.
func NewSkillUseTool(catalog SkillCatalog) *SkillUseTool {
	return &SkillUseTool{catalog: catalog}
}

// Info returns metadata for the skill use tool.
func (t *SkillUseTool) Info() *tools.ToolInfo {
	return &tools.ToolInfo{
		Name: ToolNameSkillUse,
		Description: strings.Join([]string{
			"管理和使用技能（Skill）。",
			"支持 list 列出所有可用技能，get 获取指定技能完整内容和可注入上下文，read_file 读取技能目录下的附加文件。",
			"当任务需要查看、选择或加载技能说明时调用此工具。",
		}, ""),
		Provider: ProviderSkill,
		ReadOnly: true,
		InputSchema: &tools.Schema{
			Type:     "object",
			Required: []string{"action"},
			Properties: map[string]*tools.Property{
				"action": {
					Type:        "string",
					Enum:        []string{actionList, actionGet, actionReadFile},
					Description: "操作类型：list 列出技能，get 获取技能正文，read_file 读取技能目录下的文件",
				},
				"name": {
					Type:        "string",
					Description: "技能名称，get 和 read_file 时必填",
				},
				"path": {
					Type:        "string",
					Description: "技能目录内的相对文件路径，read_file 时必填",
				},
			},
		},
	}
}

// Validate checks skill use tool input.
func (t *SkillUseTool) Validate(input map[string]interface{}) error {
	if input == nil {
		return fmt.Errorf("input is required")
	}

	action := stringValue(input, "action")
	switch action {
	case actionList:
		return nil
	case actionGet:
		if stringValue(input, "name") == "" {
			return fmt.Errorf("name is required")
		}
		return nil
	case actionReadFile:
		if stringValue(input, "name") == "" {
			return fmt.Errorf("name is required")
		}
		if stringValue(input, "path") == "" {
			return fmt.Errorf("path is required")
		}
		return nil
	case "":
		return fmt.Errorf("action is required")
	default:
		return fmt.Errorf("unsupported action %q", action)
	}
}

// Execute performs the requested skill catalog action.
func (t *SkillUseTool) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	if err := t.Validate(input); err != nil {
		return nil, err
	}
	if t == nil || t.catalog == nil {
		return nil, fmt.Errorf("skill catalog is required")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	switch stringValue(input, "action") {
	case actionList:
		return t.listSkills(), nil
	case actionGet:
		return t.getSkill(stringValue(input, "name")), nil
	case actionReadFile:
		return t.readSkillFile(stringValue(input, "name"), stringValue(input, "path")), nil
	default:
		return nil, fmt.Errorf("unsupported action %q", stringValue(input, "action"))
	}
}

func (t *SkillUseTool) listSkills() map[string]interface{} {
	summaries := t.catalog.List()
	skills := make([]map[string]interface{}, 0, len(summaries))
	for _, summary := range summaries {
		skills = append(skills, summaryMap(summary))
	}

	return map[string]interface{}{
		"ok":     true,
		"count":  len(skills),
		"skills": skills,
	}
}

func (t *SkillUseTool) getSkill(name string) map[string]interface{} {
	entry, err := t.findSkill(name)
	if err != nil {
		return skillNotFound(name, t.catalog.List())
	}

	files, err := t.catalog.ListFiles(entry.Manifest.Name, defaultSkillFileListLimit)
	if err != nil {
		return map[string]interface{}{
			"ok":      false,
			"message": err.Error(),
		}
	}

	output := formatSkillContent(entry, files)
	metadata := map[string]interface{}{
		"name":            entry.Manifest.Name,
		"dir":             entry.Dir,
		"path":            entry.Path,
		"files":           files,
		"file_list_limit": defaultSkillFileListLimit,
	}

	return map[string]interface{}{
		"ok":       true,
		"title":    fmt.Sprintf("Loaded skill: %s", entry.Manifest.Name),
		"output":   output,
		"metadata": metadata,
		"skill": map[string]interface{}{
			"name":           entry.Manifest.Name,
			"description":    entry.Manifest.Description,
			"version":        entry.Manifest.Version,
			"category":       entry.Manifest.Metadata.SingerOS.Category,
			"tags":           entry.Manifest.Metadata.SingerOS.Tags,
			"always":         entry.Manifest.Metadata.SingerOS.Always,
			"requires_tools": entry.Manifest.Metadata.SingerOS.RequiresTools,
			"dir":            entry.Dir,
			"path":           entry.Path,
			"scope":          "catalog",
			"skill_type":     "file",
			"enabled":        true,
			"files":          files,
			"body":           entry.Body,
			"content":        entry.Body,
		},
	}
}

func (t *SkillUseTool) readSkillFile(name string, relativePath string) map[string]interface{} {
	entry, err := t.findSkill(name)
	if err != nil {
		return skillNotFound(name, t.catalog.List())
	}

	content, err := t.catalog.ReadFile(entry.Manifest.Name, relativePath)
	if err != nil {
		return map[string]interface{}{
			"ok":      false,
			"message": err.Error(),
		}
	}

	displayContent, truncated := truncateFileContent(content, maxSkillFileReadBytes)
	return map[string]interface{}{
		"ok":              true,
		"name":            entry.Manifest.Name,
		"path":            relativePath,
		"content":         displayContent,
		"size":            len(content),
		"truncated":       truncated,
		"max_read_bytes":  maxSkillFileReadBytes,
		"original_length": len(content),
	}
}

func summaryMap(summary Summary) map[string]interface{} {
	return map[string]interface{}{
		"name":           summary.Name,
		"description":    summary.Description,
		"version":        summary.Version,
		"category":       summary.Category,
		"tags":           summary.Tags,
		"always":         summary.Always,
		"requires_tools": summary.RequiresTools,
		"scope":          "catalog",
		"skill_type":     "file",
		"enabled":        true,
	}
}

func (t *SkillUseTool) findSkill(name string) (*Entry, error) {
	entry, err := t.catalog.Get(name)
	if err == nil {
		return entry, nil
	}

	for _, summary := range t.catalog.List() {
		if !strings.EqualFold(summary.Name, name) {
			continue
		}
		return t.catalog.Get(summary.Name)
	}

	return nil, err
}

func formatSkillContent(entry *Entry, files []string) string {
	var builder strings.Builder
	skillName := entry.Manifest.Name
	baseDir := entry.Dir
	if baseDir == "" {
		baseDir = "."
	}

	builder.WriteString(`<skill_content name="`)
	builder.WriteString(html.EscapeString(skillName))
	builder.WriteString(`">`)
	builder.WriteString("\n")
	builder.WriteString("# Skill: ")
	builder.WriteString(skillName)
	builder.WriteString("\n\n")
	builder.WriteString(strings.TrimSpace(entry.Body))
	builder.WriteString("\n\n")
	builder.WriteString("Base directory for this skill: ")
	builder.WriteString(baseDir)
	builder.WriteString("\n")
	builder.WriteString("Relative paths in this skill are relative to this base directory.")
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("File list is sampled up to %d files.", defaultSkillFileListLimit))
	builder.WriteString("\n\n")
	builder.WriteString("<skill_files>")
	builder.WriteString("\n")
	builder.WriteString(strings.Join(files, "\n"))
	builder.WriteString("\n")
	builder.WriteString("</skill_files>")
	builder.WriteString("\n")
	builder.WriteString("</skill_content>")

	return builder.String()
}

func truncateFileContent(content []byte, maxBytes int) (string, bool) {
	if maxBytes <= 0 || len(content) <= maxBytes {
		return string(content), false
	}

	truncated := content[:maxBytes]
	for len(truncated) > 0 && !utf8.Valid(truncated) {
		truncated = truncated[:len(truncated)-1]
	}

	return string(truncated), true
}

func skillNotFound(name string, summaries []Summary) map[string]interface{} {
	available := make([]string, 0, len(summaries))
	for _, summary := range summaries {
		available = append(available, summary.Name)
	}

	return map[string]interface{}{
		"ok":        false,
		"message":   fmt.Sprintf("skill %q not found", name),
		"available": available,
	}
}

func stringValue(input map[string]interface{}, key string) string {
	value, _ := input[key].(string)
	return strings.TrimSpace(value)
}
