package main

import (
	"os"
	"path/filepath"

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
	Err
)

type RemoveMsg struct{}

type FileManager struct {
	selector  FileSelector
	searchBar TextInput
	spinner   Spinner
	state     State
	winSize   tea.WindowSizeMsg
	errMsg    string
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
	switch msg := msg.(type) {
	case FilterMsg, FilterQueryMessage:
		selector, cmd := m.selector.Update(msg)
		m.selector = selector.(FileSelector)
		return m, cmd
	case error:
		m.errMsg = "error: " + msg.Error()
		return m.setState(Err)
	}
	switch m.state {
	case Loading:
		return m.updateLoading(msg)
	case Normal:
		return m.updateNormal(msg)
	case Search:
		return m.updateSearch(msg)
	case Delete:
		return m.updateDelete(msg)
	case Err:
		return m.updateErr(msg)
	}
	return m, nil
}

func (m FileManager) updateLoading(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.Quit):
			return m, tea.Quit
		}
	case SpinnerTickMsg:
		spinner, cmd := m.spinner.Update(msg)
		m.spinner = spinner.(Spinner)
		return m, cmd
	case Files:
		selector, _ := m.selector.Update(msg)
		m.selector = selector.(FileSelector)
		return m.setState(Normal)
	}
	return m, nil
}

func (m FileManager) updateNormal(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.Quit):
			return m, tea.Quit
		case key.Matches(msg, DefaultKeyMap.Top):
			m.selector = m.selector.GoTop()
			return m, nil
		case key.Matches(msg, DefaultKeyMap.Bottom):
			m.selector = m.selector.GoBottom()
			return m, nil
		case key.Matches(msg, DefaultKeyMap.Delete):
			if len(m.selector.GetSelectedFiles()) == 0 {
				m.errMsg = "Nothing selected"
				return m.setState(Err)
			}
			return m.setState(Delete)
		case key.Matches(msg, DefaultKeyMap.EnterSelectedDir):
			return m.enterSelectedDir()
		case key.Matches(msg, DefaultKeyMap.EnterParentDir):
			return m.enterParentDir()
		case key.Matches(msg, DefaultKeyMap.Refresh):
			return m.setState(Loading)
		case key.Matches(msg, DefaultKeyMap.ShowSearch):
			return m.setState(Search)
		case key.Matches(msg, DefaultKeyMap.ClearSearch):
			m.searchBar.Content = ""
			m.selector = m.selector.ResetFilter()
			return m, nil
		case key.Matches(msg, DefaultKeyMap.Help):
			cmd, err := Less(DefaultKeyMap.GetHelpText())
			if err != nil {
				m.errMsg = err.Error()
				return m.setState(Err)
			}
			return m, cmd
		}
		selector, cmd := m.selector.Update(msg)
		m.selector = selector.(FileSelector)
		return m, cmd
	case tea.WindowSizeMsg:
		m.winSize = msg
		m.selector.ViewSize = ViewSize(msg)
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
			return m.setState(Normal)
		case key.Matches(msg, DefaultKeyMap.StartSearch):
			m, cmd := m.setState(Normal)
			return m, tea.Batch(
				cmd,
				m.selector.ApplyFilter(m.searchBar.Content),
			)
		}
		searchBar, _ := m.searchBar.Update(msg)
		m.searchBar = searchBar.(TextInput)
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
			return m.setState(Normal)
		}
	case tea.WindowSizeMsg:
		m.winSize = msg
		m.selector.ViewSize = ViewSize{Width: msg.Width, Height: msg.Height - 1}
		return m, nil
	case RemoveMsg:
		return m.setState(Loading)
	}
	return m, nil
}

func (m FileManager) updateErr(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		m, cmd := m.setState(Normal)
		model, cmd2 := m.Update(msg)
		m = model.(FileManager)
		return m, tea.Sequence(cmd, cmd2)
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
	case Err:
		return m.viewErr()
	}
	invalidView := tea.NewView(
		"Invalid program state. There is nothing you can do ¯\\_(ツ)_/¯",
	)
	invalidView.AltScreen = true
	return invalidView
}

func (m FileManager) viewSearch() tea.View {
	selectorString := createSelectorViewStyle(m.winSize.Height - 1).
		Render(m.selector.View().Content)
	view := tea.NewView(selectorString + "\n" + m.searchBar.View().Content)
	view.AltScreen = true
	return view
}

func (m FileManager) viewDelete() tea.View {
	selectorString := createSelectorViewStyle(m.winSize.Height - 1).
		Render(m.selector.View().Content)
	view := tea.NewView(
		selectorString +
			"\nConfirm deletion by pressing '" +
			DefaultKeyMap.Delete.Help().Key +
			"' again.",
	)
	view.AltScreen = true
	return view
}

func (m FileManager) viewErr() tea.View {
	selectorString := createSelectorViewStyle(m.winSize.Height - 1).
		Render(m.selector.View().Content)
	view := tea.NewView(selectorString + "\n" + m.errMsg)
	view.AltScreen = true
	return view
}

func (m FileManager) removeSelectedFiles() tea.Msg {
	for _, file := range m.selector.GetSelectedFiles() {
		if err := os.Remove(file.AbsolutePath); err != nil {
			return err
		}
	}
	return RemoveMsg{}
}

func (m FileManager) setState(state State) (FileManager, tea.Cmd) {
	m.state = state
	if state == Loading {
		return m, tea.Batch(m.selector.RefreshFiles, m.spinner.Tick())
	}
	return m, tea.RequestWindowSize
}

func (m FileManager) enterSelectedDir() (FileManager, tea.Cmd) {
	files := m.selector.GetSelectedFiles()
	if len(files) != 1 {
		m.errMsg = "Multiple entries selected"
		return m.setState(Err)
	}
	path := files[0].AbsolutePath
	if !files[0].IsDir {
		path = filepath.Dir(path)
		if path == m.selector.Root() {
			m.errMsg = "Not a directory"
			return m.setState(Err)
		}
	}
	return m.enterDir(path)
}

func (m FileManager) enterParentDir() (FileManager, tea.Cmd) {
	return m.enterDir(filepath.Dir(m.selector.Root()))
}

func (m FileManager) enterDir(dir string) (FileManager, tea.Cmd) {
	selector, err := NewFileSelector(dir)
	if err != nil {
		m.errMsg = err.Error()
		return m.setState(Err)
	}
	m.selector = selector
	return m.setState(Loading)
}

func createSelectorViewStyle(height int) lipgloss.Style {
	return lipgloss.NewStyle().MaxHeight(height).Height(height)
}
