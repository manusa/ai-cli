package keyring

import "errors"

type fallbackKeyring struct{}

var _ Keyring = &fallbackKeyring{}

func (k *fallbackKeyring) GetKey(key string) (string, error) {
	return "", errors.New("not implemented")
}

func (k *fallbackKeyring) SetKey(key, value string) error {
	return errors.New("not implemented")
}
