package main

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

type SelectedFiles map[File]bool
type Files = []File

type ViewSize = tea.WindowSizeMsg

type FileSelector struct {
	ViewSize ViewSize

	root          string
	cursor        int
	files         Files
	visibleFiles  Files
	selectedFiles SelectedFiles
	filter        string
}

// Create new file selector that lists all files
// in path recursively.
func NewFileSelector(path string) (FileSelector, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return FileSelector{}, err
	}
	return FileSelector{
		root: absPath,
	}, nil
}

func (s FileSelector) Init() tea.Cmd {
	return s.RefreshFiles
}

func (s FileSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		s = s.handleUserInput(msg)
		s.cursor = s.validateCursor()
		return s, nil
	case Files:
		s.files = msg
		s.selectedFiles = make(SelectedFiles)
		s = s.updateVisibleFiles()
		return s, nil
	}
	return s, nil
}

func (s FileSelector) View() tea.View {
	if s.ViewSize.Height <= 0 {
		return tea.NewView("")
	}
	viewStringBuilder := strings.Builder{}
	viewStringBuilder.WriteString(s.root)
	viewStringBuilder.WriteRune('\n')
	maxVisibileEntries := s.ViewSize.Height - 1
	startIndex := s.cursor / maxVisibileEntries * maxVisibileEntries
	endIndex := startIndex + (maxVisibileEntries)
	if endIndex > len(s.visibleFiles) {
		endIndex = len(s.visibleFiles)
	}
	for i := startIndex; i < endIndex; i++ {
		viewStringBuilder.WriteString(s.fileStringView(i))
		viewStringBuilder.WriteRune('\n')
	}
	view := tea.NewView(viewStringBuilder.String())
	view.AltScreen = true
	return view
}

func (s FileSelector) SetFilter(filter string) FileSelector {
	s.filter = filter
	return s.updateVisibleFiles()
}

func (s FileSelector) ResetFilter() FileSelector {
	return s.SetFilter("")
}

func (s FileSelector) RefreshFiles() tea.Msg {
	files := Files{}
	if err := filepath.Walk(s.root,
		func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			relativePath := path[len(s.root):]
			if len(relativePath) == 0 {
				return nil
			}
			file := File{
				Name:         info.Name(),
				AbsolutePath: path,
				RelativePath: relativePath[1:],
				IsDir:        info.IsDir(),
			}
			files = append(files, file)
			return nil
		}); err != nil {
		return err
	}
	return files
}

func (s FileSelector) updateVisibleFiles() FileSelector {
	s.visibleFiles = s.getVisibleFiles()
	s.cursor = s.validateCursor()
	return s
}

func (s FileSelector) GetSelectedFiles() []File {
	cursorFile := s.visibleFiles[s.cursor]
	files := []File{cursorFile}
	for i, v := range s.selectedFiles {
		if v && i != cursorFile {
			files = append(files, i)
		}
	}
	return files
}

func (s FileSelector) handleUserInput(msg tea.KeyPressMsg) FileSelector {
	switch msg.String() {
	case "down", "j":
		if len(s.visibleFiles) > 0 {
			s.cursor++
		}
	case "up", "k":
		if len(s.visibleFiles) > 0 {
			s.cursor--
			if s.cursor == -1 {
				s.cursor = len(s.visibleFiles) - 1
			}
		}
	case "space", "s":
		selectedFile := s.visibleFiles[s.cursor]
		s.selectedFiles[selectedFile] = !s.selectedFiles[selectedFile]
	}
	return s
}

func (s FileSelector) validateCursor() int {
	if s.cursor >= len(s.visibleFiles) {
		return 0
	} else if s.cursor < 0 {
		return len(s.visibleFiles) - 1
	}
	return s.cursor
}

func (s FileSelector) getVisibleFiles() Files {
	if s.filter == "" {
		return s.files
	}
	return s.filterFiles()
}

func (s FileSelector) filterFiles() Files {
	files := Files{}
	for _, file := range s.files {
		regexMatch, _ := regexp.MatchString(s.filter, file.RelativePath)
		if regexMatch {
			files = slices.Insert(files, 0, file)
		} else if fuzzy.MatchNormalizedFold(s.filter, file.RelativePath) {
			files = append(files, file)
		}
	}
	return files
}

func (s FileSelector) fileStringView(index int) string {
	fileViewBuilder := strings.Builder{}
	fileViewBuilder.WriteRune(' ')
	if s.selectedFiles[s.visibleFiles[index]] {
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
	fileViewBuilder.WriteString(s.visibleFiles[index].RelativePath)
	return fileViewBuilder.String()
}
