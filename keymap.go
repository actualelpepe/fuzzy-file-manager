package main

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"charm.land/bubbles/v2/key"
)

type KeyMap struct {
	Up           key.Binding
	Down         key.Binding
	Top          key.Binding
	Bottom       key.Binding
	Quit         key.Binding
	ShowSearch   key.Binding
	StartSearch  key.Binding
	CancelSearch key.Binding
	ClearSearch  key.Binding
	Refresh      key.Binding
	Delete       key.Binding
	Select       key.Binding
	Help         key.Binding
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
	Top: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "go to the top of the list"),
	),
	Bottom: key.NewBinding(
		key.WithKeys("G"),
		key.WithHelp("G", "go to the bottom of the list"),
	),
	Quit: key.NewBinding(
		key.WithKeys("esc", "q"),
		key.WithHelp("esc q", "quit the program"),
	),
	ShowSearch: key.NewBinding(
		key.WithKeys("f", "/"),
		key.WithHelp("f /", "show search bar"),
	),
	CancelSearch: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel search"),
	),
	StartSearch: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("↵", "start search"),
	),
	ClearSearch: key.NewBinding(
		key.WithKeys("F"),
		key.WithHelp("F", "clear search"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh files"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete selected entries"),
	),
	Select: key.NewBinding(
		key.WithKeys("space"),
		key.WithHelp("⎵", "select entry"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "show this text"),
	),
}

func (k *KeyMap) GetHelpText() string {
	stringBuilder := strings.Builder{}
	stringBuilder.WriteString(AsciiLogo)
	w := tabwriter.NewWriter(&stringBuilder, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, k.Quit.Help().Key, "\t", k.Quit.Help().Desc)
	fmt.Fprintln(w, k.Up.Help().Key, "\t", k.Up.Help().Desc)
	fmt.Fprintln(w, k.Down.Help().Key, "\t", k.Down.Help().Desc)
	fmt.Fprintln(w, k.ShowSearch.Help().Key, "\t", k.ShowSearch.Help().Desc)
	fmt.Fprintln(w, k.CancelSearch.Help().Key, "\t", k.CancelSearch.Help().Desc)
	fmt.Fprintln(w, k.ClearSearch.Help().Key, "\t", k.ClearSearch.Help().Desc)
	fmt.Fprintln(w, k.Select.Help().Key, "\t", k.Select.Help().Desc)
	fmt.Fprintln(w, k.Delete.Help().Key, "\t", k.Delete.Help().Desc)
	fmt.Fprintln(w, k.Refresh.Help().Key, "\t", k.Refresh.Help().Desc)
	fmt.Fprintln(w, k.Help.Help().Key, "\t", k.Help.Help().Desc)
	w.Flush()
	return stringBuilder.String()
}
