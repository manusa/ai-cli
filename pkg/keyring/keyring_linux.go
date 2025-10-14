//go:build linux
// +build linux

package keyring

import (
	"errors"

	"github.com/keybase/go-keychain/secretservice"
)

const (
	service = "com.redhat.ai-cli"
)

type linuxKeychain struct{}

var _ Keyring = &linuxKeychain{}

func (k *linuxKeychain) GetKey(key string) (string, error) {
	srv, err := secretservice.NewService()
	if err != nil {
		return "", err
	}
	collection := secretservice.DefaultCollection
	items, err := srv.SearchCollection(collection, map[string]string{
		"key":     key,
		"service": service,
	})
	if err != nil {
		return "", err
	}
	if len(items) == 0 {
		return "", errors.New("password not found")
	}
	if len(items) > 1 {
		return "", errors.New("more than one password found, should not happen")
	}

	session, err := srv.OpenSession(secretservice.AuthenticationDHAES)
	if err != nil {
		return "", err
	}
	defer srv.CloseSession(session)

	secret, err := srv.GetSecret(items[0], *session)
	if err != nil {
		return "", err
	}
	return string(secret), nil
}

func (k *linuxKeychain) SetKey(key, value string) error {
	srv, err := secretservice.NewService()
	if err != nil {
		return err
	}
	session, err := srv.OpenSession(secretservice.AuthenticationDHAES)
	if err != nil {
		return err
	}
	defer srv.CloseSession(session)

	collection := secretservice.DefaultCollection

	secret, err := session.NewSecret([]byte(value))
	if err != nil {
		return err
	}

	_, err = srv.CreateItem(
		collection,
		secretservice.NewSecretProperties("Gemini API Key for ai-cli", map[string]string{
			"key":     key,
			"service": service,
		}),
		secret,
		secretservice.ReplaceBehaviorReplace,
	)
	return err
}

func (k *linuxKeychain) DeleteKey(key string) (bool, error) {
	srv, err := secretservice.NewService()
	if err != nil {
		return false, err
	}
	collection := secretservice.DefaultCollection
	items, err := srv.SearchCollection(collection, map[string]string{
		"key":     key,
		"service": service,
	})
	if err != nil {
		return false, err
	}
	if len(items) == 0 {
		return false, nil
	}
	if len(items) > 1 {
		return false, errors.New("more than one password found, should not happen")
	}

	err = srv.DeleteItem(items[0])
	if err != nil {
		return false, err
	}
	return true, nil
}

func init() {
	provider = &linuxKeychain{}
}
