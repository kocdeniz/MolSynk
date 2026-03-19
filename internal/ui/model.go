package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	imapconn "github.com/kocdeniz/mailmole/internal/imap"
	"github.com/kocdeniz/mailmole/internal/sync"
	"github.com/kocdeniz/mailmole/internal/web"
)

// WebServerInterface defines the interface for web dashboard integration
type WebServerInterface interface {
	UpdateFromSyncMsg(msg sync.StatusUpdateMsg)
	AddLog(level, text string)
}

// ---- Connection state --------------------------------------------------------

type ConnState int

const (
	ConnIdle ConnState = iota
	ConnConnecting
	ConnReady
	ConnFailed
)

func (c ConnState) String() string {
	switch c {
	case ConnConnecting:
		return "Connecting..."
	case ConnReady:
		return "Ready"
	case ConnFailed:
		return "Failed"
	default:
		return "Idle"
	}
}

// ConnConfig holds resolved credentials for one IMAP endpoint.
type ConnConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	TLS      bool
}

// ---- Folder state ------------------------------------------------------------

type FolderState struct {
	Name     string
	Total    int
	Synced   int
	Done     bool
	Skipped  bool
	ErrorMsg string
}

// ---- Preview data structures --------------------------------------------------

// FolderPreview holds preview information for a single folder
type FolderPreview struct {
	Name         string
	MessageCount int
	SizeEstimate int64
	Selected     bool // user can toggle selection
}

// PreviewInfo holds the complete migration preview data
type PreviewInfo struct {
	SourceAccount        string
	DestAccount          string
	SourceHost           string
	DestHost             string
	Folders              []FolderPreview
	TotalMessages        int
	TotalSizeEstimate    int64
	EstimatedDuration    time.Duration
	SelectedFolders      int
	SelectedMessages     int
	SelectedSizeEstimate int64
	FocusedFolder        int // for navigation
}

// ---- Manual form field indices -----------------------------------------------

const (
	fieldSrcHost = iota
	fieldSrcUser
	fieldSrcPass
	fieldDstHost
	fieldDstUser
	fieldDstPass
	fieldCount // sentinel — always last
)

func fieldLabel(i int) string {
	switch i {
	case fieldSrcHost:
		return "Source Host:Port"
	case fieldSrcUser:
		return "Source Username"
	case fieldSrcPass:
		return "Source Password"
	case fieldDstHost:
		return "Dest Host:Port"
	case fieldDstUser:
		return "Dest Username"
	case fieldDstPass:
		return "Dest Password"
	}
	return ""
}

// ---- Bulk form field indices --------------------------------------------------

const (
	bulkFieldSrcHost = iota
	bulkFieldDstHost
	bulkFieldFile
	bulkFieldCount // sentinel
)

func bulkFieldLabel(i int) string {
	switch i {
	case bulkFieldSrcHost:
		return "Source Host:Port"
	case bulkFieldDstHost:
		return "Dest Host:Port"
	case bulkFieldFile:
		return "Accounts File"
	}
	return ""
}

// ---- Phase & state enums -----------------------------------------------------

// AppPhase is the top-level screen router.
type AppPhase int

const (
	PhaseIntro   AppPhase = iota // branding / splash
	PhaseSelect                  // choose Manual or Bulk
	PhaseManual                  // 6-field credential form
	PhaseBulk                    // bulk 3-field form
	PhasePreview                 // migration preview screen
	PhaseDash                    // migration dashboard
)

// AppState is the migration state machine inside the dashboard.
type AppState int

const (
	StateIdle AppState = iota
	StateConnecting
	StateSyncing
	StateDone
	StateError
)

// InputMode records which path the user chose on the selection screen.
type InputMode int

const (
	ModeNone   InputMode = iota
	ModeManual           // single account pair via form
	ModeBulk             // list of pairs via CSV/TXT file
)

// ---- Account queue state (Bulk mode) -----------------------------------------

// AccountState tracks per-account migration progress in bulk mode.
type AccountState struct {
	Username string
	Done     bool
	Failed   bool
	ErrMsg   string

	MigratedMessages int
	MigratedBytes    int64
	SkippedMessages  int
	FolderErrors     []string
}

// ---- Root model --------------------------------------------------------------

type Model struct {
	// Navigation
	Phase      AppPhase
	InputMode  InputMode
	IntroFrame int

	// ---- PhaseManual fields ------------------------------------------------
	Inputs       [fieldCount]textinput.Model
	FocusedField int
	SetupErr     string

	// ---- PhaseBulk fields (3-field form) -----------------------------------
	BulkInputs       [bulkFieldCount]textinput.Model
	BulkFocusedField int
	BulkErr          string

	// ---- Dashboard fields --------------------------------------------------

	// Single-account mode
	SrcConfig ConnConfig
	DstConfig ConnConfig
	SrcState  ConnState
	DstState  ConnState

	// Folder list and per-folder progress (both modes)
	Folders        []FolderState
	CurrentFolder  int
	TotalMessages  int
	SyncedMessages int

	// Bulk-mode account queue
	AccountQueue      []AccountState
	CurrentAccountIdx int
	ActiveAccount     string // username of account currently being migrated

	// ---- Preview Mode fields ------------------------------------------------
	Preview PreviewInfo

	// ---- Web Dashboard ------------------------------------------------------
	WebServer WebServerInterface

	// Status update channel — kept on model so Update can re-schedule reads
	StatusCh <-chan sync.StatusUpdateMsg

	Progress progress.Model
	Log      []LogEntry
	LogView  viewport.Model
	State    AppState

	// Live transfer speed metrics
	SpeedStartedAt  time.Time
	SpeedMsgCount   int
	SpeedBytesTotal int64
	SpeedMailsPerS  float64
	SpeedKBPerS     float64
	StateSaving     bool
	LogFilePath     string

	// Final summary aggregates for the completed state
	OverallStartedAt      time.Time
	OverallEndedAt        time.Time
	OverallMigratedMails  int
	OverallSkippedMails   int
	OverallTransferredB   int64
	OverallAvgMailsPerSec float64

	// Live IMAP connections (manual mode)
	SrcClient *imapconn.Client
	DstClient *imapconn.Client

	// Terminal dimensions
	Width  int
	Height int
}

// ---- Log helpers -------------------------------------------------------------

type LogEntry struct {
	Time  time.Time
	Text  string
	Level LogLevel
}

type LogLevel int

const (
	LogInfo LogLevel = iota
	LogWarn
	LogError
	LogSuccess
)

func (m *Model) AddLog(level LogLevel, text string) {
	const maxEntries = 200
	follow := m.LogView.AtBottom()
	entry := LogEntry{Time: time.Now(), Level: level, Text: text}
	m.Log = append(m.Log, entry)
	if len(m.Log) > maxEntries {
		m.Log = m.Log[len(m.Log)-maxEntries:]
	}
	m.appendLogToFile(entry)
	m.refreshLogViewport(follow)

	// Send to web dashboard if available
	m.SendLogToWeb(level, text)
}

func (m *Model) appendLogToFile(e LogEntry) {
	path := m.LogFilePath
	if path == "" {
		path = "mailmole.log"
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()

	level := "INFO"
	switch e.Level {
	case LogWarn:
		level = "WARN"
	case LogError:
		level = "ERROR"
	case LogSuccess:
		level = "SUCCESS"
	}
	_, _ = f.WriteString(fmt.Sprintf("%s [%s] %s\n", e.Time.Format(time.RFC3339), level, e.Text))
}

func (m *Model) refreshLogViewport(followBottom bool) {
	if m.LogView.Height <= 0 {
		m.LogView.Height = 8
	}
	if m.LogView.Width <= 0 {
		m.LogView.Width = 80
	}
	lines := make([]string, 0, len(m.Log))
	for _, e := range m.Log {
		lines = append(lines, fmt.Sprintf("[%s] %s", e.Time.Format("15:04:05"), e.Text))
	}
	m.LogView.SetContent(strings.Join(lines, "\n"))
	if followBottom {
		m.LogView.GotoBottom()
	}
}

func (m Model) currentVisibleLogText() string {
	if len(m.Log) == 0 {
		return ""
	}
	start := m.LogView.YOffset
	if start < 0 {
		start = 0
	}
	end := start + m.LogView.Height
	if end > len(m.Log) {
		end = len(m.Log)
	}
	if start >= len(m.Log) {
		start = len(m.Log) - 1
	}

	lines := make([]string, 0, end-start)
	for _, e := range m.Log[start:end] {
		lines = append(lines, fmt.Sprintf("[%s] %s", e.Time.Format("15:04:05"), e.Text))
	}
	return strings.Join(lines, "\n")
}

// ---- Constructor -------------------------------------------------------------

// NewModel returns a fully initialised Model starting at the intro screen.
func NewModel(prog progress.Model) Model {
	// Manual credential inputs (6 fields)
	var inputs [fieldCount]textinput.Model
	for i := range inputs {
		t := textinput.New()
		t.Prompt = ""
		t.CharLimit = 256
		switch i {
		case fieldSrcHost:
			t.Placeholder = "mail.example.com:993"
		case fieldSrcUser:
			t.Placeholder = "user@example.com"
		case fieldSrcPass:
			t.Placeholder = "password"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '*'
		case fieldDstHost:
			t.Placeholder = "mail.dest.com:993"
		case fieldDstUser:
			t.Placeholder = "user@dest.com"
		case fieldDstPass:
			t.Placeholder = "password"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '*'
		}
		inputs[i] = t
	}
	inputs[0].Focus()

	// Bulk form inputs (3 fields: src host, dst host, file path)
	var bulkInputs [bulkFieldCount]textinput.Model
	for i := range bulkInputs {
		t := textinput.New()
		t.Prompt = ""
		t.CharLimit = 512
		switch i {
		case bulkFieldSrcHost:
			t.Placeholder = "mail.source.com:993"
		case bulkFieldDstHost:
			t.Placeholder = "mail.dest.com:993"
		case bulkFieldFile:
			t.Placeholder = "/path/to/accounts.csv"
		}
		bulkInputs[i] = t
	}
	bulkInputs[0].Focus()

	return Model{
		Phase:       PhaseIntro,
		Inputs:      inputs,
		BulkInputs:  bulkInputs,
		LogView:     viewport.New(80, 8),
		LogFilePath: "mailmole.log",
		Progress:    prog,
	}
}

// SetWebServer sets the web server for dashboard integration
func (m *Model) SetWebServer(server *web.Server) {
	m.WebServer = server
}

// SendLogToWeb sends a log entry to the web dashboard if available
func (m *Model) SendLogToWeb(level LogLevel, text string) {
	if m.WebServer != nil {
		levelStr := "INFO"
		switch level {
		case LogWarn:
			levelStr = "WARN"
		case LogError:
			levelStr = "ERROR"
		case LogSuccess:
			levelStr = "SUCCESS"
		}
		m.WebServer.AddLog(levelStr, text)
	}
}
