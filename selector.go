package main

import (
	"io/fs"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type Files = []File

type FileSelector struct {
	root          string
	files         Files
	cursor        int
	selectedFiles map[int]bool

	showSpinner bool
	spinner     Spinner

	showSearchBar bool
	searchBar     TextInput

	winSize tea.WindowSizeMsg
}

// Create new file selector that lists all files
// in path recursively.
func NewFileSelector(path string) (FileSelector, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return FileSelector{}, err
	}
	return FileSelector{
		root:          absPath,
		showSpinner:   true,
		spinner:       NewSpinner(),
		showSearchBar: false,
		searchBar:     NewTextInput(),
	}, nil
}

func (s FileSelector) Init() tea.Cmd {
	return tea.Batch(walkDir(s.root), s.spinner.Init(), s.searchBar.Init())
}

func (s FileSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if !s.showSpinner && !s.showSearchBar {
			return s.handleUserInput(msg)
		}
	case tea.WindowSizeMsg:
		s.winSize = msg
	case Files:
		s.files = msg
		s.selectedFiles = make(map[int]bool)
		s.showSpinner = false
		s.cursor = 0
	case spinnerTickMsg:
		if s.showSpinner {
			spinner, spinnerCmd := s.spinner.Update(msg)
			s.spinner = spinner.(Spinner)
			return s, spinnerCmd
		}
	}
	if s.showSearchBar {
		searchBar, _ := s.searchBar.Update(msg)
		s.searchBar = searchBar.(TextInput)
		if s.searchBar.EndOfInput {
			s.showSearchBar = false
		}
	}
	return s, nil
}

func (s FileSelector) View() tea.View {
	if s.winSize.Height <= 0 {
		return tea.NewView("")
	}
	if s.showSpinner {
		return s.spinner.View()
	}
	viewStringBuilder := strings.Builder{}
	maxVisibileEntries := s.winSize.Height - 1
	startIndex := s.cursor / maxVisibileEntries * maxVisibileEntries
	endIndex := startIndex + (maxVisibileEntries)
	if endIndex > len(s.files) {
		endIndex = len(s.files)
	}
	for i := startIndex; i < endIndex; i++ {
		viewStringBuilder.WriteString(s.fileStringView(i))
	}
	if s.showSearchBar {
		viewStringBuilder.WriteString(s.searchBar.View().Content)
	}
	view := tea.NewView(viewStringBuilder.String())
	view.AltScreen = true
	return view
}

func (s FileSelector) handleUserInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "down", "j":
		s.cursor++
		if s.cursor == len(s.files) {
			s.cursor = 0
		}
	case "up", "k":
		s.cursor--
		if s.cursor == -1 {
			s.cursor = len(s.files) - 1
		}
	case "space", "s":
		s.selectedFiles[s.cursor] = !s.selectedFiles[s.cursor]
	case "f", "/":
		s.searchBar = NewTextInput()
		s.showSearchBar = true
	case "ctrl+c", "q", "esc":
		return s, tea.Quit
	}
	return s, nil
}

func (s FileSelector) fileStringView(index int) string {
	fileViewBuilder := strings.Builder{}
	fileViewBuilder.WriteRune(' ')
	if s.selectedFiles[index] {
		fileViewBuilder.WriteRune('*')
	} else {
		fileViewBuilder.WriteRune(' ')
	}
	if index == s.cursor {
		fileViewBuilder.WriteRune('>')
	} else {
		fileViewBuilder.WriteRune(' ')
	}
	fileViewBuilder.WriteRune(' ')
	fileViewBuilder.WriteString(s.files[index].path)
	fileViewBuilder.WriteRune('\n')
	return fileViewBuilder.String()
}

func walkDir(path string) tea.Cmd {
	return func() tea.Msg {
		files := make(Files, 0)
		if err := filepath.Walk(path,
			func(path string, info fs.FileInfo, err error) error {
				if err != nil {
					// I must collect the errors and put them in some
					// kind of error screen
					return nil
				}
				files = append(files, File{name: info.Name(), path: path})
				return nil
			}); err != nil {
			return err
		}
		return files
	}
}
