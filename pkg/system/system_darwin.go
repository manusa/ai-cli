//go:build darwin
// +build darwin

package system

import "os/exec"

type darwinSystem struct{}

var _ System = &darwinSystem{}

func (k *darwinSystem) OpenBrowser(url string) error {
	return exec.Command("open", url).Run()
}

func init() {
	provider = &darwinSystem{}
}
