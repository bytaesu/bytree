package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type resultMsg struct{ err error }

// SpinWhile shows a spinner with the given message while fn runs.
// Returns the error from fn.
func SpinWhile(msg string, fn func() error) error {
	var fnErr error
	m := spinnerModel{
		spinner: spinner.New(
			spinner.WithSpinner(spinner.Dot),
			spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("6"))),
		),
		message: msg,
		fn:      func() tea.Msg { fnErr = fn(); return resultMsg{fnErr} },
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return err
	}
	return fnErr
}

type spinnerModel struct {
	spinner spinner.Model
	message string
	fn      func() tea.Msg
	done    bool
}

func (m spinnerModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.fn)
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case resultMsg:
		m.done = true
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m spinnerModel) View() string {
	if m.done {
		return ""
	}
	return fmt.Sprintf("%s %s", m.spinner.View(), Dim.Render(m.message))
}
