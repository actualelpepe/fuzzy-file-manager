package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

var logFile *os.File

func SetupLogging() error {
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		return err
	}
	logFile = f
	return nil
}

func Logf(format string, a ...any) (int, error) {
	return fmt.Fprintf(logFile, format, a...)
}

func Log(a ...any) (int, error) {
	return fmt.Fprint(logFile, a...)
}

func Logln(a ...any) (int, error) {
	return fmt.Fprintln(logFile, a...)
}

func CloseLogging() {
	if logFile != nil {
		logFile.Close()
		logFile = nil
	}
}
