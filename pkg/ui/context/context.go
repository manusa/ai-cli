package context

import (
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/ui/styles"
)

type ModelContext struct {
	Ai      api.Ai
	Theme   *styles.Theme
	Width   int
	Height  int
	Version string
}
