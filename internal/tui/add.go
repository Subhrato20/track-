package tui

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type addModel struct {
	trackingInput textinput.Model
	nicknameInput textinput.Model
	focusIndex    int
	loading       bool
	spinner       spinner.Model
	err           error
	width         int
}

func newAddModel() addModel {
	ti := textinput.New()
	ti.Placeholder = "9400111899223456789012"
	ti.Focus()
	ti.CharLimit = 40
	ti.Width = 35

	ni := textinput.New()
	ni.Placeholder = "Mom's birthday gift (optional)"
	ni.CharLimit = 50
	ni.Width = 35

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7c3aed"))

	return addModel{
		trackingInput: ti,
		nicknameInput: ni,
		spinner:       s,
	}
}

func (m *addModel) Reset() {
	m.trackingInput.SetValue("")
	m.nicknameInput.SetValue("")
	m.trackingInput.Focus()
	m.nicknameInput.Blur()
	m.focusIndex = 0
	m.loading = false
	m.err = nil
}

func (m addModel) TrackingNumber() string {
	return strings.TrimSpace(m.trackingInput.Value())
}

func (m addModel) Nickname() string {
	return strings.TrimSpace(m.nicknameInput.Value())
}

func (m addModel) Validate() error {
	tn := m.TrackingNumber()
	if tn == "" {
		return fmt.Errorf("tracking number is required")
	}

	// USPS tracking numbers are typically 20-34 digits
	digits := 0
	for _, r := range tn {
		if unicode.IsDigit(r) {
			digits++
		}
	}
	if digits < 10 {
		return fmt.Errorf("tracking number seems too short")
	}

	return nil
}

func (m addModel) Update(msg tea.Msg) (addModel, tea.Cmd) {
	if m.loading {
		if tickMsg, ok := msg.(spinner.TickMsg); ok {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(tickMsg)
			return m, cmd
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab":
			if m.focusIndex == 0 {
				m.focusIndex = 1
				m.trackingInput.Blur()
				m.nicknameInput.Focus()
			} else {
				m.focusIndex = 0
				m.nicknameInput.Blur()
				m.trackingInput.Focus()
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	if m.focusIndex == 0 {
		m.trackingInput, cmd = m.trackingInput.Update(msg)
	} else {
		m.nicknameInput, cmd = m.nicknameInput.Update(msg)
	}
	return m, cmd
}

func (m addModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Add New Package"))
	b.WriteString("\n\n")

	b.WriteString("  Tracking Number:\n")
	b.WriteString("  " + m.trackingInput.View())
	b.WriteString("\n\n")

	b.WriteString("  Nickname:\n")
	b.WriteString("  " + m.nicknameInput.View())
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString("  " + m.spinner.View() + " Looking up tracking number...")
		b.WriteString("\n")
	}

	if m.err != nil {
		b.WriteString("  " + ErrorStyle.Render("Error: "+m.err.Error()))
		b.WriteString("\n")
	}

	help := "  tab switch fields  enter submit  esc cancel"
	b.WriteString(HelpStyle.Render(help))

	return b.String()
}
