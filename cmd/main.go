package main

import (
	"github.com/manusa/ai-cli/pkg/cmd"
	"os"

	"github.com/spf13/pflag"
)

func main() {
	flags := pflag.NewFlagSet("kubernetes-mcp-server", pflag.ExitOnError)
	pflag.CommandLine = flags

	root := cmd.NewAiCli()
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
