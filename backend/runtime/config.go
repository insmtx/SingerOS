package runtime

import (
	skillcatalog "github.com/insmtx/SingerOS/backend/skills/catalog"
	"github.com/insmtx/SingerOS/backend/toolruntime"
	"github.com/insmtx/SingerOS/backend/tools"
)

// Config stores runtime dependencies that are orthogonal to the agent implementation.
type Config struct {
	SkillsCatalog *skillcatalog.Catalog
	ToolRegistry  *tools.Registry
	ToolRuntime   *toolruntime.Runtime
}
