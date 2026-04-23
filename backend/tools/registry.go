package tools

import (
	"fmt"
	"slices"
	"sync"

	"github.com/ygpkg/yg-go/logs"
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

// Register 注册一个 Tool，同名 Tool 会被新实例覆盖。
func (r *Registry) Register(tool Tool) error {
	if tool == nil {
		return fmt.Errorf("tool is nil")
	}

	name := tool.Name()
	if name == "" {
		return fmt.Errorf("tool name is required")
	}

	r.mu.Lock()
	previous, exists := r.tools[name]
	r.tools[name] = tool
	r.mu.Unlock()

	if exists {
		logs.Infof("Overwrote tool registration: name=%s previous=%T current=%T", name, previous, tool)
	} else {
		logs.Infof("Registered tool: name=%s current=%T", name, tool)
	}

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
