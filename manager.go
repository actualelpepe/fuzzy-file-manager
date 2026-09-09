package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type State int

const (
	Refresh State = iota
	Normal
	Search
	Delete
	NewFileOrDir
	NewFile
	NewDir
	Rename
	Logs
)

type RemoveMsg struct{}

type FileManager struct {
	selector   FileSelector
	searchBar  TextInput
	nameBar    TextInput
	spinner    Spinner
	state      State
	winSize    tea.WindowSizeMsg
	fileBuffer Files
	errMsg     string
}

var NotADirectoryError error = errors.New("Not a directory")
var FileAlreadyExistsError error = errors.New("File already exists")

func NewFileManager(path string) (FileManager, error) {
	selector, err := NewFileSelector(path)
	if err != nil {
		return FileManager{}, err
	}
	searchBar := NewTextInput()
	searchBar.Prompt = "/"
	return FileManager{
		state:     Refresh,
		selector:  selector,
		spinner:   NewSpinner(),
		nameBar:   NewTextInput(),
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
		return m.setState(Logs)
	}
	switch m.state {
	case Refresh:
		return m.updateRefresh(msg)
	case Normal:
		return m.updateNormal(msg)
	case Search:
		return m.updateSearch(msg)
	case Delete:
		return m.updateDelete(msg)
	case NewFileOrDir:
		return m.updateFileOrDir(msg)
	case Rename:
		return m.updateInput(msg, m.renameSelectedFile)
	case NewFile:
		return m.updateInput(msg, m.newFile)
	case NewDir:
		return m.updateInput(msg, m.newDir)
	case Logs:
		return m.updateLogs(msg)
	}
	return m, nil
}

func (m FileManager) updateRefresh(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			_, err := m.selector.GetSelectedFiles()
			if err != nil {
				return m.logs(err)
			}
			return m.setState(Delete)
		case key.Matches(msg, DefaultKeyMap.Yank):
			fileBuffer, err := m.selector.GetSelectedFiles()
			if err != nil {
				return m.logs(err)
			}
			m.fileBuffer = fileBuffer
			m.selector = m.selector.DeselectFiles()
			return m.logs("Yanked ", len(m.fileBuffer), " files")
		case key.Matches(msg, DefaultKeyMap.Copy):
			return m.copyBufferedFiles()
		case key.Matches(msg, DefaultKeyMap.Move):
			return m.moveBufferedFiles()
		case key.Matches(msg, DefaultKeyMap.EnterSelectedDir):
			return m.enterSelectedDir()
		case key.Matches(msg, DefaultKeyMap.EnterParentDir):
			return m.enterParentDir()
		case key.Matches(msg, DefaultKeyMap.NewFileOrDir):
			return m.setState(NewFileOrDir)
		case key.Matches(msg, DefaultKeyMap.Refresh):
			return m.setState(Refresh)
		case key.Matches(msg, DefaultKeyMap.Rename):
			file, err := m.selector.GetOneSelectedFile()
			if err != nil {
				return m.logs(err)
			}
			m.nameBar.Prompt = "rename: "
			m.nameBar.Content = file.Name
			return m.setState(Rename)
		case key.Matches(msg, DefaultKeyMap.ShowSearch):
			return m.setState(Search)
		case key.Matches(msg, DefaultKeyMap.ClearSearch):
			m.searchBar.Content = ""
			m.selector = m.selector.ResetFilter()
			return m, nil
		case key.Matches(msg, DefaultKeyMap.Help):
			cmd, err := Less(DefaultKeyMap.GetHelpText())
			if err != nil {
				return m.logs(err)
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
		case key.Matches(msg, DefaultKeyMap.Cancel):
			m.searchBar.Content = ""
			m.selector = m.selector.ResetFilter()
			return m.setState(Normal)
		case key.Matches(msg, DefaultKeyMap.Accept):
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
		return m.setState(Refresh)
	}
	return m, nil
}

func (m FileManager) updateInput(msg tea.Msg, cmd func(string) tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.Cancel):
			m.nameBar.Content = ""
			return m.setState(Normal)
		case key.Matches(msg, DefaultKeyMap.Accept):
			input := m.nameBar.Content
			m.nameBar.Content = ""
			m, stateCmd := m.setState(Refresh)
			return m, tea.Sequence(
				cmd(input),
				stateCmd,
			)
		}
		nameBar, _ := m.nameBar.Update(msg)
		m.nameBar = nameBar.(TextInput)
		return m, nil
	case tea.WindowSizeMsg:
		m.winSize = msg
		m.selector.ViewSize = ViewSize{Width: msg.Width, Height: msg.Height - 1}
		return m, nil
	}
	return m, nil
}

func (m FileManager) updateFileOrDir(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.NewDir):
			m.nameBar.Prompt = "new directory name: "
			return m.setState(NewDir)
		case key.Matches(msg, DefaultKeyMap.NewFile):
			m.nameBar.Prompt = "new file name: "
			return m.setState(NewFile)
		default:
			return m.setState(Normal)
		}
	case tea.WindowSizeMsg:
		m.winSize = msg
		m.selector.ViewSize = ViewSize{Width: msg.Width, Height: msg.Height - 1}
		return m, nil
	}
	return m, nil
}

func (m FileManager) updateLogs(msg tea.Msg) (tea.Model, tea.Cmd) {
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
	case Refresh:
		return m.spinner.View()
	case Normal:
		return m.selector.View()
	case Search:
		return m.viewSearch()
	case Delete:
		return m.viewDelete()
	case NewFileOrDir:
		return m.viewNewFileOrDir()
	case NewFile, NewDir, Rename:
		return m.viewInput()
	case Logs:
		return m.viewErr()
	default:
		invalidView := tea.NewView(
			"Invalid program state. There is nothing you can do ¯\\_(ツ)_/¯",
		)
		invalidView.AltScreen = true
		return invalidView
	}
}

func (m FileManager) viewSearch() tea.View {
	selectorString := createSelectorViewStyle(m.winSize.Height - 1).
		Render(m.selector.View().Content)
	view := tea.NewView(selectorString + "\n" + m.searchBar.View().Content)
	view.AltScreen = true
	return view
}

func (m FileManager) viewInput() tea.View {
	selectorString := createSelectorViewStyle(m.winSize.Height - 1).
		Render(m.selector.View().Content)
	view := tea.NewView(selectorString + "\n" + m.nameBar.View().Content)
	view.AltScreen = true
	return view
}

func (m FileManager) viewNewFileOrDir() tea.View {
	selectorString := createSelectorViewStyle(m.winSize.Height - 1).
		Render(m.selector.View().Content)
	view := tea.NewView(selectorString + "\n" + "New [f]ile or [d]irectory?")
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
	files, err := m.selector.GetSelectedFiles()
	if err != nil {
		return err
	}
	for _, file := range files {
		if err := os.RemoveAll(file.AbsolutePath); err != nil {
			return err
		}
	}
	return RemoveMsg{}
}

func (m FileManager) newFile(name string) tea.Cmd {
	return func() tea.Msg {
		newFilePath := filepath.Join(m.selector.Root(), name)
		if _, err := os.Stat(newFilePath); err == nil {
			return FileAlreadyExistsError
		}
		if _, err := os.Create(newFilePath); err != nil {
			return err
		}
		return nil
	}
}

func (m FileManager) newDir(name string) tea.Cmd {
	Logln("Creating dir", name)
	return func() tea.Msg {
		newDirPath := filepath.Join(m.selector.Root(), name)
		if _, err := os.Stat(newDirPath); err == nil {
			return FileAlreadyExistsError
		}
		if err := os.Mkdir(newDirPath, 0777); err != nil {
			return err
		}
		return nil
	}
}

func (m FileManager) renameSelectedFile(newName string) tea.Cmd {
	return func() tea.Msg {
		file, err := m.selector.GetOneSelectedFile()
		if err != nil {
			return err
		}
		dir := filepath.Dir(file.AbsolutePath)
		if err := os.Rename(file.AbsolutePath, filepath.Join(dir, newName)); err != nil {
			return err
		}
		return nil
	}
}

func (m FileManager) copyBufferedFiles() (FileManager, tea.Cmd) {
	if len(m.fileBuffer) == 0 {
		return m.logs("No files in buffer")
	}
	sourcePaths := make([]string, 0, len(m.fileBuffer))
	for _, file := range m.fileBuffer {
		sourcePaths = append(sourcePaths, file.AbsolutePath)
	}
	args := []string{"-r"}
	args = append(args, sourcePaths...)
	args = append(args, "-t", m.selector.Root())
	cmd := exec.Command("cp", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		less, errLess := Less(string(output))
		if errLess != nil {
			return m.logs(errLess)
		}
		return m, less
	}
	m.fileBuffer = Files{}
	return m.setState(Refresh)
}

func (m FileManager) moveBufferedFiles() (FileManager, tea.Cmd) {
	if len(m.fileBuffer) == 0 {
		return m.logs("No files in buffer")
	}
	sourcePaths := make([]string, 0, len(m.fileBuffer))
	for _, file := range m.fileBuffer {
		sourcePaths = append(sourcePaths, file.AbsolutePath)
	}
	args := sourcePaths
	args = append(args, "-t", m.selector.Root())
	cmd := exec.Command("mv", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		less, errLess := Less(string(output))
		if errLess != nil {
			return m.logs(errLess)
		}
		return m, less
	}
	m.fileBuffer = Files{}
	return m.setState(Refresh)
}

func (m FileManager) setState(state State) (FileManager, tea.Cmd) {
	m.state = state
	if state == Refresh {
		return m, tea.Batch(m.selector.RefreshFiles, m.spinner.Tick())
	}
	return m, tea.RequestWindowSize
}

func (m FileManager) enterSelectedDir() (FileManager, tea.Cmd) {
	file, err := m.selector.GetOneSelectedFile()
	if err != nil {
		return m.logs(err)
	}
	path := file.AbsolutePath
	if !file.IsDir {
		path = filepath.Dir(path)
		if path == m.selector.Root() {
			return m.logs(NotADirectoryError)
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
		return m.logs(err)
	}
	m.selector = selector
	return m.setState(Refresh)
}

func (m FileManager) logs(a ...any) (FileManager, tea.Cmd) {
	m.errMsg = fmt.Sprint(a...)
	return m.setState(Logs)
}

func createSelectorViewStyle(height int) lipgloss.Style {
	return lipgloss.NewStyle().MaxHeight(height).Height(height)
}
