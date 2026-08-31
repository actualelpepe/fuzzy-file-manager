package main

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type State int

const (
	Loading State = iota
	Normal
	Searching
)

type FileManager struct {
	selector  FileSelector
	searchBar TextInput
	spinner   Spinner
	state     State
	winSize   tea.WindowSizeMsg
}

func NewFileManager(path string) (FileManager, error) {
	selector, err := NewFileSelector(path)
	if err != nil {
		return FileManager{}, err
	}
	searchBar := NewTextInput()
	searchBar.Prompt = "/"
	return FileManager{
		selector:  selector,
		spinner:   NewSpinner(),
		state:     Loading,
		searchBar: searchBar,
	}, nil

}

func (m FileManager) Init() tea.Cmd {
	return tea.Batch(m.selector.Init(), m.spinner.Init(), m.searchBar.Init())
}

func (m FileManager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case Loading:
		return m.updateLoading(msg)
	case Normal:
		return m.updateNormal(msg)
	case Searching:
		return m.updateSearching(msg)
	}
	return m, nil
}

func (m FileManager) updateLoading(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		}
	case SpinnerTickMsg:
		spinner, cmd := m.spinner.Update(msg)
		m.spinner = spinner.(Spinner)
		return m, cmd
	case DirWalkMsg:
		m.state = Normal
		selector, cmd := m.selector.Update(msg)
		m.selector = selector.(FileSelector)
		return m, tea.Batch(tea.RequestWindowSize, cmd)
	}
	return m, nil
}

func (m FileManager) updateNormal(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "f", "/":
			m.state = Searching
			return m, tea.RequestWindowSize
		case "F":
			m.selector.Filter = ""
			m.searchBar.Content = ""
			selector, cmd := m.selector.Update(RequestVisibleFilesUpdate())
			m.selector = selector.(FileSelector)
			return m, cmd
		}
		selector, cmd := m.selector.Update(msg)
		m.selector = selector.(FileSelector)
		return m, cmd
	case tea.WindowSizeMsg:
		m.winSize = msg
		m.selector.ViewSize = ViewSize{msg.Width, msg.Height}
		selector, cmd := m.selector.Update(RequestVisibleFilesUpdate())
		m.selector = selector.(FileSelector)
		return m, cmd
	}
	return m, nil
}

func (m FileManager) updateSearching(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "enter":
			m.state = Normal
			return m, tea.RequestWindowSize
		}
		searchBar, searchBarCmd := m.searchBar.Update(msg)
		m.searchBar = searchBar.(TextInput)
		m.selector.Filter = m.searchBar.Content
		selector, selectorCmd := m.selector.Update(RequestVisibleFilesUpdate())
		m.selector = selector.(FileSelector)
		return m, tea.Batch(searchBarCmd, selectorCmd)
	case tea.WindowSizeMsg:
		m.winSize = msg
		m.selector.ViewSize = ViewSize{msg.Width, msg.Height - 1}
		selector, cmd := m.selector.Update(RequestVisibleFilesUpdate())
		m.selector = selector.(FileSelector)
		return m, cmd
	}
	return m, nil
}

func (m FileManager) View() tea.View {
	switch m.state {
	case Loading:
		return m.spinner.View()
	case Normal:
		return m.selector.View()
	case Searching:
		return m.viewSearching()
	}
	invalidView := tea.NewView("Invalid program state. There is nothing you can do ¯\\_(ツ)_/¯")
	invalidView.AltScreen = true
	return invalidView
}

func (m FileManager) viewSearching() tea.View {
	selectorStyle := lipgloss.NewStyle().MaxHeight(m.winSize.Height - 1).Height(m.winSize.Height - 1).SetString(m.selector.View().Content)
	view := tea.NewView(selectorStyle.String() + "\n" + m.searchBar.View().Content)
	view.AltScreen = true
	return view
}
