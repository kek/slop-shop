package main

import (
	"fmt"
)

// Spinner represents a simple terminal spinner
type Spinner struct {
	frames []string
	index  int
}

// NewSpinner creates a new spinner with default frames
func NewSpinner() *Spinner {
	return &Spinner{
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		index:  0,
	}
}

// Next returns the next spinner frame
func (s *Spinner) Next() string {
	frame := s.frames[s.index]
	s.index = (s.index + 1) % len(s.frames)
	return frame
}

// Spin displays the spinner with a message
func (s *Spinner) Spin(message string) {
	fmt.Printf("\r%s %s", s.Next(), message)
}

// Stop clears the spinner line
func (s *Spinner) Stop() {
	fmt.Print("\r\033[K") // Clear the line
}
