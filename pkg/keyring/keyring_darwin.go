//go:build darwin
// +build darwin

package keyring

import "github.com/keybase/go-keychain"

const (
	service = "com.redhat.ai-cli"
)

type macOSXKeychain struct{}

var _ Keyring = &macOSXKeychain{}

func (k *macOSXKeychain) GetKey(key string) (string, error) {
	password, err := keychain.GetGenericPassword(service, key, "", "")
	return string(password), err
}

func (k *macOSXKeychain) SetKey(key, value string) error {
	err := keychain.DeleteGenericPasswordItem(service, key)
	if err != keychain.ErrorItemNotFound {
		return err
	}

	item := keychain.NewGenericPassword(service, key, "", []byte(value), "")
	item.SetSynchronizable(keychain.SynchronizableNo)
	item.SetAccessible(keychain.AccessibleWhenUnlocked)
	return keychain.AddItem(item)
}

func (k *macOSXKeychain) DeleteKey(key string) (done bool, err error) {
	password, err := keychain.GetGenericPassword(service, key, "", "")
	if err != nil {
		return false, err
	}
	if password == nil {
		return false, nil
	}
	return true, keychain.DeleteGenericPasswordItem(service, key)
}

func init() {
	provider = &macOSXKeychain{}
}
