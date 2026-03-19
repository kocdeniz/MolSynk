package ui

import (
	"fmt"
	"net"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	imapconn "github.com/kocdeniz/mailmole/internal/imap"
	"github.com/kocdeniz/mailmole/internal/sync"
)

// Init — start at the intro screen.
func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, introTick())
}

type introTickMsg time.Time

type logOpenResultMsg struct {
	Err error
}

func introTick() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg {
		return introTickMsg(t)
	})
}

func openLogFileCmd(path string) tea.Cmd {
	return func() tea.Msg {
		if strings.TrimSpace(path) == "" {
			path = "mailmole.log"
		}

		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", path)
		case "linux":
			cmd = exec.Command("xdg-open", path)
		case "windows":
			cmd = exec.Command("cmd", "/c", "start", path)
		default:
			return logOpenResultMsg{Err: fmt.Errorf("unsupported platform for auto-open")}
		}

		if err := cmd.Start(); err != nil {
			return logOpenResultMsg{Err: err}
		}
		return logOpenResultMsg{}
	}
}

// Update — global resize + connection messages, then phase dispatch.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// ---- Global: terminal resize -----------------------------------------
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.Width = ws.Width
		m.Height = ws.Height
		m.Progress.Width = ws.Width - 8
		lw := ws.Width - 14
		if lw < 20 {
			lw = 20
		}
		m.LogView.Width = lw
		m.LogView.Height = 8
		m.refreshLogViewport(false)
		return m, nil
	}

	// ---- Global: IMAP connection results ---------------------------------
	// These are fired from a tea.Cmd that may have been launched from any
	// phase. Handle them here so they are never lost due to a phase change.
	switch msg := msg.(type) {
	case sync.ConnectedMsg:
		return m.applyConnected(msg)

	case sync.ConnErrMsg:
		// Return the user to the manual form with the error shown.
		m.Phase = PhaseManual
		m.SrcState = ConnFailed
		m.DstState = ConnFailed
		m.State = StateIdle
		m.SetupErr = msg.Error()
		// Re-focus the first field so the user can correct the mistake.
		for i := range m.Inputs {
			m.Inputs[i].Blur()
		}
		m.FocusedField = 0
		m.Inputs[0].Focus()
		return m, textinput.Blink
	}

	switch m.Phase {
	case PhaseIntro:
		return m.updateIntro(msg)
	case PhaseSelect:
		return m.updateSelect(msg)
	case PhaseManual:
		return m.updateManual(msg)
	case PhaseBulk:
		return m.updateBulk(msg)
	case PhasePreview:
		return m.updatePreview(msg)
	default:
		return m.updateDash(msg)
	}
}

// ============================================================
// PhaseIntro
// ============================================================

func (m Model) updateIntro(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case introTickMsg:
		_ = msg
		m.IntroFrame++
		return m, introTick()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		default:
			m.Phase = PhaseSelect
			return m, nil
		}
	}

	return m, nil
}

// ============================================================
// PhaseSelect
// ============================================================

func (m Model) updateSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q", "esc":
			m.Phase = PhaseIntro
			return m, nil
		case "1":
			m.InputMode = ModeManual
			m.Phase = PhaseManual
			m.SetupErr = ""
			for i := range m.Inputs {
				m.Inputs[i].Blur()
			}
			m.FocusedField = 0
			m.Inputs[0].Focus()
			return m, textinput.Blink
		case "2":
			m.InputMode = ModeBulk
			m.Phase = PhaseBulk
			m.BulkErr = ""
			for i := range m.BulkInputs {
				m.BulkInputs[i].Blur()
			}
			m.BulkFocusedField = 0
			m.BulkInputs[0].Focus()
			return m, textinput.Blink
		}
	}
	return m, nil
}

// ============================================================
// PhaseManual — 6-field credential form
// ============================================================

func (m Model) updateManual(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.Phase = PhaseSelect
			m.SetupErr = ""
			return m, nil
		case "tab", "down":
			m = m.moveFocus(1)
			return m, textinput.Blink
		case "shift+tab", "up":
			m = m.moveFocus(-1)
			return m, textinput.Blink
		case "enter":
			if m.FocusedField == fieldCount-1 {
				return m.submitManualForm()
			}
			m = m.moveFocus(1)
			return m, textinput.Blink
		}

	}

	var cmd tea.Cmd
	m.Inputs[m.FocusedField], cmd = m.Inputs[m.FocusedField].Update(msg)
	return m, cmd
}

func (m Model) moveFocus(delta int) Model {
	m.Inputs[m.FocusedField].Blur()
	m.FocusedField = (m.FocusedField + delta + fieldCount) % fieldCount
	m.Inputs[m.FocusedField].Focus()
	return m
}

func (m Model) submitManualForm() (tea.Model, tea.Cmd) {
	m.SetupErr = ""

	srcHost, srcPort, err := splitHostPort(m.Inputs[fieldSrcHost].Value())
	if err != nil {
		m.SetupErr = "Source: " + err.Error()
		return m, nil
	}
	dstHost, dstPort, err := splitHostPort(m.Inputs[fieldDstHost].Value())
	if err != nil {
		m.SetupErr = "Destination: " + err.Error()
		return m, nil
	}

	srcUser := strings.TrimSpace(m.Inputs[fieldSrcUser].Value())
	srcPass := m.Inputs[fieldSrcPass].Value()
	dstUser := strings.TrimSpace(m.Inputs[fieldDstUser].Value())
	dstPass := m.Inputs[fieldDstPass].Value()

	if srcUser == "" || srcPass == "" || dstUser == "" || dstPass == "" {
		m.SetupErr = "All fields are required."
		return m, nil
	}

	srcTLS := srcPort == 993 || isIPAddr(srcHost)
	dstTLS := dstPort == 993 || isIPAddr(dstHost)

	// Store full config including password so the sync engine can re-connect.
	m.SrcConfig = ConnConfig{Host: srcHost, Port: srcPort, Username: srcUser, Password: srcPass, TLS: srcTLS}
	m.DstConfig = ConnConfig{Host: dstHost, Port: dstPort, Username: dstUser, Password: dstPass, TLS: dstTLS}

	// Go to Preview mode instead of directly connecting
	m.Phase = PhasePreview
	m.AddLog(LogInfo, "Fetching migration preview...")

	return m, sync.FetchPreview(
		imapconn.Config{Host: srcHost, Port: srcPort, Username: srcUser, Password: srcPass, TLS: srcTLS},
		imapconn.Config{Host: dstHost, Port: dstPort, Username: dstUser, Password: dstPass, TLS: dstTLS},
	)
}

// ============================================================
// PhaseBulk — 3-field form: src host, dst host, file path
// ============================================================

func (m Model) updateBulk(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.Phase = PhaseSelect
			m.BulkErr = ""
			for i := range m.BulkInputs {
				m.BulkInputs[i].Blur()
			}
			return m, nil
		case "tab", "down":
			m = m.moveBulkFocus(1)
			return m, textinput.Blink
		case "shift+tab", "up":
			m = m.moveBulkFocus(-1)
			return m, textinput.Blink
		case "enter":
			if m.BulkFocusedField == bulkFieldCount-1 {
				return m.submitBulkForm()
			}
			m = m.moveBulkFocus(1)
			return m, textinput.Blink
		}
	}

	var cmd tea.Cmd
	m.BulkInputs[m.BulkFocusedField], cmd = m.BulkInputs[m.BulkFocusedField].Update(msg)
	return m, cmd
}

func (m Model) moveBulkFocus(delta int) Model {
	m.BulkInputs[m.BulkFocusedField].Blur()
	m.BulkFocusedField = (m.BulkFocusedField + delta + bulkFieldCount) % bulkFieldCount
	m.BulkInputs[m.BulkFocusedField].Focus()
	return m
}

func (m Model) submitBulkForm() (tea.Model, tea.Cmd) {
	m.BulkErr = ""

	srcHost, srcPort, err := splitHostPort(m.BulkInputs[bulkFieldSrcHost].Value())
	if err != nil {
		m.BulkErr = "Source: " + err.Error()
		return m, nil
	}
	dstHost, dstPort, err := splitHostPort(m.BulkInputs[bulkFieldDstHost].Value())
	if err != nil {
		m.BulkErr = "Destination: " + err.Error()
		return m, nil
	}

	filePath := strings.TrimSpace(m.BulkInputs[bulkFieldFile].Value())
	if filePath == "" {
		m.BulkErr = "File path is required."
		return m, nil
	}
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".csv" && ext != ".txt" {
		m.BulkErr = "Only .csv or .txt files are accepted."
		return m, nil
	}

	srcTLS := srcPort == 993
	dstTLS := dstPort == 993

	pairs, err := sync.ParseQueueFile(filePath, srcHost, srcPort, srcTLS, dstHost, dstPort, dstTLS)
	if err != nil {
		m.BulkErr = err.Error()
		return m, nil
	}

	pending, skipped, statePath, err := sync.FilterCompletedAccounts(pairs)
	if err != nil {
		m.BulkErr = err.Error()
		return m, nil
	}
	if skipped > 0 {
		m.AddLog(LogInfo, fmt.Sprintf("[STATE] Loaded. Skipping %d completed accounts.", skipped))
	}
	if len(pending) == 0 {
		m.Phase = PhaseDash
		m.State = StateDone
		m.AddLog(LogSuccess, "No pending accounts. All accounts are already completed in state file.")
		return m, nil
	}

	// Build the account queue for the dashboard
	m.AccountQueue = make([]AccountState, len(pending))
	for i, p := range pending {
		m.AccountQueue[i] = AccountState{Username: p.SrcCfg.Username}
	}
	m.CurrentAccountIdx = 0
	m.ActiveAccount = pending[0].SrcCfg.Username

	// Global host display info
	m.SrcConfig = ConnConfig{Host: srcHost, Port: srcPort, TLS: srcTLS}
	m.DstConfig = ConnConfig{Host: dstHost, Port: dstPort, TLS: dstTLS}

	// Transition to dashboard and launch the real engine
	m.Phase = PhaseDash
	m.State = StateSyncing
	m.SrcState = ConnReady
	m.DstState = ConnReady
	m.AddLog(LogInfo, fmt.Sprintf(
		"Starting bulk migration: %d accounts from %s -> %s",
		len(pending), srcHost, dstHost,
	))

	cmd, ch := sync.RunMigrationWithCheckpoint(pending, statePath)
	m.StatusCh = ch
	return m, cmd
}

// ============================================================
// PhaseDash
// ============================================================

func (m Model) updateDash(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case logOpenResultMsg:
		if msg.Err != nil {
			m.AddLog(LogWarn, fmt.Sprintf("[LOG] Could not open log file: %v", msg.Err))
		} else {
			m.AddLog(LogInfo, "[LOG] Opening log file...")
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.SrcClient != nil {
				m.SrcClient.Close()
			}
			if m.DstClient != nil {
				m.DstClient.Close()
			}
			return m, tea.Quit
		case "s":
			// Manual mode: start mock-tick sync (real transfer via 's')
			if m.InputMode == ModeManual && m.State == StateIdle && len(m.Folders) > 0 {
				m.State = StateSyncing
				m.CurrentFolder = 0
				// Build a single-pair queue and launch the real engine
				pairs := []sync.AccountPair{{
					SrcCfg: imapconn.Config{
						Host:     m.SrcConfig.Host,
						Port:     m.SrcConfig.Port,
						Username: m.SrcConfig.Username,
						Password: m.SrcConfig.Password,
						TLS:      m.SrcConfig.TLS,
					},
					DstCfg: imapconn.Config{
						Host:     m.DstConfig.Host,
						Port:     m.DstConfig.Port,
						Username: m.DstConfig.Username,
						Password: m.DstConfig.Password,
						TLS:      m.DstConfig.TLS,
					},
				}}
				m.AddLog(LogInfo, "Starting migration...")
				cmd, ch := sync.RunMigration(pairs)
				m.StatusCh = ch
				return m, cmd
			}
		}

		if m.State == StateDone && msg.String() == "r" {
			return m, openLogFileCmd(m.LogFilePath)
		}

		// Clipboard copy of currently visible log lines.
		if msg.String() == "c" {
			text := m.currentVisibleLogText()
			if err := clipboard.WriteAll(text); err != nil {
				m.AddLog(LogWarn, fmt.Sprintf("[CLIPBOARD] Copy failed: %v", err))
			} else {
				m.AddLog(LogInfo, "[CLIPBOARD] Current log view copied.")
			}
			return m, nil
		}

		// Log viewport scrolling
		switch msg.String() {
		case "up", "down", "pgup", "pgdown", "home", "end":
			var cmd tea.Cmd
			m.LogView, cmd = m.LogView.Update(msg)
			return m, cmd
		}

	// ---- Real engine status updates ------------------------------------
	case sync.StatusUpdateMsg:
		return m.applyStatusUpdate(msg)

	// ---- Progress bar animation ----------------------------------------
	case progress.FrameMsg:
		prog, cmd := m.Progress.Update(msg)
		m.Progress = prog.(progress.Model)
		return m, cmd
	}

	return m, nil
}

// applyStatusUpdate dispatches a StatusUpdateMsg to the appropriate model fields
// and schedules the next read from the channel.
func (m Model) applyStatusUpdate(msg sync.StatusUpdateMsg) (tea.Model, tea.Cmd) {
	switch msg.Kind {

	case sync.StatusAccountStart:
		m.ActiveAccount = msg.Account
		if m.OverallStartedAt.IsZero() {
			m.OverallStartedAt = time.Now()
		}
		m.SpeedStartedAt = time.Now()
		m.SpeedMsgCount = 0
		m.SpeedBytesTotal = 0
		m.SpeedMailsPerS = 0
		m.SpeedKBPerS = 0
		// Find and mark active in queue
		for i := range m.AccountQueue {
			if m.AccountQueue[i].Username == msg.Account {
				m.CurrentAccountIdx = i
				m.AccountQueue[i].ErrMsg = ""
				m.AccountQueue[i].Failed = false
				m.AccountQueue[i].Done = false
				m.AccountQueue[i].MigratedMessages = 0
				m.AccountQueue[i].MigratedBytes = 0
				m.AccountQueue[i].SkippedMessages = 0
				m.AccountQueue[i].FolderErrors = nil
				break
			}
		}
		m.Folders = nil
		m.SyncedMessages = 0
		m.TotalMessages = 0
		m.AddLog(LogInfo, fmt.Sprintf("[%s] starting...", msg.Account))

	case sync.StatusAccountDone:
		for i := range m.AccountQueue {
			if m.AccountQueue[i].Username == msg.Account {
				m.AccountQueue[i].Done = true
				if msg.Stats != nil {
					m.AccountQueue[i].MigratedMessages = msg.Stats.MigratedMessages
					m.AccountQueue[i].MigratedBytes = msg.Stats.MigratedBytes
					m.AccountQueue[i].SkippedMessages = msg.Stats.SkippedDuplicates
					m.AccountQueue[i].FolderErrors = append([]string(nil), msg.Stats.FolderErrors...)
				}
				break
			}
		}
		m.AddLog(LogSuccess, fmt.Sprintf("[%s] complete", msg.Account))

	case sync.StatusAccountError:
		for i := range m.AccountQueue {
			if m.AccountQueue[i].Username == msg.Account {
				m.AccountQueue[i].Failed = true
				m.AccountQueue[i].ErrMsg = msg.Err.Error()
				if msg.Folder != "" {
					m.AccountQueue[i].FolderErrors = append(m.AccountQueue[i].FolderErrors, fmt.Sprintf("%s: %v", msg.Folder, msg.Err))
				}
				break
			}
		}
		m.AddLog(LogError, fmt.Sprintf("[%s] error: %s", msg.Account, msg.Err))

	case sync.StatusRetrying:
		waitSec := msg.RetryAfterS
		if waitSec <= 0 {
			waitSec = 2
		}
		errText := strings.ToLower(fmt.Sprint(msg.Err))
		if strings.Contains(errText, "closed network connection") || strings.Contains(errText, "timeout") {
			m.AddLog(LogWarn, "[NETWORK] Connection lost. Attempting auto-reconnection....")
		}
		m.AddLog(LogWarn, fmt.Sprintf("[RETRYING] Connection lost. Reconnecting in %d seconds... (%v)", waitSec, msg.Err))

	case sync.StatusFolderStart:
		// Add or reset the folder entry
		found := false
		for i := range m.Folders {
			if m.Folders[i].Name == msg.Folder {
				m.Folders[i].Total = msg.Total
				m.Folders[i].Synced = 0
				m.Folders[i].Done = false
				found = true
				break
			}
		}
		if !found {
			m.Folders = append(m.Folders, FolderState{Name: msg.Folder, Total: msg.Total})
			m.CurrentFolder = len(m.Folders) - 1
		}
		m.TotalMessages += msg.Total
		m.AddLog(LogInfo, fmt.Sprintf("[%s] folder %s (%d msgs)", msg.Account, msg.Folder, msg.Total))

	case sync.StatusMessageCopied:
		for i := range m.Folders {
			if m.Folders[i].Name == msg.Folder {
				m.Folders[i].Synced = msg.Copied
				break
			}
		}
		for i := range m.AccountQueue {
			if m.AccountQueue[i].Username == msg.Account {
				m.AccountQueue[i].MigratedMessages++
				m.AccountQueue[i].MigratedBytes += msg.MovedBytesDelta
				break
			}
		}
		m.SpeedMsgCount++
		m.SpeedBytesTotal += msg.MovedBytesDelta
		m.OverallMigratedMails++
		m.OverallTransferredB += msg.MovedBytesDelta
		if !m.SpeedStartedAt.IsZero() {
			elapsed := time.Since(m.SpeedStartedAt).Seconds()
			if elapsed > 0 {
				m.SpeedMailsPerS = float64(m.SpeedMsgCount) / elapsed
				m.SpeedKBPerS = (float64(m.SpeedBytesTotal) / 1024.0) / elapsed
			}
		}
		// Recompute total synced
		m.SyncedMessages = 0
		for _, f := range m.Folders {
			m.SyncedMessages += f.Synced
		}

	case sync.StatusMessageSkipped:
		for i := range m.AccountQueue {
			if m.AccountQueue[i].Username == msg.Account {
				m.AccountQueue[i].SkippedMessages += msg.SkippedDelta
				break
			}
		}
		m.OverallSkippedMails += msg.SkippedDelta

	case sync.StatusReportPlaced:
		m.AddLog(LogInfo, "[REPORT] Summary email placed in destination Inbox.")

	case sync.StatusStateSaving:
		m.StateSaving = true
		m.AddLog(LogInfo, "[STATE] Saving checkpoint...")

	case sync.StatusStateSaved:
		m.StateSaving = false
		m.AddLog(LogInfo, "[STATE] Checkpoint saved.")

	case sync.StatusFolderDone:
		for i := range m.Folders {
			if m.Folders[i].Name == msg.Folder {
				m.Folders[i].Done = true
				m.Folders[i].Synced = msg.Copied
				break
			}
		}
		m.AddLog(LogSuccess, fmt.Sprintf("[%s] folder %s done (%d msgs)", msg.Account, msg.Folder, msg.Copied))

	case sync.StatusMigrationDone:
		m.State = StateDone
		m.StateSaving = false
		m.OverallEndedAt = time.Now()
		if !m.OverallStartedAt.IsZero() {
			elapsed := m.OverallEndedAt.Sub(m.OverallStartedAt).Seconds()
			if elapsed > 0 {
				m.OverallAvgMailsPerSec = float64(m.OverallMigratedMails) / elapsed
			}
		}
		m.AddLog(LogSuccess, "All migrations complete.")
		// Send to web dashboard
		m.SendLogToWeb(LogSuccess, "All migrations complete.")
		if m.WebServer != nil {
			m.WebServer.UpdateFromSyncMsg(msg)
		}
		// Update progress bar to 100%
		return m, m.Progress.SetPercent(1.0)
	}

	// Send status update to web dashboard
	if m.WebServer != nil {
		m.WebServer.UpdateFromSyncMsg(msg)
	}

	// Update progress bar
	pct := 0.0
	if m.TotalMessages > 0 {
		pct = float64(m.SyncedMessages) / float64(m.TotalMessages)
	}

	// Schedule next status read — engine sends StatusMigrationDone when done
	if msg.Kind != sync.StatusMigrationDone {
		return m, tea.Batch(m.Progress.SetPercent(pct), sync.WaitForNext(m.StatusCh))
	}
	return m, m.Progress.SetPercent(pct)
}

// ============================================================
// Shared helpers
// ============================================================

func (m Model) applyConnected(msg sync.ConnectedMsg) (tea.Model, tea.Cmd) {
	m.Phase = PhaseDash
	m.SrcState = ConnReady
	m.DstState = ConnReady
	m.SrcClient = msg.Src
	m.DstClient = msg.Dst
	total := 0
	m.Folders = make([]FolderState, len(msg.Folders))
	for i, f := range msg.Folders {
		m.Folders[i] = FolderState{Name: f.Name, Total: f.Count}
		total += f.Count
	}
	m.TotalMessages = total
	m.State = StateIdle
	m.AddLog(LogSuccess, fmt.Sprintf(
		"Connected. %d folders, %d messages. Press [s] to start sync.",
		len(m.Folders), total,
	))
	return m, nil
}

func (m *Model) nextPendingFolder() int {
	for i, f := range m.Folders {
		if !f.Done {
			return i
		}
	}
	return -1
}

func splitHostPort(raw string) (host string, port int, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", 0, fmt.Errorf("host:port is required")
	}
	idx := strings.LastIndex(raw, ":")
	if idx < 0 {
		return "", 0, fmt.Errorf("expected host:port (e.g. mail.example.com:993)")
	}
	host = raw[:idx]
	port, err = strconv.Atoi(raw[idx+1:])
	if err != nil || port < 1 || port > 65535 {
		return "", 0, fmt.Errorf("invalid port number")
	}
	return host, port, nil
}

// isIPAddr returns true when host is a bare IPv4 or IPv6 address.
func isIPAddr(host string) bool {
	return net.ParseIP(host) != nil
}

// ============================================================
// PhasePreview - Migration Preview Screen
// ============================================================

func (m Model) updatePreview(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			// Go back to previous phase based on mode
			if m.InputMode == ModeManual {
				m.Phase = PhaseManual
			} else {
				m.Phase = PhaseBulk
			}
			return m, nil
		case "up":
			if m.Preview.FocusedFolder > 0 {
				m.Preview.FocusedFolder--
			}
			return m, nil
		case "down":
			if m.Preview.FocusedFolder < len(m.Preview.Folders)-1 {
				m.Preview.FocusedFolder++
			}
			return m, nil
		case " ":
			// Toggle selection
			if m.Preview.FocusedFolder < len(m.Preview.Folders) {
				f := &m.Preview.Folders[m.Preview.FocusedFolder]
				f.Selected = !f.Selected
				m.recalculatePreviewTotals()
			}
			return m, nil
		case "a":
			// Select all
			for i := range m.Preview.Folders {
				m.Preview.Folders[i].Selected = true
			}
			m.recalculatePreviewTotals()
			return m, nil
		case "n":
			// Select none
			for i := range m.Preview.Folders {
				m.Preview.Folders[i].Selected = false
			}
			m.recalculatePreviewTotals()
			return m, nil
		case "enter":
			// Start migration with selected folders
			return m.startMigrationFromPreview()
		}

	case sync.PreviewMsg:
		if msg.Error != nil {
			m.SetupErr = msg.Error.Error()
			// Return to previous phase
			if m.InputMode == ModeManual {
				m.Phase = PhaseManual
			} else {
				m.Phase = PhaseBulk
			}
			return m, nil
		}

		// Initialize preview data
		m.Preview.SourceAccount = msg.Data.SourceAccount
		m.Preview.DestAccount = msg.Data.DestAccount
		m.Preview.SourceHost = msg.Data.SourceHost
		m.Preview.DestHost = msg.Data.DestHost
		m.Preview.Folders = make([]FolderPreview, len(msg.Data.Folders))

		var totalMessages int
		var totalSize int64

		for i, f := range msg.Data.Folders {
			m.Preview.Folders[i] = FolderPreview{
				Name:         f.Name,
				MessageCount: f.MessageCount,
				SizeEstimate: f.SizeEstimate,
				Selected:     true, // Default: all selected
			}
			totalMessages += f.MessageCount
			totalSize += f.SizeEstimate
		}

		m.Preview.TotalMessages = totalMessages
		m.Preview.TotalSizeEstimate = totalSize
		m.Preview.SelectedFolders = len(msg.Data.Folders)
		m.Preview.SelectedMessages = totalMessages
		m.Preview.SelectedSizeEstimate = totalSize
		m.Preview.EstimatedDuration = sync.CalculateEstimatedDuration(totalMessages)
		m.Preview.FocusedFolder = 0

		return m, nil
	}

	return m, nil
}

func (m *Model) recalculatePreviewTotals() {
	var selectedFolders, selectedMessages int
	var selectedSize int64

	for _, f := range m.Preview.Folders {
		if f.Selected {
			selectedFolders++
			selectedMessages += f.MessageCount
			selectedSize += f.SizeEstimate
		}
	}

	m.Preview.SelectedFolders = selectedFolders
	m.Preview.SelectedMessages = selectedMessages
	m.Preview.SelectedSizeEstimate = selectedSize
	m.Preview.EstimatedDuration = sync.CalculateEstimatedDuration(selectedMessages)
}

func (m Model) startMigrationFromPreview() (tea.Model, tea.Cmd) {
	// Filter to only selected folders
	var selectedFolders []string
	for _, f := range m.Preview.Folders {
		if f.Selected {
			selectedFolders = append(selectedFolders, f.Name)
		}
	}

	if len(selectedFolders) == 0 {
		return m, nil // Don't start if nothing selected
	}

	// Transition to dashboard
	m.Phase = PhaseDash
	m.State = StateSyncing
	m.SrcState = ConnReady
	m.DstState = ConnReady

	// Build folder states from preview
	m.Folders = make([]FolderState, 0, len(selectedFolders))
	totalMessages := 0
	for _, f := range m.Preview.Folders {
		if f.Selected {
			m.Folders = append(m.Folders, FolderState{
				Name:  f.Name,
				Total: f.MessageCount,
			})
			totalMessages += f.MessageCount
		}
	}
	m.TotalMessages = totalMessages

	m.AddLog(LogInfo, fmt.Sprintf("Starting migration with %d folders, %d messages", len(selectedFolders), totalMessages))

	// Build account pairs
	if m.InputMode == ModeManual {
		pairs := []sync.AccountPair{{
			SrcCfg: imapconn.Config{
				Host:     m.SrcConfig.Host,
				Port:     m.SrcConfig.Port,
				Username: m.SrcConfig.Username,
				Password: m.SrcConfig.Password,
				TLS:      m.SrcConfig.TLS,
			},
			DstCfg: imapconn.Config{
				Host:     m.DstConfig.Host,
				Port:     m.DstConfig.Port,
				Username: m.DstConfig.Username,
				Password: m.DstConfig.Password,
				TLS:      m.DstConfig.TLS,
			},
		}}
		cmd, ch := sync.RunMigration(pairs)
		m.StatusCh = ch
		return m, cmd
	} else {
		// Bulk mode - pairs already built in bulk mode
		// This would need to be handled differently
		return m, nil
	}
}
