package utils

type Utils interface {
	GenerateRandomPassword(length int) string
	GetFreePort() (port int, err error)
}

var provider Utils

func GenerateRandomPassword(length int) string {
	return provider.GenerateRandomPassword(length)
}

func GetFreePort() (port int, err error) {
	return provider.GetFreePort()
}
