package main

import (
	"io/fs"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type Files = []File

type FileSelector struct {
	root        string
	files       Files
	showSpinner bool
	spinner     Spinner
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
		case "ctrl+c", "q", "esc":
			return s, tea.Quit
		}
	case Files:
		s.files = msg
		s.showSpinner = false
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
	for _, file := range s.files {
		viewStringBuilder.WriteString(file.path)
		viewStringBuilder.WriteRune('\n')
	}
	view := tea.NewView(viewStringBuilder.String())
	view.AltScreen = true
	return view
}

func walkDir(path string) tea.Cmd {
	return func() tea.Msg {
		files := Files{}
		if err := filepath.Walk(path,
			func(path string, info fs.FileInfo, err error) error {
				if err != nil {
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
