package catalog

import (
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"
)

const skillFileName = "SKILL.md"

// Catalog stores discovered file-based skills for runtime prompt assembly.
type Catalog struct {
	fs      fs.FS
	entries map[string]*Entry
}

// New creates a catalog by scanning the provided filesystem for SKILL.md files.
func New(skillFS fs.FS) (*Catalog, error) {
	entries := make(map[string]*Entry)

	err := fs.WalkDir(skillFS, ".", func(filePath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || path.Base(filePath) != skillFileName {
			return nil
		}

		raw, err := fs.ReadFile(skillFS, filePath)
		if err != nil {
			return fmt.Errorf("read skill file %s: %w", filePath, err)
		}

		manifest, body, err := ParseDocument(raw)
		if err != nil {
			return fmt.Errorf("parse skill file %s: %w", filePath, err)
		}

		skillDir := path.Dir(filePath)
		if skillDir == "." {
			skillDir = ""
		}
		manifest.Normalize(path.Base(skillDir))

		entry := &Entry{
			Manifest: *manifest,
			Body:     body,
			Dir:      skillDir,
			Path:     filePath,
		}
		if _, exists := entries[entry.Manifest.Name]; exists {
			return fmt.Errorf("duplicate skill name %q", entry.Manifest.Name)
		}
		entries[entry.Manifest.Name] = entry
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Catalog{
		fs:      skillFS,
		entries: entries,
	}, nil
}

// List returns skill summaries sorted by name.
func (c *Catalog) List() []Summary {
	if c == nil {
		return nil
	}

	summaries := make([]Summary, 0, len(c.entries))
	for _, entry := range c.entries {
		summaries = append(summaries, entry.Summary())
	}

	slices.SortFunc(summaries, func(left, right Summary) int {
		return strings.Compare(left.Name, right.Name)
	})

	return summaries
}

// Get returns a full skill entry by name.
func (c *Catalog) Get(name string) (*Entry, error) {
	if c == nil {
		return nil, fmt.Errorf("catalog is nil")
	}

	entry, ok := c.entries[name]
	if !ok {
		return nil, fmt.Errorf("skill %q not found", name)
	}

	return entry, nil
}

// ReadFile reads an additional file under the skill directory.
func (c *Catalog) ReadFile(name string, relativePath string) ([]byte, error) {
	entry, err := c.Get(name)
	if err != nil {
		return nil, err
	}

	cleanPath := path.Clean(relativePath)
	if cleanPath == "." || strings.HasPrefix(cleanPath, "../") || path.IsAbs(cleanPath) {
		return nil, fmt.Errorf("invalid skill file path %q", relativePath)
	}

	fullPath := cleanPath
	if entry.Dir != "" {
		fullPath = path.Join(entry.Dir, cleanPath)
	}

	content, err := fs.ReadFile(c.fs, fullPath)
	if err != nil {
		return nil, fmt.Errorf("read skill file %s: %w", fullPath, err)
	}

	return content, nil
}
