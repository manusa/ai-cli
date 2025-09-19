package main

import (
	"os"

	"github.com/manusa/ai-cli/pkg/cmd"

	"github.com/spf13/pflag"
)

func main() {
	flags := pflag.NewFlagSet("ai-cli", pflag.ExitOnError)
	pflag.CommandLine = flags

	if len(os.Args) == 2 && os.Args[1] == "--version" {
		os.Args[1] = "version" // Normalize the version flag to match the sub-command
	}

	root := cmd.NewAiCli()
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
