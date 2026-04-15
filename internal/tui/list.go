package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/Subhrato20/track-/internal/db"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type listModel struct {
	packages   []db.Package
	cursor     int
	width      int
	height     int
	refreshing bool
	spinner    spinner.Model
	err        error
}

func newListModel() listModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7c3aed"))

	return listModel{
		spinner: s,
	}
}

func (m listModel) Update(msg tea.Msg) (listModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.packages)-1 {
				m.cursor++
			}
		}

	case PackagesLoadedMsg:
		m.packages = msg.Packages
		if m.cursor >= len(m.packages) {
			m.cursor = max(0, len(m.packages)-1)
		}

	case TrackingUpdatedMsg:
		for i, p := range m.packages {
			if p.TrackingNumber == msg.TrackingNumber {
				m.packages[i] = msg.Package
				break
			}
		}

	case PackageAddedMsg:
		m.packages = append([]db.Package{msg.Package}, m.packages...)
		m.cursor = 0

	case PackageDeletedMsg:
		for i, p := range m.packages {
			if p.TrackingNumber == msg.TrackingNumber {
				m.packages = append(m.packages[:i], m.packages[i+1:]...)
				if m.cursor >= len(m.packages) {
					m.cursor = max(0, len(m.packages)-1)
				}
				break
			}
		}

	case AllRefreshedMsg:
		m.refreshing = false

	case spinner.TickMsg:
		if m.refreshing {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m listModel) View() string {
	var b strings.Builder

	title := BorderBox.Render(TitleStyle.Render("📦 Package Tracker"))
	b.WriteString(title)
	b.WriteString("\n\n")

	if len(m.packages) == 0 {
		b.WriteString(DimStyle.Render("  No packages tracked yet. Press 'a' to add one."))
		b.WriteString("\n")
	} else {
		header := fmt.Sprintf("  %-3s %-18s %-22s %-20s %s",
			"", "STATUS", "TRACKING #", "NICKNAME", "UPDATED")
		b.WriteString(HeaderStyle.Render(header))
		b.WriteString("\n")
		b.WriteString(DimStyle.Render("  " + strings.Repeat("─", min(m.width-4, 85))))
		b.WriteString("\n")

		for i, p := range m.packages {
			icon := StatusIcon(p.StatusCategory)
			label := StatusLabel(p.StatusCategory)

			trackNum := p.TrackingNumber
			if len(trackNum) > 20 {
				trackNum = trackNum[:20] + "…"
			}

			nickname := p.Nickname
			if len(nickname) > 18 {
				nickname = nickname[:18] + "…"
			}
			if nickname == "" {
				nickname = DimStyle.Render("—")
			}

			updated := formatTimeAgo(p.LastUpdated)

			line := fmt.Sprintf("  %s %-18s %-22s %-20s %s",
				icon, label, trackNum, nickname, updated)

			if i == m.cursor {
				line = SelectedStyle.Render(line)
			}

			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	if m.refreshing {
		b.WriteString("  " + m.spinner.View() + " Refreshing packages...")
		b.WriteString("\n")
	}

	if m.err != nil {
		b.WriteString("  " + ErrorStyle.Render("Error: "+m.err.Error()))
		b.WriteString("\n")
	}

	help := "  ↑/↓ navigate  a add  d delete  enter details  r refresh  q quit"
	b.WriteString(HelpStyle.Render(help))

	return b.String()
}

func (m listModel) SelectedPackage() *db.Package {
	if len(m.packages) == 0 {
		return nil
	}
	return &m.packages[m.cursor]
}

func formatTimeAgo(isoTime string) string {
	t, err := time.Parse(time.RFC3339, isoTime)
	if err != nil {
		return isoTime
	}

	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	default:
		return t.Format("Jan 2")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
