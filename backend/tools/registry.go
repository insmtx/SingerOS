package tools

import (
	"fmt"
	"slices"
	"strings"
	"sync"
)

// Registry 提供最小 Tool 注册与获取能力。
type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

// NewRegistry 创建一个新的 Tool 注册表。
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register 注册一个 Tool。
func (r *Registry) Register(tool Tool) error {
	if tool == nil || tool.Info() == nil {
		return fmt.Errorf("tool or tool info is nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	name := tool.Info().Name
	if name == "" {
		return fmt.Errorf("tool name is required")
	}
	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool %s already registered", name)
	}

	r.tools[name] = tool
	return nil
}

// Get 根据名称获取 Tool。
func (r *Registry) Get(name string) (Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("tool %s not found", name)
	}

	return tool, nil
}

// List returns all registered tools sorted by tool name.
func (r *Registry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	slices.Sort(names)

	result := make([]Tool, 0, len(names))
	for _, name := range names {
		result = append(result, r.tools[name])
	}

	return result
}

// ListInfos returns compact tool definitions sorted by name.
func (r *Registry) ListInfos() []ToolInfo {
	tools := r.List()
	infos := make([]ToolInfo, 0, len(tools))
	for _, tool := range tools {
		info := tool.Info()
		if info == nil {
			continue
		}
		infos = append(infos, *info)
	}

	return infos
}

// ListInfosByProvider filters tool definitions by provider.
func (r *Registry) ListInfosByProvider(provider string) []ToolInfo {
	allInfos := r.ListInfos()
	if provider == "" {
		return allInfos
	}

	filtered := make([]ToolInfo, 0, len(allInfos))
	for _, info := range allInfos {
		if strings.EqualFold(info.Provider, provider) {
			filtered = append(filtered, info)
		}
	}

	return filtered
}
