package keyring

import "errors"

type mockProvider struct{}

var _ Keyring = &mockProvider{}

func (k *mockProvider) GetKey(key string) (string, error) {
	return "", errors.New("not implemented")
}

func (k *mockProvider) SetKey(key, value string) error {
	return errors.New("not implemented")
}

func MockInit() {
	provider = &mockProvider{}
}
