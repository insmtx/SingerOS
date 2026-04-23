package runtime

import (
	"github.com/insmtx/SingerOS/backend/tools"
	skilltools "github.com/insmtx/SingerOS/backend/tools/skill"
)

// Config stores runtime dependencies that are orthogonal to the agent implementation.
type Config struct {
	SkillsCatalog *skilltools.Catalog
	ToolRegistry  *tools.Registry
}
