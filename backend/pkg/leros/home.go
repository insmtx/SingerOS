// Package leros provides shared Leros filesystem conventions.
package leros

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	// EnvWorkspaceRoot is the worker-local root used for Leros state.
	EnvWorkspaceRoot = "LEROS_WORKSPACE_ROOT"

	defaultWorkspaceRootUnix = "/workspace"
	defaultWindowsAppName    = "Leros"
	defaultWorkspaceDirName  = "workspace"
)

// WorkspaceRoot returns $LEROS_WORKSPACE_ROOT, or the platform default workspace root when unset.
func WorkspaceRoot() (string, error) {
	root := strings.TrimSpace(os.Getenv(EnvWorkspaceRoot))
	if root == "" {
		var err error
		root, err = defaultWorkspaceRoot(runtime.GOOS)
		if err != nil {
			return "", err
		}
	}

	absolute, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve %s: %w", EnvWorkspaceRoot, err)
	}
	return absolute, nil
}

// JoinWorkspace joins path elements under the Leros workspace root.
func JoinWorkspace(elem ...string) (string, error) {
	root, err := WorkspaceRoot()
	if err != nil {
		return "", err
	}
	parts := append([]string{root}, elem...)
	return filepath.Join(parts...), nil
}

// SkillsDir returns the default Leros skills directory.
func SkillsDir() (string, error) {
	return JoinWorkspace("skills")
}

// MemoryDir returns the default Leros memory directory.
func MemoryDir() (string, error) {
	return JoinWorkspace("memory")
}

func defaultWorkspaceRoot(goos string) (string, error) {
	if goos != "windows" {
		return defaultWorkspaceRootUnix, nil
	}

	localAppData := strings.TrimSpace(os.Getenv("LOCALAPPDATA"))
	if localAppData == "" {
		userCacheDir, err := os.UserCacheDir()
		if err != nil {
			return "", fmt.Errorf("resolve Windows local application data: %w", err)
		}
		localAppData = strings.TrimSpace(userCacheDir)
	}
	if localAppData == "" {
		return "", fmt.Errorf("Windows local application data is empty")
	}
	return filepath.Join(localAppData, defaultWindowsAppName, defaultWorkspaceDirName), nil
}
