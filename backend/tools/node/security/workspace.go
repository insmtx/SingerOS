package security

import (
	"fmt"
	"path/filepath"

	"github.com/insmtx/Leros/backend/pkg/leros"
)

// WorkspaceRoot 获取工作区根目录。
func WorkspaceRoot() (string, error) {
	return leros.WorkspaceRoot()
}

// RealWorkspaceRoot 获取工作区根目录的真实路径（解析符号链接）。
func RealWorkspaceRoot() (string, error) {
	root, err := WorkspaceRoot()
	if err != nil {
		return "", err
	}
	root, err = filepath.Abs(filepath.Clean(root))
	if err != nil {
		return "", fmt.Errorf("resolve workspace root: %w", err)
	}
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("resolve workspace root symlinks: %w", err)
	}
	return filepath.Clean(realRoot), nil
}
