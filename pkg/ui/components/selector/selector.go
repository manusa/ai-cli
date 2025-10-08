package selector

import "github.com/charmbracelet/bubbles/v2/list"

var provider Selector = &implSelector{}

type Selector interface {
	Select(title string, items []list.Item) (string, error)
}

func Select(title string, items []list.Item) (string, error) {
	return provider.Select(title, items)
}
