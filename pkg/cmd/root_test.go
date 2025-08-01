package cmd

import (
	"io"
	"os"
	"testing"
)

func captureOutput(f func() error) (string, error) {
	originalOut := os.Stdout
	defer func() {
		os.Stdout = originalOut
	}()
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := f()
	_ = w.Close()
	out, _ := io.ReadAll(r)
	return string(out), err
}

func TestVersion(t *testing.T) {
	rootCmd := NewAiCli()
	rootCmd.SetArgs([]string{"version"})
	o, err := captureOutput(rootCmd.Execute)
	if o != "0.0.0\n" {
		t.Fatalf("Expected version for command 'ai-cli version', got %s %v", o, err)
	}
}
