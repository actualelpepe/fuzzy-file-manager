package main

import "charm.land/bubbles/v2/key"

type KeyMap struct {
	Up            key.Binding
	Down          key.Binding
	Cancel        key.Binding
	Exit          key.Binding
	Search        key.Binding
	ConfirmSearch key.Binding
	CancelSearch  key.Binding
	ResetSearch   key.Binding
	Refresh       key.Binding
	Delete        key.Binding
	Select        key.Binding
	Help          key.Binding
}

var DefaultKeyMap = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("↑ k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("↓ j", "move down"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("ctrl+c", "esc"),
		key.WithHelp("ctrl+c esc", "cancel current action"),
	),
	Exit: key.NewBinding(
		key.WithKeys("ctrl+c", "esc", "q"),
		key.WithHelp("ctrl+c esc q", "exit the program"),
	),
	Search: key.NewBinding(
		key.WithKeys("f", "/"),
		key.WithHelp("f /", "enable search mode"),
	),
	ConfirmSearch: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("↵", "confirm search pattern and exit search mode"),
	),
	CancelSearch: key.NewBinding(
		key.WithKeys("ctrl+c", "esc"),
		key.WithHelp("ctrl+c esc", "reset search pattern and exit search mode"),
	),
	ResetSearch: key.NewBinding(
		key.WithKeys("F"),
		key.WithHelp("F", "reset search pattern"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh files"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d", "backspace", "delete"),
		key.WithHelp("⌫ d del", "delete selected files"),
	),
	Select: key.NewBinding(
		key.WithKeys("space"),
		key.WithHelp("⎵", "select file under the cursor"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "show this text"),
	),
}
