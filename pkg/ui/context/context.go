package context

import (
	"github.com/manusa/ai-cli/pkg/ai"
	"github.com/manusa/ai-cli/pkg/ui/styles"
)

type ModelContext struct {
	Ai      *ai.Ai
	Theme   *styles.Theme
	Width   int
	Height  int
	Version string
}
