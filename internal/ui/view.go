package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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
	default:
		return m.viewDash()
	}
}

// ============================================================
// PhaseIntro
// ============================================================

const molsynkArt = `
   /\_____/\
  /  o   o  \
 ( ==  ^  == )
  )         (
 (           )
( (  )   (  ) )
(__(__)___(__)__)
`

func (m Model) viewIntro() string {
	content := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().Foreground(colorPrimary).Bold(true).Render(molsynkArt),
		lipgloss.NewStyle().Foreground(colorFg).Bold(true).MarginTop(1).Render("M O L S Y N K"),
		lipgloss.NewStyle().Foreground(colorMuted).Render("High-performance IMAP Migration Tool"),
		lipgloss.NewStyle().Foreground(colorBorder).Render("v0.1.0"),
		lipgloss.NewStyle().Foreground(colorMuted).MarginTop(2).Render("Press any key to continue   [q] Quit"),
	)
	return lipgloss.NewStyle().
		Width(m.Width).Height(m.Height).
		Align(lipgloss.Center, lipgloss.Center).
		Background(colorBg).
		Render(content)
}

// ============================================================
// PhaseSelect
// ============================================================

func (m Model) viewSelect() string {
	w := clamp(m.Width-8, 50, 90)
	body := lipgloss.JoinVertical(lipgloss.Left,
		TitleStyle.Render("MOLSYNK  --  Select Migration Mode"),
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
	sb.WriteString(TitleStyle.Render("MOLSYNK  --  Manual Credentials") + "\n")
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

	sb.WriteString(TitleStyle.Render("MOLSYNK  --  Bulk Migration") + "\n")
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

func (m Model) renderDashTitle() string {
	mode := ""
	switch m.InputMode {
	case ModeBulk:
		mode = "  [Bulk]"
	case ModeManual:
		mode = "  [Manual]"
	}
	return TitleStyle.Render("MOLSYNK  --  Migration Dashboard" + mode)
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
	return PanelStyle.Width(m.Width - 8).Render(
		header + "\n" + bar + "\n" +
			lipgloss.NewStyle().Foreground(colorWarning).Render(m.stateLabel()),
	)
}

func (m Model) renderLog() string {
	const visible = 6
	header := SectionLabelStyle.Render("ACTIVITY LOG")
	if len(m.Log) == 0 {
		return LogStyle.Width(m.Width - 8).Render(
			header + "\n" + lipgloss.NewStyle().Foreground(colorMuted).Render("  No activity yet."),
		)
	}
	start := len(m.Log) - visible
	if start < 0 {
		start = 0
	}
	var sb strings.Builder
	sb.WriteString(header + "\n")
	for _, e := range m.Log[start:] {
		style := LogLineStyle
		switch e.Level {
		case LogWarn:
			style = style.Foreground(colorWarning)
		case LogError:
			style = style.Foreground(colorError)
		case LogSuccess:
			style = style.Foreground(colorSuccess)
		default:
			style = style.Foreground(colorFg)
		}
		sb.WriteString("  " + style.Render(e.Text) + "\n")
	}
	return LogStyle.Width(m.Width - 8).Render(strings.TrimRight(sb.String(), "\n"))
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
