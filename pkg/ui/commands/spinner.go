package commands

import (
	"fmt"
	"os"
	"sync"
	"time"

	"golang.org/x/term"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner provides an animated spinner for long-running operations.
type Spinner struct {
	message string
	stop    chan struct{}
	done    chan struct{}
	mu      sync.Mutex
	active  bool
}

// NewSpinner creates a new spinner with the given message.
func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}
}

// Start begins the spinner animation.
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.active {
		s.mu.Unlock()
		return
	}
	s.active = true
	s.mu.Unlock()

	if !term.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Printf("%s %s\n", InfoStyle.Render("[INFO]"), s.message)
		close(s.done)
		return
	}

	go func() {
		defer close(s.done)
		frame := 0
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-s.stop:
				fmt.Print("\r\033[K")
				return
			case <-ticker.C:
				fmt.Printf("\r%s %s", InfoStyle.Render(spinnerFrames[frame]), s.message)
				frame = (frame + 1) % len(spinnerFrames)
			}
		}
	}()
}

// Stop stops the spinner animation.
func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		return
	}
	s.active = false
	s.mu.Unlock()

	close(s.stop)
	<-s.done
}

// UpdateMessage changes the spinner message while running.
func (s *Spinner) UpdateMessage(message string) {
	s.mu.Lock()
	s.message = message
	s.mu.Unlock()
}

// WithSpinner runs a function while displaying a spinner.
func WithSpinner(message string, fn func()) {
	spinner := NewSpinner(message)
	spinner.Start()
	defer spinner.Stop()
	fn()
}
