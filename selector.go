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
	selectedFiles map[int]bool
	cursor        int
	showSpinner   bool
	spinner       Spinner
}

// Create new file selector that lists all files
// in path recursively.
func NewFileSelector(path string) (FileSelector, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return FileSelector{}, err
	}
	return FileSelector{
		root:        absPath,
		showSpinner: true,
		spinner:     NewSpinner(),
	}, nil
}

func (s FileSelector) Init() tea.Cmd {
	return tea.Batch(walkDir(s.root), s.spinner.Init())
}

func (s FileSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
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
		case "ctrl+c", "q", "esc":
			return s, tea.Quit
		}
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
	return s, nil
}

func (s FileSelector) View() tea.View {
	if s.showSpinner {
		return s.spinner.View()
	}
	viewStringBuilder := strings.Builder{}
	for index, file := range s.files {
		viewStringBuilder.WriteRune(' ')
		if s.selectedFiles[index] {
			viewStringBuilder.WriteRune('*')
		} else {
			viewStringBuilder.WriteRune(' ')
		}
		if index == s.cursor {
			viewStringBuilder.WriteRune('>')
		} else {
			viewStringBuilder.WriteRune(' ')
		}
		viewStringBuilder.WriteRune(' ')
		viewStringBuilder.WriteString(file.path)
		viewStringBuilder.WriteRune('\n')
	}
	view := tea.NewView(viewStringBuilder.String())
	view.AltScreen = true
	return view
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
