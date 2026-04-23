package skilltools

import (
	"fmt"

	"github.com/insmtx/SingerOS/backend/tools"
)

// NewTools returns all skill catalog tools for registration.
func NewTools(catalog SkillCatalog) []tools.Tool {
	return []tools.Tool{
		NewSkillUseTool(catalog),
	}
}

// Register registers all skill catalog tools into the provided registry.
func Register(registry *tools.Registry, catalog SkillCatalog) error {
	if registry == nil {
		return fmt.Errorf("tool registry is required")
	}

	for _, tool := range NewTools(catalog) {
		if err := registry.Register(tool); err != nil {
			return err
		}
	}

	return nil
}
