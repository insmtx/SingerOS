package bundled

import "embed"

// FS embeds bundled knowledge skills so runtime does not depend on the process cwd.
//
//go:embed *
var FS embed.FS
