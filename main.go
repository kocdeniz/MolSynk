// imapsync - IMAP-to-IMAP mail migration tool
// Build with CGO_ENABLED=0 for maximum portability.
package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"

	"imapsync/internal/ui"
)

func main() {
	prog := progress.New(
		progress.WithDefaultGradient(),
		progress.WithoutPercentage(),
	)

	p := tea.NewProgram(
		ui.NewModel(prog),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
