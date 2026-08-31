package main

import (
	tea "charm.land/bubbletea/v2"
)

type FileManager struct {
	selector FileSelector
}

func NewFileManager(path string) (FileManager, error) {
	selector, err := NewFileSelector(path)
	if err != nil {
		return FileManager{}, err
	}
	return FileManager{selector: selector}, nil

}

func (m FileManager) Init() tea.Cmd {
	return m.selector.Init()
}

func (m FileManager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		}
	}
	selector, cmd := m.selector.Update(msg)
	m.selector = selector.(FileSelector)
	return m, cmd
}

func (m FileManager) View() tea.View {
	return m.selector.View()
}
