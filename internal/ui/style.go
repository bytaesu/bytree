package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

var (
	Bold    = lipgloss.NewStyle().Bold(true)
	Title   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6")) // cyan
	Success = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))            // green
	Dim     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))            // gray
	Err     = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))            // red
)

func Print(a ...any)                 { fmt.Println(a...) }
func Printf(format string, a ...any) { fmt.Printf(format+"\n", a...) }

func Successf(format string, a ...any) {
	fmt.Println(Success.Render(fmt.Sprintf("✓ "+format, a...)))
}

func Dimf(format string, a ...any) {
	fmt.Println(Dim.Render(fmt.Sprintf(format, a...)))
}

func Errorf(format string, a ...any) {
	fmt.Fprintln(os.Stderr, Err.Render(fmt.Sprintf("error: "+format, a...)))
}
