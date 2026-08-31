package main

import (
	tea "charm.land/bubbletea/v2"
)

type TextInput struct {
	Content string
	Prompt  string
}

func NewTextInput() TextInput {
	return TextInput{
		Content: "",
		Prompt:  "> ",
	}
}

func (t TextInput) Init() tea.Cmd {
	return nil
}

func (t TextInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.Text != "" {
			t.Content += msg.Text
		} else {
			switch msg.Code {
			case tea.KeyBackspace:
				if t.Content != "" {
					t.Content = t.Content[:len(t.Content)-1]
				}
			}
		}
	}
	return t, nil
}

func (t TextInput) View() tea.View {
	view := tea.NewView(t.Prompt + t.Content)
	view.AltScreen = true
	return view
}
