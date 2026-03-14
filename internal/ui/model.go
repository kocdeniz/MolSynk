package ui

import (
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	imapconn "imapsync/internal/imap"
	"imapsync/internal/sync"
)

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
	PhaseIntro  AppPhase = iota // branding / splash
	PhaseSelect                 // choose Manual or Bulk
	PhaseManual                 // 6-field credential form
	PhaseBulk                   // bulk 3-field form
	PhaseDash                   // migration dashboard
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
}

// ---- Root model --------------------------------------------------------------

type Model struct {
	// Navigation
	Phase     AppPhase
	InputMode InputMode

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

	// Status update channel — kept on model so Update can re-schedule reads
	StatusCh <-chan sync.StatusUpdateMsg

	Progress progress.Model
	Log      []LogEntry
	State    AppState

	// Live IMAP connections (manual mode)
	SrcClient *imapconn.Client
	DstClient *imapconn.Client

	// Terminal dimensions
	Width  int
	Height int
}

// ---- Log helpers -------------------------------------------------------------

type LogEntry struct {
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
	m.Log = append(m.Log, LogEntry{Level: level, Text: text})
	if len(m.Log) > maxEntries {
		m.Log = m.Log[len(m.Log)-maxEntries:]
	}
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
		Phase:      PhaseIntro,
		Inputs:     inputs,
		BulkInputs: bulkInputs,
		Progress:   prog,
	}
}
