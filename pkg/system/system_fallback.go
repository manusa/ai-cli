package system

import "errors"

type fallbackSystem struct{}

var _ System = &fallbackSystem{}

func (k *fallbackSystem) OpenBrowser(url string) error {
	return errors.New("not implemented")
}
