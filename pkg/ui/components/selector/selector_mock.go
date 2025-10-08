package selector

import "github.com/charmbracelet/bubbles/v2/list"

type mockProvider struct {
	selectMock func(title string, items []list.Item) (string, error)
}

var _ Selector = &mockProvider{}

func (m *mockProvider) Select(title string, items []list.Item) (string, error) {
	return m.selectMock(title, items)
}

func MockInit(selectMock func(title string, items []list.Item) (string, error)) {
	provider = &mockProvider{
		selectMock: selectMock,
	}
}
