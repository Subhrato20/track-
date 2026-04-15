package tui

import (
	"strings"
	"time"

	"github.com/Subhrato20/track-/internal/db"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type detailModel struct {
	pkg      db.Package
	events   []db.TrackingEvent
	viewport viewport.Model
	ready    bool
	width    int
	height   int
}

func newDetailModel() detailModel {
	return detailModel{}
}

func (m *detailModel) SetPackage(pkg db.Package) {
	m.pkg = pkg
	m.ready = false
}

func (m *detailModel) SetEvents(events []db.TrackingEvent) {
	m.events = events
	m.updateContent()
}

func (m *detailModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	if !m.ready {
		m.viewport = viewport.New(width-4, height-10)
		m.viewport.YPosition = 0
		m.ready = true
	} else {
		m.viewport.Width = width - 4
		m.viewport.Height = height - 10
	}
	m.updateContent()
}

func (m *detailModel) updateContent() {
	if !m.ready {
		return
	}

	var b strings.Builder

	if len(m.events) == 0 {
		b.WriteString(DimStyle.Render("  No tracking events yet."))
	} else {
		for _, e := range m.events {
			dateStr := formatEventDate(e.EventDate)
			b.WriteString(EventDateStyle.Render("  " + dateStr))
			b.WriteString("  ")
			b.WriteString(EventDescStyle.Render(e.EventDescription))
			b.WriteString("\n")

			loc := formatLocation(e.City, e.State, e.Zip)
			if loc != "" {
				b.WriteString(strings.Repeat(" ", 20))
				b.WriteString("  ")
				b.WriteString(EventLocStyle.Render(loc))
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}
	}

	m.viewport.SetContent(b.String())
}

func (m detailModel) Update(msg tea.Msg) (detailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case EventsLoadedMsg:
		m.events = msg.Events
		m.updateContent()
		return m, nil

	case tea.KeyMsg:
		// Let viewport handle scrolling
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m detailModel) View() string {
	var b strings.Builder

	b.WriteString(DimStyle.Render("  ← Back (esc)"))
	b.WriteString("\n\n")

	// Package header
	b.WriteString("  ")
	b.WriteString(TitleStyle.Render("Package: " + m.pkg.TrackingNumber))
	if m.pkg.Nickname != "" {
		b.WriteString("  " + DimStyle.Render("\""+m.pkg.Nickname+"\""))
	}
	b.WriteString("\n")

	b.WriteString("  Status:  ")
	b.WriteString(StatusIcon(m.pkg.StatusCategory))
	b.WriteString(" ")
	b.WriteString(StatusLabel(m.pkg.StatusCategory))
	b.WriteString("\n")

	if m.pkg.ExpectedDelivery != "" {
		b.WriteString("  Expected: ")
		b.WriteString(m.pkg.ExpectedDelivery)
		b.WriteString("\n")
	}

	if m.pkg.DestCity != "" || m.pkg.DestState != "" {
		b.WriteString("  Destination: ")
		b.WriteString(formatLocation(m.pkg.DestCity, m.pkg.DestState, ""))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(DimStyle.Render("  ─── Tracking History " + strings.Repeat("─", max(0, min(m.width-27, 60)))))
	b.WriteString("\n\n")

	if m.ready {
		b.WriteString(m.viewport.View())
	}

	b.WriteString("\n")
	help := "  ↑/↓ scroll  r refresh  esc back"
	b.WriteString(HelpStyle.Render(help))

	return b.String()
}

func formatEventDate(dateStr string) string {
	// Try common formats
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"January 2, 2006",
	} {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t.Format("Jan 2  3:04 PM")
		}
	}
	return dateStr
}

func formatLocation(city, state, zip string) string {
	parts := []string{}
	if city != "" {
		parts = append(parts, city)
	}
	if state != "" {
		parts = append(parts, state)
	}
	loc := strings.Join(parts, ", ")
	if zip != "" {
		loc += " " + zip
	}
	return loc
}
