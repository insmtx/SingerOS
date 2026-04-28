package engines

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ygpkg/yg-go/logs"
)

var defaultBuiltinSkillDirs = []string{
	"./backend/skills",
	"/app/backend/skills",
}

var defaultSkillTargetDirs = []string{
	"~/.claude/skills",
	"~/.agents/skills",
}

const skillManifestFile = "SKILL.md"

// SyncBuiltinSkills copies SingerOS built-in skills into external agent skill directories.
func SyncBuiltinSkills(sourceDir string, targetDirs []string) error {
	sourceDir, err := resolveBuiltinSkillsSource(sourceDir)
	if err != nil {
		return err
	}
	if len(targetDirs) == 0 {
		targetDirs = defaultSkillTargetDirs
	}

	for _, targetDir := range targetDirs {
		targetDir, err := expandPath(targetDir)
		if err != nil {
			return err
		}
		if err := syncSkillDir(sourceDir, targetDir); err != nil {
			return err
		}
		logs.Infof("Synced built-in skills from %s to %s", sourceDir, targetDir)
	}
	return nil
}

func resolveBuiltinSkillsSource(sourceDir string) (string, error) {
	candidates := defaultBuiltinSkillDirs
	if strings.TrimSpace(sourceDir) != "" {
		candidates = append([]string{sourceDir}, candidates...)
	}
	if configured := strings.TrimSpace(os.Getenv("SINGEROS_SKILLS_DIR")); configured != "" {
		candidates = append([]string{configured}, candidates...)
	}

	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("built-in skills directory not found")
}

func syncSkillDir(sourceDir string, targetDir string) error {
	skillDirs, err := listSkillDirs(sourceDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return err
	}

	for _, skillDir := range skillDirs {
		if err := syncSingleSkillDir(sourceDir, targetDir, skillDir); err != nil {
			return err
		}
	}
	return nil
}

func listSkillDirs(sourceDir string) ([]string, error) {
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return nil, err
	}

	var skillDirs []string
	for _, entry := range entries {
		if !entry.IsDir() {
			return nil, fmt.Errorf("invalid skill source entry %s: expected skill directory", filepath.Join(sourceDir, entry.Name()))
		}
		manifestPath := filepath.Join(sourceDir, entry.Name(), skillManifestFile)
		info, err := os.Stat(manifestPath)
		if err != nil {
			return nil, fmt.Errorf("invalid skill directory %s: missing %s", filepath.Join(sourceDir, entry.Name()), skillManifestFile)
		}
		if info.IsDir() {
			return nil, fmt.Errorf("invalid skill directory %s: %s must be a file", filepath.Join(sourceDir, entry.Name()), skillManifestFile)
		}
		skillDirs = append(skillDirs, entry.Name())
	}
	if len(skillDirs) == 0 {
		return nil, fmt.Errorf("no skill directories found in %s", sourceDir)
	}
	return skillDirs, nil
}

func syncSingleSkillDir(sourceDir string, targetDir string, skillDir string) error {
	skillSourceDir := filepath.Join(sourceDir, skillDir)
	return filepath.WalkDir(skillSourceDir, func(sourcePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relPath, err := filepath.Rel(sourceDir, sourcePath)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(targetDir, relPath)
		if entry.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}
		return copyFileIfChanged(sourcePath, targetPath, entry)
	})
}

func copyFileIfChanged(sourcePath string, targetPath string, entry fs.DirEntry) error {
	info, err := entry.Info()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("unsupported non-regular skill file %s", sourcePath)
	}

	same, err := sameFileContent(sourcePath, targetPath)
	if err != nil {
		return err
	}
	if same {
		return os.Chmod(targetPath, info.Mode().Perm())
	}

	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}

	target, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer target.Close()

	if _, err := io.Copy(target, source); err != nil {
		return err
	}
	return target.Chmod(info.Mode().Perm())
}

func sameFileContent(sourcePath string, targetPath string) (bool, error) {
	targetInfo, err := os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if targetInfo.IsDir() {
		return false, fmt.Errorf("target path %s is a directory", targetPath)
	}

	sourceHash, err := fileSHA256(sourcePath)
	if err != nil {
		return false, err
	}
	targetHash, err := fileSHA256(targetPath)
	if err != nil {
		return false, err
	}
	return bytes.Equal(sourceHash, targetHash), nil
}

func fileSHA256(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hashValue := sha256.New()
	if _, err := io.Copy(hashValue, file); err != nil {
		return nil, err
	}
	return hashValue.Sum(nil), nil
}

func expandPath(pathValue string) (string, error) {
	pathValue = strings.TrimSpace(pathValue)
	if pathValue == "" {
		return "", fmt.Errorf("path is required")
	}
	if pathValue == "~" || strings.HasPrefix(pathValue, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if pathValue == "~" {
			return home, nil
		}
		return filepath.Join(home, strings.TrimPrefix(pathValue, "~/")), nil
	}
	return pathValue, nil
}
