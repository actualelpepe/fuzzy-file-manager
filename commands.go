package main

import (
	"io/fs"
	"path/filepath"
	"time"

	tea "charm.land/bubbletea/v2"
)

type VisibleFilesUpdateMsg struct{}
type SpinnerTickMsg struct{}
type DirWalkMsg = []File

func RequestVisibleFilesUpdate() VisibleFilesUpdateMsg {
	return VisibleFilesUpdateMsg{}
}

func SpinnerTick(speed time.Duration) tea.Cmd {
	return tea.Tick(speed, func(time.Time) tea.Msg { return SpinnerTickMsg{} })
}

func WalkDir(root string) tea.Cmd {
	return func() tea.Msg {
		files := DirWalkMsg{}
		if err := filepath.Walk(root,
			func(path string, info fs.FileInfo, err error) error {
				if err != nil {
					// I must collect the errors and put them in some
					// kind of error screen
					return nil
				}
				relativePath := path[len(root):]
				if len(relativePath) == 0 {
					return nil
				}
				files = append(files, File{name: info.Name(), path: relativePath[1:]})
				return nil
			}); err != nil {
			return err
		}
		return files
	}
}
