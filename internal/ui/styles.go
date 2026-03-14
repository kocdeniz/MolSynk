package ui

import "github.com/charmbracelet/lipgloss"

// Colour palette
const (
	colorPrimary = lipgloss.Color("#5FAFFF")
	colorSuccess = lipgloss.Color("#5FAF5F")
	colorWarning = lipgloss.Color("#FFAF5F")
	colorError   = lipgloss.Color("#FF5F5F")
	colorMuted   = lipgloss.Color("#6C6C6C")
	colorBorder  = lipgloss.Color("#3A3A3A")
	colorFg      = lipgloss.Color("#EEEEEE")
	colorBg      = lipgloss.Color("#1C1C1C")
)

// ---- Structural styles -------------------------------------------------------

var (
	AppStyle = lipgloss.NewStyle().
			Background(colorBg).
			Foreground(colorFg).
			Padding(1, 2)

	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	TitleStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			MarginBottom(1)

	SectionLabelStyle = lipgloss.NewStyle().
				Foreground(colorMuted).
				Bold(true)

	// Log panel
	LogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1).
			Height(8)

	LogLineStyle = lipgloss.NewStyle().
			Foreground(colorFg)

	LogTimestampStyle = lipgloss.NewStyle().
				Foreground(colorMuted)
)

// ---- Status badge helpers ----------------------------------------------------

func StatusBadge(connected bool) string {
	if connected {
		return lipgloss.NewStyle().Foreground(colorSuccess).Render("[CONNECTED]")
	}
	return lipgloss.NewStyle().Foreground(colorError).Render("[DISCONNECTED]")
}

func FolderStatusIcon(done bool) string {
	if done {
		return lipgloss.NewStyle().Foreground(colorSuccess).Render("[+]")
	}
	return lipgloss.NewStyle().Foreground(colorMuted).Render("[ ]")
}
