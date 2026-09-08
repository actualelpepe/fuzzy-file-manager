package main

import (
	"errors"
	"io/fs"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

type SelectedFiles map[File]bool
type Files = []File

type ViewSize tea.WindowSizeMsg

type FilterMsgType int

const (
	Regex FilterMsgType = iota
	Fuzzy
)

type FilterMsg struct {
	filter     string
	file       File
	filterType FilterMsgType
	index      int
}

type FilterQueryMessage string

type FileSelector struct {
	ViewSize ViewSize

	root          string
	filter        string
	cursor        int
	files         Files
	visibleFiles  Files
	selectedFiles SelectedFiles
}

var NoEntriesError error = errors.New("No entries selected")
var MultipleEntriesError error = errors.New("Multiple entries selected")

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
		s.visibleFiles = s.files
		s.cursor = s.validateCursor()
		return s, nil
	case FilterQueryMessage:
		filter := string(msg)
		if filter == "" {
			return s.ResetFilter(), nil
		}
		s.visibleFiles = Files{}
		s.filter = filter
		s.cursor = 0
		return s, s.applyFilter(s.filter, 0)
	case FilterMsg:
		if msg.filter != s.filter {
			return s, nil
		}
		switch msg.filterType {
		case Fuzzy:
			s.visibleFiles = append(s.visibleFiles, msg.file)
		case Regex:
			s.visibleFiles = slices.Insert(s.visibleFiles, 0, msg.file)
		}
		return s, s.applyFilter(msg.filter, msg.index+1)
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

func (s FileSelector) GoTop() FileSelector {
	s.cursor = 0
	return s
}

func (s FileSelector) GoBottom() FileSelector {
	if len(s.visibleFiles) > 0 {
		s.cursor = len(s.visibleFiles) - 1
	}
	return s
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

// Returns a list of selected files. The list is guaranteed to have at
// least one file. Returns an error if the root directory is empty an
// no files are selected.
func (s FileSelector) GetSelectedFiles() (Files, error) {
	files := s.getSelectedFiles()
	if len(files) < 1 {
		return Files{}, NoEntriesError
	}
	return files, nil
}

// Returns one and only one file selected. If there are multiple
// files selected, returns en error.
func (s FileSelector) GetOneSelectedFile() (File, error) {
	files := s.getSelectedFiles()
	if len(files) > 1 {
		return File{}, MultipleEntriesError
	}
	if len(files) < 1 {
		return File{}, NoEntriesError
	}
	return files[0], nil
}

func (s FileSelector) ApplyFilter(filter string) tea.Cmd {
	return func() tea.Msg {
		return FilterQueryMessage(filter)
	}
}

func (s FileSelector) ResetFilter() FileSelector {
	s.filter = ""
	s.visibleFiles = s.files
	return s
}

func (s FileSelector) Root() string {
	return s.root
}

func (s FileSelector) getSelectedFiles() []File {
	if len(s.visibleFiles) < 1 {
		return Files{}
	}
	files := Files{}
	for i, v := range s.selectedFiles {
		if v {
			files = append(files, i)
		}
	}
	if len(files) == 0 {
		files = append(files, s.visibleFiles[s.cursor])
	}
	return files
}

func (s FileSelector) applyFilter(filter string, index int) tea.Cmd {
	if index >= len(s.files) || index < 0 {
		return nil
	}
	return func() tea.Msg {
		for i := index; i < len(s.files); i++ {
			file := s.files[i]
			regexMatch, _ := regexp.MatchString(filter, file.Name)
			if regexMatch {
				return FilterMsg{
					filterType: Regex,
					index:      i,
					file:       file,
					filter:     filter,
				}
			} else if fuzzy.MatchNormalizedFold(filter, file.Name) {
				return FilterMsg{
					filterType: Fuzzy,
					index:      i,
					file:       file,
					filter:     filter,
				}
			}
		}
		return nil
	}
}

func (s FileSelector) handleUserInput(msg tea.KeyPressMsg) FileSelector {
	if len(s.visibleFiles) > 0 {
		switch {
		case key.Matches(msg, DefaultKeyMap.Down):
			s.cursor++
		case key.Matches(msg, DefaultKeyMap.Up):
			s.cursor--
		case key.Matches(msg, DefaultKeyMap.Select):
			selectedFile := s.visibleFiles[s.cursor]
			s.selectedFiles[selectedFile] = !s.selectedFiles[selectedFile]
		case key.Matches(msg, DefaultKeyMap.DeselectAll):
			s.selectedFiles = SelectedFiles{}
		}
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
	if s.visibleFiles[index].IsDir {
		fileViewBuilder.WriteRune('/')
	}
	return fileViewBuilder.String()
}
