package tui

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7c3aed")).
			Padding(0, 1)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#e2e8f0"))

	StatusDelivered = lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e"))
	StatusInTransit = lipgloss.NewStyle().Foreground(lipgloss.Color("#3b82f6"))
	StatusOutFor    = lipgloss.NewStyle().Foreground(lipgloss.Color("#eab308"))
	StatusAlert     = lipgloss.NewStyle().Foreground(lipgloss.Color("#ef4444"))
	StatusPreTransit = lipgloss.NewStyle().Foreground(lipgloss.Color("#9ca3af"))
	StatusUnknown   = lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))

	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#7c3aed"))

	DimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6b7280")).
			Padding(1, 0)

	BorderBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7c3aed")).
			Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ef4444"))

	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e"))

	EventDateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3b82f6")).
			Width(18)

	EventDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e2e8f0"))

	EventLocStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9ca3af"))
)

func StatusIcon(category string) string {
	switch category {
	case "delivered":
		return StatusDelivered.Render("●")
	case "in_transit", "in-transit":
		return StatusInTransit.Render("◐")
	case "out_for_delivery", "out-for-delivery":
		return StatusOutFor.Render("◑")
	case "alert", "returned":
		return StatusAlert.Render("⚠")
	case "pre_transit", "pre-transit":
		return StatusPreTransit.Render("○")
	default:
		return StatusUnknown.Render("?")
	}
}

func StatusLabel(category string) string {
	switch category {
	case "delivered":
		return StatusDelivered.Render("Delivered")
	case "in_transit", "in-transit":
		return StatusInTransit.Render("In Transit")
	case "out_for_delivery", "out-for-delivery":
		return StatusOutFor.Render("Out for Delivery")
	case "alert":
		return StatusAlert.Render("Alert")
	case "returned":
		return StatusAlert.Render("Returned")
	case "pre_transit", "pre-transit":
		return StatusPreTransit.Render("Pre-Transit")
	default:
		return StatusUnknown.Render("Unknown")
	}
}
