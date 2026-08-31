package main

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

type Files = []File
type SelectedFiles map[File]bool

type FileSelector struct {
	root          string
	cursor        int
	files         Files
	searchedFiles Files
	selectedFiles SelectedFiles

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
	searchBar := NewTextInput()
	searchBar.Prompt = "/"
	return FileSelector{
		root:          absPath,
		showSpinner:   true,
		spinner:       NewSpinner(),
		showSearchBar: false,
		searchBar:     searchBar,
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
		return s, nil
	case Files:
		s.files = msg
		s.searchedFiles = msg
		s.selectedFiles = make(SelectedFiles)
		s.showSpinner = false
		s.cursor = 0
		return s, nil
	case spinnerTickMsg:
		if s.showSpinner {
			spinner, spinnerCmd := s.spinner.Update(msg)
			s.spinner = spinner.(Spinner)
			return s, spinnerCmd
		}
	}
	s = s.updateSearchBar(msg)
	s = s.filterFiles()
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
	if endIndex > len(s.searchedFiles) {
		endIndex = len(s.searchedFiles)
	}
	entriesWritten := 0
	for i := startIndex; i < endIndex; i++ {
		viewStringBuilder.WriteString(s.fileStringView(i))
		viewStringBuilder.WriteRune('\n')
		entriesWritten++
	}
	if entriesWritten == 0 {
		viewStringBuilder.WriteString("*Crickets*\n")
		entriesWritten++
	}
	if s.showSearchBar {
		padding := ""
		if entriesWritten < maxVisibileEntries {
			for i := entriesWritten; i < maxVisibileEntries; i++ {
				padding += "\n"
			}
		}
		viewStringBuilder.WriteString(padding + s.searchBar.View().Content)
	}
	view := tea.NewView(viewStringBuilder.String())
	view.AltScreen = true
	return view
}

func (s FileSelector) filterFiles() FileSelector {
	if s.searchBar.Content != "" {
		s.searchedFiles = Files{}
		for _, file := range s.files {
			regexMatch, _ := regexp.MatchString(s.searchBar.Content, file.name)
			if regexMatch || fuzzy.MatchNormalizedFold(s.searchBar.Content, file.name) {
				s.searchedFiles = append(s.searchedFiles, file)
			}
		}
		s.cursor = 0
	} else {
		s.searchedFiles = s.files
	}
	return s
}

func (s FileSelector) updateSearchBar(msg tea.Msg) FileSelector {
	if s.showSearchBar {
		searchBar, _ := s.searchBar.Update(msg)
		s.searchBar = searchBar.(TextInput)
		if s.searchBar.EndOfInput {
			s.showSearchBar = false
		}
	}
	return s
}

func (s FileSelector) handleUserInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "down", "j":
		if len(s.searchedFiles) > 0 {
			s.cursor++
			if s.cursor == len(s.searchedFiles) {
				s.cursor = 0
			}
		}
	case "up", "k":
		if len(s.searchedFiles) > 0 {
			s.cursor--
			if s.cursor == -1 {
				s.cursor = len(s.searchedFiles) - 1
			}
		}
	case "space", "s":
		selectedFile := s.searchedFiles[s.cursor]
		s.selectedFiles[selectedFile] = !s.selectedFiles[selectedFile]
	case "f", "/":
		s.searchBar.EndOfInput = false
		s.showSearchBar = true
	case "F":
		s.searchBar.Content = ""
		s.searchedFiles = s.files
		s.cursor = 0
	}
	return s, nil
}

func (s FileSelector) fileStringView(index int) string {
	fileViewBuilder := strings.Builder{}
	fileViewBuilder.WriteRune(' ')
	if s.selectedFiles[s.searchedFiles[index]] {
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
	fileViewBuilder.WriteString(s.searchedFiles[index].path)
	return fileViewBuilder.String()
}

func walkDir(path string) tea.Cmd {
	return func() tea.Msg {
		files := Files{}
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
