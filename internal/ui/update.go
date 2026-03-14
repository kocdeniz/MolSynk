package ui

import (
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	imapconn "imapsync/internal/imap"
	"imapsync/internal/sync"
)

// Init — start at the intro screen.
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update — global resize + connection messages, then phase dispatch.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// ---- Global: terminal resize -----------------------------------------
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.Width = ws.Width
		m.Height = ws.Height
		m.Progress.Width = ws.Width - 8
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
	default:
		return m.updateDash(msg)
	}
}

// ============================================================
// PhaseIntro
// ============================================================

func (m Model) updateIntro(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
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
	m.State = StateConnecting
	m.SrcState = ConnConnecting
	m.DstState = ConnConnecting

	return m, sync.Connect(
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

	// Build the account queue for the dashboard
	m.AccountQueue = make([]AccountState, len(pairs))
	for i, p := range pairs {
		m.AccountQueue[i] = AccountState{Username: p.SrcCfg.Username}
	}
	m.CurrentAccountIdx = 0
	m.ActiveAccount = pairs[0].SrcCfg.Username

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
		len(pairs), srcHost, dstHost,
	))

	cmd, ch := sync.RunMigration(pairs)
	m.StatusCh = ch
	return m, cmd
}

// ============================================================
// PhaseDash
// ============================================================

func (m Model) updateDash(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

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
		// Find and mark active in queue
		for i := range m.AccountQueue {
			if m.AccountQueue[i].Username == msg.Account {
				m.CurrentAccountIdx = i
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
				break
			}
		}
		m.AddLog(LogSuccess, fmt.Sprintf("[%s] complete", msg.Account))

	case sync.StatusAccountError:
		for i := range m.AccountQueue {
			if m.AccountQueue[i].Username == msg.Account {
				m.AccountQueue[i].Failed = true
				m.AccountQueue[i].ErrMsg = msg.Err.Error()
				break
			}
		}
		m.AddLog(LogError, fmt.Sprintf("[%s] error: %s", msg.Account, msg.Err))

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
		// Recompute total synced
		m.SyncedMessages = 0
		for _, f := range m.Folders {
			m.SyncedMessages += f.Synced
		}

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
		m.AddLog(LogSuccess, "All migrations complete.")
		// Update progress bar to 100%
		return m, m.Progress.SetPercent(1.0)
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
