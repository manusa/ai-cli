package keyring

import "errors"

type mockProvider struct {
	keys map[string]string
}

var _ Keyring = &mockProvider{}

func (k *mockProvider) GetKey(key string) (string, error) {
	if value, ok := k.keys[key]; ok {
		return value, nil
	}
	return "", errors.New("key not found")
}

func (k *mockProvider) SetKey(key, value string) error {
	k.keys[key] = value
	return nil
}

func MockInit() {
	provider = &mockProvider{
		keys: make(map[string]string),
	}
}
