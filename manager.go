package main

import (
	"os"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type State int

const (
	Loading State = iota
	Normal
	Search
	Delete
	Message
)

type FileManager struct {
	selector  FileSelector
	searchBar TextInput
	spinner   Spinner
	state     State
	winSize   tea.WindowSizeMsg
	message   string
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
	case Search:
		return m.updateSearch(msg)
	case Delete:
		return m.updateDelete(msg)
	case Message:
		return m.updateMessage(msg)
	}
	return m, nil
}

func (m FileManager) updateLoading(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.Exit):
			return m, tea.Quit
		}
	case SpinnerTickMsg:
		spinner, cmd := m.spinner.Update(msg)
		m.spinner = spinner.(Spinner)
		return m, cmd
	case Files:
		selector, _ := m.selector.Update(msg)
		m.selector = selector.(FileSelector)
		m.state = Normal
		return m, tea.RequestWindowSize
	}
	return m, nil
}

func (m FileManager) updateNormal(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.Exit):
			return m, tea.Quit
		case key.Matches(msg, DefaultKeyMap.Delete):
			m.state = Delete
			return m, tea.RequestWindowSize
		case key.Matches(msg, DefaultKeyMap.Refresh):
			m.state = Loading
			return m, m.selector.RefreshFiles
		case key.Matches(msg, DefaultKeyMap.Search):
			m.state = Search
			return m, tea.RequestWindowSize
		case key.Matches(msg, DefaultKeyMap.ResetSearch):
			m.searchBar.Content = ""
			m.selector = m.selector.ResetFilter()
			return m, nil
		}
		selector, cmd := m.selector.Update(msg)
		m.selector = selector.(FileSelector)
		return m, cmd
	case tea.WindowSizeMsg:
		m.winSize = msg
		m.selector.ViewSize = msg
		return m, nil
	}
	return m, nil
}

func (m FileManager) updateSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.CancelSearch):
			m.searchBar.Content = ""
			m.selector = m.selector.ResetFilter()
			fallthrough
		case key.Matches(msg, DefaultKeyMap.ConfirmSearch):
			m.state = Normal
			return m, tea.RequestWindowSize
		}
		searchBar, _ := m.searchBar.Update(msg)
		m.searchBar = searchBar.(TextInput)
		m.selector = m.selector.SetFilter(m.searchBar.Content)
		return m, nil
	case tea.WindowSizeMsg:
		m.winSize = msg
		m.selector.ViewSize = ViewSize{Width: msg.Width, Height: msg.Height - 1}
		return m, nil
	}
	return m, nil
}

func (m FileManager) updateDelete(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.Delete):
			return m, m.removeSelectedFiles
		default:
			m.state = Normal
			return m, tea.RequestWindowSize
		}
	case tea.WindowSizeMsg:
		m.winSize = msg
		m.selector.ViewSize = ViewSize{Width: msg.Width, Height: msg.Height - 1}
		return m, nil
	case RemoveErrorMsg:
		m.state = Message
		m.message = "error: " + msg.Error()
		return m, tea.RequestWindowSize
	case RemoveMsg:
		m.state = Loading
		return m, m.selector.RefreshFiles
	}
	return m, nil
}

func (m FileManager) updateMessage(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		m.state = Normal
		return m, tea.RequestWindowSize
	case tea.WindowSizeMsg:
		m.winSize = msg
		m.selector.ViewSize = ViewSize{Width: msg.Width, Height: msg.Height - 1}
		return m, nil
	}
	return m, nil
}

func (m FileManager) View() tea.View {
	switch m.state {
	case Loading:
		return m.spinner.View()
	case Normal:
		return m.selector.View()
	case Search:
		return m.viewSearch()
	case Delete:
		return m.viewDelete()
	case Message:
		return m.viewMessage()
	}
	invalidView := tea.NewView("Invalid program state. There is nothing you can do ¯\\_(ツ)_/¯")
	invalidView.AltScreen = true
	return invalidView
}

func (m FileManager) viewSearch() tea.View {
	selectorString := createSelectorViewStyle(m.winSize.Height - 1).Render(m.selector.View().Content)
	view := tea.NewView(selectorString + "\n" + m.searchBar.View().Content)
	view.AltScreen = true
	return view
}

func (m FileManager) viewDelete() tea.View {
	selectorString := createSelectorViewStyle(m.winSize.Height - 1).Render(m.selector.View().Content)
	view := tea.NewView(selectorString + "\nConfirm deletion by pressing 'd' again.")
	view.AltScreen = true
	return view
}

func (m FileManager) viewMessage() tea.View {
	selectorString := createSelectorViewStyle(m.winSize.Height - 1).Render(m.selector.View().Content)
	view := tea.NewView(selectorString + "\n" + m.message)
	view.AltScreen = true
	return view
}

type RemoveErrorMsg error
type RemoveMsg struct{}

func (m FileManager) removeSelectedFiles() tea.Msg {
	for _, file := range m.selector.GetSelectedFiles() {
		if err := os.Remove(file.AbsolutePath); err != nil {
			return RemoveErrorMsg(err)
		}
	}
	return RemoveMsg{}
}

func createSelectorViewStyle(height int) lipgloss.Style {
	return lipgloss.NewStyle().MaxHeight(height).Height(height)
}
