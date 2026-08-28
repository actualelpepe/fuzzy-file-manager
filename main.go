package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

func Error(err error) {
	fmt.Printf("error: %v", err)
	os.Exit(1)
}

func main() {
	rootDir := "."
	if len(os.Args) == 2 {
		rootDir = os.Args[1]
	}
	fileSelector, err := NewFileSelector(rootDir)
	if err != nil {
		Error(err)
	}
	program := tea.NewProgram(fileSelector)
	if _, err := program.Run(); err != nil {
		Error(err)
	}
}
