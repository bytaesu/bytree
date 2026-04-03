package ui

import (
	"fmt"
	"sync/atomic"
	"time"
)

var frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// SpinWhile shows a spinner with the given message while fn runs.
func SpinWhile(msg string, fn func() error) error {
	var done atomic.Bool

	go func() {
		i := 0
		for !done.Load() {
			frame := Dim.Render(frames[i%len(frames)] + " " + msg)
			fmt.Printf("\r\033[2K%s", frame)
			i++
			time.Sleep(80 * time.Millisecond)
		}
	}()

	err := fn()
	done.Store(true)
	fmt.Print("\r\033[2K")

	return err
}
