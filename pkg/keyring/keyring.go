package keyring

var provider Keyring = &fallbackKeyring{}

type Keyring interface {
	SetKey(key, value string) error
	GetKey(key string) (string, error)
	DeleteKey(key string) (done bool, err error)
}

func SetKey(key, value string) error {
	return provider.SetKey(key, value)
}

func GetKey(key string) (string, error) {
	return provider.GetKey(key)
}

func DeleteKey(key string) (bool, error) {
	return provider.DeleteKey(key)
}
