package ui

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/kocdeniz/mailmole/internal/buildinfo"
	syncpkg "github.com/kocdeniz/mailmole/internal/sync"
)

// View dispatches rendering to the active phase.
func (m Model) View() string {
	if m.Width == 0 {
		return "Loading..."
	}
	switch m.Phase {
	case PhaseIntro:
		return m.viewIntro()
	case PhaseSelect:
		return m.viewSelect()
	case PhaseManual:
		return m.viewManual()
	case PhaseBulk:
		return m.viewBulk()
	case PhasePreview:
		return m.viewPreview()
	default:
		return m.viewDash()
	}
}

// ============================================================
// PhaseIntro
// ============================================================

const mailmoleArt = `
   /\_____/\
  /  o   o  \
 ( ==  ^  == )
  )         (
 (           )
( (  )   (  ) )
(__(__)___(__)__)
`

var (
	introSignatureOnce sync.Once
	introSignatureArt  string
)

// getIntroSignature returns a terminal image (when supported) or ASCII fallback.
//
// Supported image mode: iTerm2 inline image protocol.
// Fallback: portable ASCII art for CentOS/SSH/generic terminals.
func getIntroSignature() string {
	introSignatureOnce.Do(func() {
		introSignatureArt = mailmoleArt

		if os.Getenv("MAILMOLE_NO_IMAGE") == "1" {
			return
		}
		if os.Getenv("TERM_PROGRAM") != "iTerm.app" {
			return
		}

		imgPath := "mailmole_header.png"
		data, err := os.ReadFile(imgPath)
		if err != nil {
			return
		}

		name := base64.StdEncoding.EncodeToString([]byte(filepath.Base(imgPath)))
		b64 := base64.StdEncoding.EncodeToString(data)

		// iTerm2 inline image escape sequence.
		introSignatureArt = "\x1b]1337;File=name=" + name + ";inline=1;width=80%;preserveAspectRatio=1:" + b64 + "\a"
	})

	return introSignatureArt
}

func (m Model) viewIntro() string {
	hero := getIntroSignature()
	heroStyle := lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
	if hero != mailmoleArt {
		heroStyle = lipgloss.NewStyle()
	} else {
		hero = renderAnimatedASCII(mailmoleArt, m.IntroFrame)
		heroStyle = lipgloss.NewStyle()
	}

	content := lipgloss.JoinVertical(lipgloss.Center,
		heroStyle.Render(hero),
		lipgloss.NewStyle().Foreground(colorFg).Bold(true).MarginTop(1).Render("M A I L M O L E"),
		lipgloss.NewStyle().Foreground(colorMuted).Render("High-performance IMAP Migration Tool"),
		lipgloss.NewStyle().Foreground(colorBorder).Render(buildinfo.IntroLabel()),
		lipgloss.NewStyle().Foreground(colorMuted).MarginTop(2).Render("Press any key to continue   [q] Quit"),
	)
	return lipgloss.NewStyle().
		Width(m.Width).Height(m.Height).
		Align(lipgloss.Center, lipgloss.Center).
		Background(colorBg).
		Render(content)
}

// renderAnimatedASCII applies a pink/cyan waterfall glow over the fallback
// ASCII logo. The highlight sweeps top-to-bottom and loops.
func renderAnimatedASCII(art string, frame int) string {
	trimmed := strings.Trim(art, "\n")
	lines := strings.Split(trimmed, "\n")
	if len(lines) == 0 {
		return art
	}

	highlight := frame % (len(lines) + 4)
	highlight -= 2

	var out []string
	for i, line := range lines {
		style := lineStyleFor(i, highlight)
		out = append(out, style.Render(line))
	}

	return "\n" + strings.Join(out, "\n") + "\n"
}

func lineStyleFor(i, highlight int) lipgloss.Style {
	base := lipgloss.NewStyle().Foreground(lipgloss.Color("#54E8FF"))
	if i%2 == 1 {
		base = base.Foreground(lipgloss.Color("#FF6BD6"))
	}

	switch {
	case i == highlight:
		return base.Foreground(lipgloss.Color("#C8FFFF")).Bold(true)
	case i == highlight-1 || i == highlight+1:
		return base.Foreground(lipgloss.Color("#8DF6FF")).Bold(true)
	case i == highlight-2 || i == highlight+2:
		return base.Foreground(lipgloss.Color("#B78DFF"))
	default:
		return base
	}
}

// ============================================================
// PhaseSelect
// ============================================================

func (m Model) viewSelect() string {
	w := clamp(m.Width-8, 50, 90)
	body := lipgloss.JoinVertical(lipgloss.Left,
		TitleStyle.Render("MAILMOLE  --  Select Migration Mode"),
		lipgloss.NewStyle().Foreground(colorMuted).Render("Choose how to provide account credentials.")+"\n",
		m.renderSelectOption("1", "Manual Entry",
			"Enter source and destination credentials\ninteractively via a 6-field form."),
		"\n",
		m.renderSelectOption("2", "Bulk Migration via File",
			"Provide global hosts and a .csv/.txt file with\nmultiple account pairs (user/pass only per line)."),
		"\n",
		lipgloss.NewStyle().Foreground(colorMuted).
			Render("[1] Manual   [2] Bulk File   [Esc/q] Back   [Ctrl+C] Quit"),
	)
	return AppStyle.Render(lipgloss.NewStyle().Width(w).Render(body))
}

func (m Model) renderSelectOption(key, label, desc string) string {
	keyBadge := lipgloss.NewStyle().
		Foreground(colorBg).Background(colorPrimary).Bold(true).
		Padding(0, 1).Render(key)
	inner := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Center,
			keyBadge,
			lipgloss.NewStyle().Foreground(colorFg).Bold(true).MarginLeft(1).Render(label),
		),
		lipgloss.NewStyle().Foreground(colorMuted).MarginLeft(5).Render(desc),
	)
	return PanelStyle.Render(inner)
}

// ============================================================
// PhaseManual
// ============================================================

func (m Model) viewManual() string {
	w := clamp(m.Width-8, 50, 100)
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render("MAILMOLE  --  Manual Credentials") + "\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(colorMuted).
		Render("Fill in both IMAP accounts, then press Enter on the last field.") + "\n\n")
	sb.WriteString(SectionLabelStyle.Render("SOURCE SERVER") + "\n")
	sb.WriteString(m.renderManualField(fieldSrcHost, w) + "\n")
	sb.WriteString(m.renderManualField(fieldSrcUser, w) + "\n")
	sb.WriteString(m.renderManualField(fieldSrcPass, w) + "\n\n")
	sb.WriteString(SectionLabelStyle.Render("DESTINATION SERVER") + "\n")
	sb.WriteString(m.renderManualField(fieldDstHost, w) + "\n")
	sb.WriteString(m.renderManualField(fieldDstUser, w) + "\n")
	sb.WriteString(m.renderManualField(fieldDstPass, w) + "\n\n")
	if m.SetupErr != "" {
		errBox := lipgloss.NewStyle().
			Foreground(colorError).
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorError).
			Padding(0, 1).
			Width(w - 4).
			Render("Connection failed: " + m.SetupErr)
		sb.WriteString(errBox + "\n\n")
	}

	switch m.State {
	case StateConnecting:
		sb.WriteString(lipgloss.NewStyle().Foreground(colorWarning).
			Bold(true).Render("  Connecting, please wait...") + "\n")
	default:
		hint := "  Tab/Down: next   Shift+Tab/Up: prev   Enter: connect   Esc: back"
		sb.WriteString(lipgloss.NewStyle().Foreground(colorMuted).Render(hint) + "\n")
	}
	return AppStyle.Render(sb.String())
}

func (m Model) renderManualField(idx int, w int) string {
	focused := m.FocusedField == idx
	labelStyle := SectionLabelStyle.Copy().Width(18)
	if focused {
		labelStyle = labelStyle.Foreground(colorPrimary)
	}
	borderColor := colorBorder
	if focused {
		borderColor = colorPrimary
	}
	inputBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).BorderForeground(borderColor).
		Width(w-24).Padding(0, 1).
		Render(m.Inputs[idx].View())
	return lipgloss.JoinHorizontal(lipgloss.Center,
		labelStyle.Render(fieldLabel(idx)+":"), "  ", inputBox)
}

// ============================================================
// PhaseBulk — 3-field form
// ============================================================

func (m Model) viewBulk() string {
	w := clamp(m.Width-8, 50, 90)
	var sb strings.Builder

	sb.WriteString(TitleStyle.Render("MAILMOLE  --  Bulk Migration") + "\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(colorMuted).
		Render("Enter global server settings and the path to your accounts file.") + "\n\n")

	// File format panel
	sb.WriteString(PanelStyle.Width(w-4).Render(
		SectionLabelStyle.Render("FILE FORMAT")+"\n"+
			lipgloss.NewStyle().Foreground(colorMuted).Render(
				"  Each line: src_user,src_pass,dst_user,dst_pass\n"+
					"  Accepted: .csv  .txt   Lines starting with # are comments.",
			),
	) + "\n\n")

	// 3 input fields
	sb.WriteString(SectionLabelStyle.Render("SERVER & FILE SETTINGS") + "\n")
	for i := range m.BulkInputs {
		sb.WriteString(m.renderBulkField(i, w) + "\n")
	}
	sb.WriteString("\n")

	if m.BulkErr != "" {
		sb.WriteString(lipgloss.NewStyle().Foreground(colorError).
			Render("  Error: "+m.BulkErr) + "\n\n")
	}
	sb.WriteString(lipgloss.NewStyle().Foreground(colorMuted).
		Render("  Tab/Down: next   Shift+Tab/Up: prev   Enter on last field: Start   Esc: back") + "\n")

	return AppStyle.Render(sb.String())
}

func (m Model) renderBulkField(idx int, w int) string {
	focused := m.BulkFocusedField == idx
	labelStyle := SectionLabelStyle.Copy().Width(18)
	if focused {
		labelStyle = labelStyle.Foreground(colorPrimary)
	}
	borderColor := colorBorder
	if focused {
		borderColor = colorPrimary
	}
	inputBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).BorderForeground(borderColor).
		Width(w-24).Padding(0, 1).
		Render(m.BulkInputs[idx].View())
	return lipgloss.JoinHorizontal(lipgloss.Center,
		labelStyle.Render(bulkFieldLabel(idx)+":"), "  ", inputBox)
}

// ============================================================
// PhaseDash
// ============================================================

func (m Model) viewDash() string {
	if m.State == StateDone {
		return m.renderFinalSummary()
	}

	sections := []string{
		m.renderDashTitle(),
		m.renderConnections(),
		m.renderAccountQueue(),
		m.renderFolders(),
		m.renderProgress(),
		m.renderLog(),
		m.renderDashFooter(),
	}
	return AppStyle.Render(strings.Join(sections, "\n"))
}

func (m Model) renderFinalSummary() string {
	totalAccounts := len(m.AccountQueue)
	if totalAccounts == 0 {
		totalAccounts = 1
	}
	processed := m.OverallMigratedMails + m.OverallSkippedMails
	gb := float64(m.OverallTransferredB) / (1024 * 1024 * 1024)

	elapsedMin := 0.0
	if !m.OverallStartedAt.IsZero() && !m.OverallEndedAt.IsZero() {
		elapsedMin = m.OverallEndedAt.Sub(m.OverallStartedAt).Minutes()
	}

	title := lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true).
		Render("MAILMOLE  --  Migration Complete")

	metricLabel := lipgloss.NewStyle().Foreground(colorMuted)
	metricValue := lipgloss.NewStyle().Foreground(colorSuccess).Bold(true)

	lines := []string{
		title,
		"",
		metricLabel.Render("Total Accounts:") + " " + metricValue.Render(fmt.Sprintf("%d", totalAccounts)),
		metricLabel.Render("Total Mails Processed:") + " " + metricValue.Render(fmt.Sprintf("%d", processed)),
		metricLabel.Render("Total Data Transferred:") + " " + metricValue.Render(fmt.Sprintf("%.2f GB", gb)),
		metricLabel.Render("Average Speed:") + " " + metricValue.Render(fmt.Sprintf("%.2f mails/s", m.OverallAvgMailsPerSec)),
		metricLabel.Render("Total Time Elapsed:") + " " + metricValue.Render(fmt.Sprintf("%.2f min", elapsedMin)),
		"",
		lipgloss.NewStyle().Foreground(colorWarning).Render("Press [q] to exit or [r] to open the log file"),
		"",
		lipgloss.NewStyle().Foreground(colorPrimary).Render("   /\\_____/\\"),
		lipgloss.NewStyle().Foreground(colorPrimary).Render("  (  Job Done! )"),
		lipgloss.NewStyle().Foreground(colorPrimary).Render("   \\_____/"),
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(1, 2).
		Width(clamp(m.Width-12, 52, 90)).
		Render(strings.Join(lines, "\n"))

	return lipgloss.NewStyle().
		Width(m.Width).
		Height(m.Height).
		Align(lipgloss.Center, lipgloss.Center).
		Background(colorBg).
		Render(box)
}

func (m Model) renderDashTitle() string {
	mode := ""
	switch m.InputMode {
	case ModeBulk:
		mode = "  [Bulk]"
	case ModeManual:
		mode = "  [Manual]"
	}
	return TitleStyle.Render("MAILMOLE  --  Migration Dashboard" + mode)
}

func (m Model) renderConnections() string {
	half := (m.Width - 10) / 2

	srcUser := m.SrcConfig.Username
	if srcUser == "" && m.ActiveAccount != "" {
		srcUser = m.ActiveAccount
	}
	srcLine := fmt.Sprintf("%s:%d", m.SrcConfig.Host, m.SrcConfig.Port)
	if srcUser != "" {
		srcLine += "  " + srcUser
	}
	srcPanel := PanelStyle.Width(half).Render(
		SectionLabelStyle.Render("SOURCE") + "  " + StatusBadge(m.SrcState == ConnReady) + "\n" +
			lipgloss.NewStyle().Foreground(colorMuted).Render(srcLine),
	)

	dstLine := fmt.Sprintf("%s:%d", m.DstConfig.Host, m.DstConfig.Port)
	dstPanel := PanelStyle.Width(half).Render(
		SectionLabelStyle.Render("DESTINATION") + "  " + StatusBadge(m.DstState == ConnReady) + "\n" +
			lipgloss.NewStyle().Foreground(colorMuted).Render(dstLine),
	)

	return lipgloss.JoinHorizontal(lipgloss.Top, srcPanel, "  ", dstPanel)
}

// renderAccountQueue is shown only in Bulk mode.
func (m Model) renderAccountQueue() string {
	if m.InputMode != ModeBulk || len(m.AccountQueue) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(SectionLabelStyle.Render("ACCOUNT QUEUE") + "\n")

	for i, a := range m.AccountQueue {
		var icon string
		nameStyle := lipgloss.NewStyle().Foreground(colorFg)
		switch {
		case a.Failed:
			icon = lipgloss.NewStyle().Foreground(colorError).Render("[!]")
			nameStyle = nameStyle.Foreground(colorError)
		case a.Done:
			icon = lipgloss.NewStyle().Foreground(colorSuccess).Render("[+]")
			nameStyle = nameStyle.Foreground(colorMuted)
		case i == m.CurrentAccountIdx && m.State == StateSyncing:
			icon = lipgloss.NewStyle().Foreground(colorPrimary).Render("[>]")
			nameStyle = nameStyle.Foreground(colorPrimary).Bold(true)
		default:
			icon = lipgloss.NewStyle().Foreground(colorMuted).Render("[ ]")
		}
		sb.WriteString(fmt.Sprintf("  %s %s\n", icon, nameStyle.Render(a.Username)))
	}

	return PanelStyle.Width(m.Width - 8).Render(strings.TrimRight(sb.String(), "\n"))
}

func (m Model) renderFolders() string {
	if len(m.Folders) == 0 {
		if m.State == StateSyncing {
			return PanelStyle.Width(m.Width - 8).Render(
				SectionLabelStyle.Render("FOLDERS") + "\n" +
					lipgloss.NewStyle().Foreground(colorMuted).Render("  Connecting to account..."),
			)
		}
		return PanelStyle.Width(m.Width - 8).Render(
			SectionLabelStyle.Render("FOLDERS") + "\n" +
				lipgloss.NewStyle().Foreground(colorMuted).Render("  Waiting for folder list..."),
		)
	}

	var sb strings.Builder
	activeLabel := ""
	if m.ActiveAccount != "" {
		activeLabel = "  " + lipgloss.NewStyle().Foreground(colorPrimary).Render(m.ActiveAccount)
	}
	sb.WriteString(SectionLabelStyle.Render("FOLDERS") + activeLabel + "\n")

	for i, f := range m.Folders {
		active := i == m.CurrentFolder && m.State == StateSyncing && !f.Done
		nameStyle := lipgloss.NewStyle().Foreground(colorFg)
		if active {
			nameStyle = nameStyle.Foreground(colorPrimary).Bold(true)
		}
		if f.Done {
			nameStyle = nameStyle.Foreground(colorMuted)
		}
		pct := ""
		if f.Total > 0 {
			ratio := float64(f.Synced) / float64(f.Total) * 100
			pct = fmt.Sprintf("  %3.0f%%  (%d/%d)", ratio, f.Synced, f.Total)
		}
		sb.WriteString(fmt.Sprintf("  %s %s%s\n",
			FolderStatusIcon(f.Done),
			nameStyle.Render(f.Name),
			lipgloss.NewStyle().Foreground(colorMuted).Render(pct),
		))
	}
	return PanelStyle.Width(m.Width - 8).Render(strings.TrimRight(sb.String(), "\n"))
}

func (m Model) renderProgress() string {
	// For bulk mode: show account N of M
	extra := ""
	if m.InputMode == ModeBulk && len(m.AccountQueue) > 0 {
		done := 0
		for _, a := range m.AccountQueue {
			if a.Done || a.Failed {
				done++
			}
		}
		extra = fmt.Sprintf("  Account %d/%d", done+1, len(m.AccountQueue))
	}

	header := fmt.Sprintf("%s  %d / %d messages%s",
		SectionLabelStyle.Render("OVERALL PROGRESS"),
		m.SyncedMessages, m.TotalMessages, extra,
	)
	bar := m.Progress.View()
	if bar == "" {
		bar = lipgloss.NewStyle().Foreground(colorMuted).Render("  (no active sync)")
	}
	speed := lipgloss.NewStyle().Foreground(colorMuted).Render(
		fmt.Sprintf("Speed: %.2f mails/s  %.1f KB/s", m.SpeedMailsPerS, m.SpeedKBPerS),
	)
	stateLine := ""
	if m.StateSaving {
		stateLine = "\n" + lipgloss.NewStyle().Foreground(colorWarning).Render("[STATE] saving checkpoint...")
	}
	return PanelStyle.Width(m.Width - 8).Render(
		header + "\n" + bar + "\n" + speed + stateLine + "\n" +
			lipgloss.NewStyle().Foreground(colorWarning).Render(m.stateLabel()),
	)
}

func (m Model) renderLog() string {
	header := SectionLabelStyle.Render("ACTIVITY LOG")
	if len(m.Log) == 0 {
		return LogStyle.Width(m.Width - 8).Render(
			header + "\n" + lipgloss.NewStyle().Foreground(colorMuted).Render("  No activity yet."),
		)
	}
	content := m.LogView.View()
	if strings.TrimSpace(content) == "" {
		content = lipgloss.NewStyle().Foreground(colorMuted).Render("  No activity yet.")
	}
	hints := lipgloss.NewStyle().Foreground(colorMuted).Render("[up/down/pgup/pgdown] scroll  [c] copy")
	return LogStyle.Width(m.Width - 8).Render(header + "\n" + content + "\n" + hints)
}

func (m Model) renderDashFooter() string {
	hints := []string{"[s] Start Sync (manual)", "[q] Quit"}
	if m.State == StateSyncing {
		hints = []string{"Syncing...   [q] Quit"}
	} else if m.State == StateDone {
		hints = []string{"Migration complete.   [q] Quit"}
	}
	return lipgloss.NewStyle().Foreground(colorMuted).MarginTop(1).
		Render(strings.Join(hints, "   "))
}

func (m Model) stateLabel() string {
	switch m.State {
	case StateConnecting:
		return "Connecting..."
	case StateSyncing:
		return "Syncing..."
	case StateDone:
		return "Done."
	case StateError:
		return "Error."
	default:
		return "Idle"
	}
}

// ============================================================
// PhasePreview - Migration Preview Screen
// ============================================================

func (m Model) viewPreview() string {
	w := clamp(m.Width-8, 60, 100)
	p := m.Preview

	// Header
	header := lipgloss.JoinVertical(lipgloss.Left,
		TitleStyle.Render("MAILMOLE  --  Migration Preview"),
		lipgloss.NewStyle().Foreground(colorMuted).Render("Review your migration before starting.")+"\n",
	)

	// Connection Info Panel
	connPanel := PanelStyle.Width(w - 4).Render(
		SectionLabelStyle.Render("CONNECTION DETAILS") + "\n" +
			lipgloss.NewStyle().Foreground(colorFg).Render(fmt.Sprintf("  Source:      %s@%s", p.SourceAccount, p.SourceHost)) + "\n" +
			lipgloss.NewStyle().Foreground(colorFg).Render(fmt.Sprintf("  Destination: %s@%s", p.DestAccount, p.DestHost)),
	)

	// Summary Stats Panel
	summaryContent := fmt.Sprintf(
		"Total Folders:      %d\n"+
			"Total Messages:     %d\n"+
			"Total Size:         %s\n"+
			"Selected Folders:   %d\n"+
			"Selected Messages:  %d\n"+
			"Selected Size:      %s\n"+
			"Estimated Duration: %s",
		len(p.Folders),
		p.TotalMessages,
		syncpkg.FormatSize(p.TotalSizeEstimate),
		p.SelectedFolders,
		p.SelectedMessages,
		syncpkg.FormatSize(p.SelectedSizeEstimate),
		syncpkg.FormatDuration(p.EstimatedDuration),
	)

	summaryPanel := PanelStyle.Width(w - 4).BorderForeground(colorPrimary).Render(
		SectionLabelStyle.Render("MIGRATION SUMMARY") + "\n" +
			lipgloss.NewStyle().Foreground(colorFg).Render(summaryContent),
	)

	// Folder List
	var folderLines []string
	folderLines = append(folderLines, SectionLabelStyle.Render("FOLDERS (Space to toggle, Enter to start)")+"\n")

	visibleStart := 0
	visibleEnd := len(p.Folders)
	maxVisible := m.Height - 25 // Reserve space for other elements

	if maxVisible > 0 && len(p.Folders) > maxVisible {
		// Simple scroll logic
		visibleStart = p.FocusedFolder - maxVisible/2
		if visibleStart < 0 {
			visibleStart = 0
		}
		visibleEnd = visibleStart + maxVisible
		if visibleEnd > len(p.Folders) {
			visibleEnd = len(p.Folders)
			visibleStart = visibleEnd - maxVisible
			if visibleStart < 0 {
				visibleStart = 0
			}
		}
	}

	for i := visibleStart; i < visibleEnd && i < len(p.Folders); i++ {
		f := p.Folders[i]
		checkbox := "[ ]"
		if f.Selected {
			checkbox = "[x]"
		}

		checkboxStyle := lipgloss.NewStyle().Foreground(colorMuted)
		if f.Selected {
			checkboxStyle = lipgloss.NewStyle().Foreground(colorSuccess)
		}

		nameStyle := lipgloss.NewStyle().Foreground(colorFg)
		if i == p.FocusedFolder {
			nameStyle = nameStyle.Foreground(colorPrimary).Bold(true)
		}

		infoStyle := lipgloss.NewStyle().Foreground(colorMuted)
		info := fmt.Sprintf("(%d msgs, %s)", f.MessageCount, syncpkg.FormatSize(f.SizeEstimate))

		line := fmt.Sprintf("  %s %s %s",
			checkboxStyle.Render(checkbox),
			nameStyle.Render(truncateString(f.Name, 25)),
			infoStyle.Render(info),
		)
		folderLines = append(folderLines, line)
	}

	if visibleEnd < len(p.Folders) {
		folderLines = append(folderLines, lipgloss.NewStyle().Foreground(colorMuted).Render(fmt.Sprintf("  ... and %d more folders", len(p.Folders)-visibleEnd)))
	}

	folderPanel := PanelStyle.Width(w - 4).Render(strings.Join(folderLines, "\n"))

	// Help text
	help := lipgloss.NewStyle().Foreground(colorMuted).Render(
		"[Space] Toggle selection  [a] Select all  [n] Select none  [↑/↓] Navigate  [Enter] Start migration  [Esc] Go back",
	)

	// Combine all sections
	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		connPanel,
		"",
		summaryPanel,
		"",
		folderPanel,
		"",
		help,
	)

	return AppStyle.Render(lipgloss.NewStyle().Width(w).Render(content))
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// ---- Utility -----------------------------------------------------------------

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
