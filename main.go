package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

func Error(err error) {
	fmt.Printf("error: %v\n", err)
	os.Exit(1)
}

func main() {
	if err := SetupLogging(); err != nil {
		Error(err)
	}
	defer CloseLogging()
	rootDir := "."
	if len(os.Args) == 2 {
		rootDir = os.Args[1]
	}
	fileSelector, err := NewFileManager(rootDir)
	if err != nil {
		Error(err)
	}
	program := tea.NewProgram(fileSelector)
	if _, err := program.Run(); err != nil {
		Error(err)
	}
}
