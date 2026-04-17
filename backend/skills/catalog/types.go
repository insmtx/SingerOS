package catalog

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Manifest describes the metadata section of a file-based skill.
type Manifest struct {
	Name        string           `yaml:"name"`
	Description string           `yaml:"description"`
	Version     string           `yaml:"version,omitempty"`
	Metadata    ManifestMetadata `yaml:"metadata,omitempty"`
}

// Normalize fills derived defaults after parsing a skill document.
func (m *Manifest) Normalize(defaultName string) {
	if m.Name == "" {
		m.Name = defaultName
	}

	if m.Description == "" {
		m.Description = m.Name
	}
}

// ManifestMetadata stores SingerOS-specific metadata extensions.
type ManifestMetadata struct {
	SingerOS SingerOSMetadata `yaml:"singeros,omitempty"`
}

// SingerOSMetadata stores the first set of skill routing hints used by runtime.
type SingerOSMetadata struct {
	Category      string   `yaml:"category,omitempty"`
	Tags          []string `yaml:"tags,omitempty"`
	Always        bool     `yaml:"always,omitempty"`
	RequiresTools []string `yaml:"requires_tools,omitempty"`
}

// Entry is a discovered skill document with parsed metadata and body.
type Entry struct {
	Manifest Manifest
	Body     string
	Dir      string
	Path     string
}

// Summary is the compact view injected into runtime prompts.
type Summary struct {
	Name          string
	Description   string
	Version       string
	Category      string
	Tags          []string
	Always        bool
	RequiresTools []string
}

// Summary returns the prompt-friendly summary for the skill entry.
func (e *Entry) Summary() Summary {
	return Summary{
		Name:          e.Manifest.Name,
		Description:   e.Manifest.Description,
		Version:       e.Manifest.Version,
		Category:      e.Manifest.Metadata.SingerOS.Category,
		Tags:          e.Manifest.Metadata.SingerOS.Tags,
		Always:        e.Manifest.Metadata.SingerOS.Always,
		RequiresTools: e.Manifest.Metadata.SingerOS.RequiresTools,
	}
}

// ParseDocument parses a SKILL.md document with optional YAML frontmatter.
func ParseDocument(raw []byte) (*Manifest, string, error) {
	manifest := &Manifest{}
	content := strings.TrimSpace(string(raw))
	if content == "" {
		return manifest, "", nil
	}

	if !strings.HasPrefix(content, "---") {
		return manifest, content, nil
	}

	lines := strings.Split(content, "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return nil, "", fmt.Errorf("invalid frontmatter header")
	}

	endIndex := -1
	for idx := 1; idx < len(lines); idx++ {
		if strings.TrimSpace(lines[idx]) == "---" {
			endIndex = idx
			break
		}
	}
	if endIndex == -1 {
		return nil, "", fmt.Errorf("frontmatter closing delimiter not found")
	}

	var yamlBuffer bytes.Buffer
	for idx := 1; idx < endIndex; idx++ {
		yamlBuffer.WriteString(lines[idx])
		yamlBuffer.WriteByte('\n')
	}
	if err := yaml.Unmarshal(yamlBuffer.Bytes(), manifest); err != nil {
		return nil, "", fmt.Errorf("unmarshal frontmatter: %w", err)
	}

	body := strings.Join(lines[endIndex+1:], "\n")
	return manifest, strings.TrimSpace(body), nil
}
