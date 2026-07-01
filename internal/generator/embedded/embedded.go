// Package embedded holds the fallback copies of every DRP template,
// embedded into the binary at compile time. The render engine checks for
// user-overridable templates on disk first (./templates/*.tpl) and falls
// back to these when none are found.
package embedded

import "embed"

//go:embed files/*.tpl
var FS embed.FS
