package cmd

import (
	_ "github.com/manusa/ai-cli/pkg/inference/gemini"
	_ "github.com/manusa/ai-cli/pkg/inference/lmstudio"
	_ "github.com/manusa/ai-cli/pkg/inference/ollama"
	_ "github.com/manusa/ai-cli/pkg/inference/ramalama"

	_ "github.com/manusa/ai-cli/pkg/tools/fs"
	_ "github.com/manusa/ai-cli/pkg/tools/github"
	_ "github.com/manusa/ai-cli/pkg/tools/kubernetes"
	_ "github.com/manusa/ai-cli/pkg/tools/playwright"
	_ "github.com/manusa/ai-cli/pkg/tools/postgresql"
)
