package skilltools

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"slices"
	"strings"
)

const skillFileName = "SKILL.md"

var defaultSkillDirs = []string{
	"./backend/skills",
	"/app/backend/skills",
}

// Catalog stores discovered file-based skills for runtime prompt assembly.
type Catalog struct {
	fs      fs.FS
	entries map[string]*Entry
}

// LoadDefaultCatalog loads skills from the default SingerOS skill directory.
func LoadDefaultCatalog() (*Catalog, string, error) {
	candidates := defaultSkillDirs
	if configured := strings.TrimSpace(os.Getenv("SINGEROS_SKILLS_DIR")); configured != "" {
		candidates = append([]string{configured}, candidates...)
	}

	var lastErr error
	for _, dir := range candidates {
		if _, err := os.Stat(dir); err != nil {
			lastErr = err
			continue
		}

		catalog, err := NewCatalog(os.DirFS(dir))
		if err != nil {
			lastErr = err
			continue
		}

		return catalog, dir, nil
	}

	if lastErr != nil {
		return nil, "", fmt.Errorf("load skills from default directories: %w", lastErr)
	}
	return nil, "", fmt.Errorf("load skills from default directories: no candidates configured")
}

// NewCatalog creates a catalog by scanning the provided filesystem for SKILL.md files.
func NewCatalog(skillFS fs.FS) (*Catalog, error) {
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

// ListFiles returns additional files under the skill directory, excluding SKILL.md.
func (c *Catalog) ListFiles(name string, limit int) ([]string, error) {
	entry, err := c.Get(name)
	if err != nil {
		return nil, err
	}

	root := entry.Dir
	if root == "" {
		root = "."
	}

	files := make([]string, 0)
	err = fs.WalkDir(c.fs, root, func(filePath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if path.Base(filePath) == skillFileName {
			return nil
		}

		relativePath := filePath
		if entry.Dir != "" {
			relativePath = strings.TrimPrefix(filePath, entry.Dir+"/")
		}
		files = append(files, relativePath)
		if limit > 0 && len(files) >= limit {
			return fs.SkipAll
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list skill files %s: %w", name, err)
	}

	slices.Sort(files)
	return files, nil
}
