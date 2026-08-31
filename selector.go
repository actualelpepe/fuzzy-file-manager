package main

import (
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

type SelectedFiles map[File]bool

type ViewSize struct {
	Width, Height int
}

type FileSelector struct {
	root          string
	cursor        int
	files         DirWalkMsg
	visibleFiles  DirWalkMsg
	selectedFiles SelectedFiles

	ViewSize ViewSize
	Filter   string
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
	return tea.Batch(WalkDir(s.root))
}

func (s FileSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		s = s.handleUserInput(msg)
		s.cursor = s.validateCursor()
		return s, nil
	case DirWalkMsg:
		s.files = msg
		s.selectedFiles = make(SelectedFiles)
		s.cursor = 0
		return s, nil
	case VisibleFilesUpdateMsg:
		s.visibleFiles = s.updateVisibleFiles()
		s.cursor = s.validateCursor()
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

func (s FileSelector) updateVisibleFiles() DirWalkMsg {
	if s.Filter == "" {
		return s.files
	}
	return s.filterFiles()
}

func (s FileSelector) filterFiles() DirWalkMsg {
	files := DirWalkMsg{}
	for _, file := range s.files {
		regexMatch, _ := regexp.MatchString(s.Filter, file.path)
		if regexMatch {
			files = slices.Insert(files, 0, file)
		} else if fuzzy.MatchNormalizedFold(s.Filter, file.path) {
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
	fileViewBuilder.WriteString(s.visibleFiles[index].path)
	return fileViewBuilder.String()
}
