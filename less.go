package main

import (
	"os"
	"os/exec"

	tea "charm.land/bubbletea/v2"
)

func Less(data string) (tea.Cmd, error) {
	file, err := os.CreateTemp("", "")
	if err != nil {
		return nil, nil
	}
	file.WriteString(data)
	return tea.ExecProcess(exec.Command("less", file.Name()),
		func(err error) tea.Msg {
			file.Close()
			os.Remove(file.Name())
			return nil
		}), nil
}
