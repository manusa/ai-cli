package context

import "github.com/manusa/ai-cli/pkg/ai"

type ModelContext struct {
	Ai      *ai.Ai
	Width   int
	Height  int
	Version string
}
