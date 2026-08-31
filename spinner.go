package main

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

type Spinner struct {
	frames       []string
	currentFrame int
	speed        time.Duration
}

func NewSpinner() Spinner {
	return Spinner{
		frames:       []string{".", "..", "..."},
		currentFrame: 0,
		speed:        time.Second / 5,
	}
}

func (s Spinner) Init() tea.Cmd {
	return SpinnerTick(s.speed)
}

func (s Spinner) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case SpinnerTickMsg:
		s.currentFrame++
		if s.currentFrame == len(s.frames) {
			s.currentFrame = 0
		}
		return s, SpinnerTick(s.speed)
	}
	return s, nil
}

func (s Spinner) View() tea.View {
	view := tea.NewView("Loading" + s.frames[s.currentFrame])
	view.AltScreen = true
	return view
}
