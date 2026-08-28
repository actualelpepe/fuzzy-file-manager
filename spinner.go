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
		frames:       []string{"/", "-", "\\", "|", "/", "-", "\\", "|"},
		currentFrame: 0,
		speed:        time.Second / 5,
	}
}

func (s Spinner) Init() tea.Cmd {
	return spinnerTick(s.speed)
}

func (s Spinner) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case spinnerTickMsg:
		s.currentFrame++
		if s.currentFrame == len(s.frames) {
			s.currentFrame = 0
		}
		return s, spinnerTick(s.speed)
	}
	return s, nil
}

func (s Spinner) View() tea.View {
	view := tea.NewView(s.frames[s.currentFrame])
	view.AltScreen = true
	return view
}

type spinnerTickMsg struct{}

func spinnerTick(speed time.Duration) tea.Cmd {
	return tea.Tick(speed, func(time.Time) tea.Msg { return spinnerTickMsg{} })
}
