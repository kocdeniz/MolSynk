// mailmole - IMAP-to-IMAP mail migration tool
// Build with CGO_ENABLED=0 for maximum portability.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/kocdeniz/mailmole/internal/ui"
	"github.com/kocdeniz/mailmole/internal/web"
)

func main() {
	var (
		webAddr  = flag.String("web", "", "Enable web dashboard on specified address (e.g., :8080)")
		onlyWeb  = flag.Bool("web-only", false, "Run only web dashboard without TUI")
		showHelp = flag.Bool("help", false, "Show help")
		showVer  = flag.Bool("version", false, "Show version")
	)

	flag.Parse()

	if *showHelp {
		printHelp()
		return
	}

	if *showVer {
		fmt.Println("MailMole v1.0.0 - IMAP Migration Tool")
		return
	}

	// Start web dashboard if requested
	var webServer *web.Server
	if *webAddr != "" {
		webServer = web.NewServer(*webAddr)
		go func() {
			if err := webServer.Start(); err != nil {
				fmt.Fprintf(os.Stderr, "Web server error: %v\n", err)
			}
		}()
	}

	// Run TUI unless web-only mode
	if !*onlyWeb {
		prog := progress.New(
			progress.WithDefaultGradient(),
			progress.WithoutPercentage(),
		)

		model := ui.NewModel(prog)

		// Pass web server to model if available
		if webServer != nil {
			model.SetWebServer(webServer)
		}

		p := tea.NewProgram(
			model,
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)

		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	} else if webServer != nil {
		// Web-only mode: wait indefinitely
		fmt.Printf("Web dashboard running on http://localhost%s\n", *webAddr)
		fmt.Println("Press Ctrl+C to stop")
		select {}
	}
}

func printHelp() {
	fmt.Println(`MailMole - IMAP-to-IMAP Mail Migration Tool

Usage:
  mailmole [options]

Options:
  -web string       Enable web dashboard on specified address (default: disabled)
                    Example: -web :8080
  -web-only         Run only web dashboard without TUI (requires -web)
  -help             Show this help message
  -version          Show version information

Examples:
  mailmole                                    # Run TUI only (default)
  mailmole -web :8080                         # Run TUI with web dashboard
  mailmole -web :8080 -web-only               # Run web dashboard only

Features:
  ✓ Terminal UI with real-time progress
  ✓ Web dashboard for browser-based monitoring
  ✓ Manual and bulk migration modes
  ✓ Preview mode: review before migrating
  ✓ Checkpoint/resume support
  ✓ Duplicate detection
  ✓ Parallel folder processing

For more information, visit: https://github.com/kocdeniz/mailmole`)
}
