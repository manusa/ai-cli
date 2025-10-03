package keyring

import "github.com/keybase/go-keychain"

const (
	service = "com.redhat.ai-cli"
)

func GetKey(key string) (string, error) {
	password, err := keychain.GetGenericPassword(service, key, "", "")
	return string(password), err
}

func SetKey(key, value string) error {
	item := keychain.NewGenericPassword(service, key, "", []byte(value), "")
	item.SetSynchronizable(keychain.SynchronizableNo)
	item.SetAccessible(keychain.AccessibleWhenUnlocked)
	return keychain.AddItem(item)
}
