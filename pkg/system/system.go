package system

var provider System = &fallbackSystem{}

type System interface {
	OpenBrowser(url string) error
}

func OpenBrowser(url string) error {
	return provider.OpenBrowser(url)
}
