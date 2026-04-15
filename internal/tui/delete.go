package tui

import (
	"strings"

	"github.com/Subhrato20/track-/internal/db"
	tea "github.com/charmbracelet/bubbletea"
)

type deleteModel struct {
	pkg db.Package
}

func newDeleteModel() deleteModel {
	return deleteModel{}
}

func (m *deleteModel) SetPackage(pkg db.Package) {
	m.pkg = pkg
}

func (m deleteModel) Update(msg tea.Msg) (deleteModel, tea.Cmd) {
	return m, nil
}

func (m deleteModel) View() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(ErrorStyle.Render("  Delete package?"))
	b.WriteString("\n\n")

	b.WriteString("  Tracking #: " + m.pkg.TrackingNumber)
	b.WriteString("\n")
	if m.pkg.Nickname != "" {
		b.WriteString("  Nickname:   " + m.pkg.Nickname)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(DimStyle.Render("  This will remove the package and all tracking history."))
	b.WriteString("\n\n")

	help := "  y confirm  n/esc cancel"
	b.WriteString(HelpStyle.Render(help))

	return b.String()
}
