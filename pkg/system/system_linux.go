//go:build linux
// +build linux

package system

import "os/exec"

type linuxSystem struct{}

var _ System = &linuxSystem{}

func (k *linuxSystem) OpenBrowser(url string) error {
	return exec.Command("xdg-open", url).Run()
}

func init() {
	provider = &linuxSystem{}
}
