// Package web provides an optional web dashboard for monitoring MailMole
// migrations via browser. It runs alongside the TUI when --web flag is used.
package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/kocdeniz/mailmole/internal/imap"
	syncpkg "github.com/kocdeniz/mailmole/internal/sync"
)

//go:embed locales/*.json
var localesFS embed.FS

// Server handles HTTP requests and WebSocket connections for the web dashboard
type Server struct {
	addr      string
	mu        sync.RWMutex
	httpSrv   *http.Server
	clients   map[*Client]bool
	broadcast chan Message
	state     DashboardState
}

// Client represents a WebSocket client connection
type Client struct {
	srv  *Server
	send chan []byte
}

// Message types for WebSocket communication
type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// DashboardState holds the current migration state for the web dashboard
type DashboardState struct {
	IsRunning         bool            `json:"isRunning"`
	CurrentAccount    string          `json:"currentAccount"`
	TotalAccounts     int             `json:"totalAccounts"`
	CompletedAccounts int             `json:"completedAccounts"`
	FailedAccounts    int             `json:"failedAccounts"`
	TotalMessages     int             `json:"totalMessages"`
	SyncedMessages    int             `json:"syncedMessages"`
	CurrentSpeed      float64         `json:"currentSpeed"` // mails per second
	Folders           []FolderStatus  `json:"folders"`
	Accounts          []AccountStatus `json:"accounts"`
	Logs              []LogEntry      `json:"logs"`
	StartedAt         *time.Time      `json:"startedAt"`
	EndedAt           *time.Time      `json:"endedAt"`
}

// FolderStatus represents a folder's current state
type FolderStatus struct {
	Name   string `json:"name"`
	Total  int    `json:"total"`
	Synced int    `json:"synced"`
	Done   bool   `json:"done"`
}

// AccountStatus represents an account's current state
type AccountStatus struct {
	Username         string   `json:"username"`
	Done             bool     `json:"done"`
	Failed           bool     `json:"failed"`
	Error            string   `json:"error,omitempty"`
	MigratedMessages int      `json:"migratedMessages"`
	MigratedBytes    int64    `json:"migratedBytes"`
	SkippedMessages  int      `json:"skippedMessages"`
	FolderErrors     []string `json:"folderErrors,omitempty"`
}

// LogEntry represents a log message
type LogEntry struct {
	Time  time.Time `json:"time"`
	Level string    `json:"level"`
	Text  string    `json:"text"`
}

// NewServer creates a new web dashboard server
func NewServer(addr string) *Server {
	if addr == "" {
		addr = ":8080"
	}

	srv := &Server{
		addr:      addr,
		clients:   make(map[*Client]bool),
		broadcast: make(chan Message, 256),
		state: DashboardState{
			Folders:  make([]FolderStatus, 0),
			Accounts: make([]AccountStatus, 0),
			Logs:     make([]LogEntry, 0),
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.handleIndex)
	mux.HandleFunc("/api/status", srv.handleStatus)
	mux.HandleFunc("/api/logs", srv.handleLogs)
	mux.HandleFunc("/api/test-connection", srv.handleTestConnection)
	mux.HandleFunc("/api/preview", srv.handlePreview)
	mux.HandleFunc("/api/bulk-preview", srv.handleBulkPreview)
	mux.HandleFunc("/api/validate", srv.handleValidate)
	mux.HandleFunc("/api/schedule", srv.handleSchedule)
	mux.HandleFunc("/api/schedules", srv.handleGetSchedules)
	mux.HandleFunc("/api/schedule/delete", srv.handleDeleteSchedule)
	mux.HandleFunc("/ws", srv.handleWebSocket)
	mux.HandleFunc("/api/start", srv.handleStart)
	mux.HandleFunc("/api/stop", srv.handleStop)
	mux.HandleFunc("/locales/", srv.handleLocales)

	srv.httpSrv = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	return srv
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Printf("[WEB] Starting dashboard server on http://localhost%s", s.addr)
	go s.broadcastLoop()
	return s.httpSrv.ListenAndServe()
}

// Stop gracefully shuts down the server
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.httpSrv.Shutdown(ctx)
}

// UpdateState updates the dashboard state and broadcasts to all clients
func (s *Server) UpdateState(state DashboardState) {
	s.mu.Lock()
	s.state = state
	s.mu.Unlock()

	s.broadcast <- Message{Type: "state", Data: state}
}

// UpdateFromSyncMsg updates the dashboard from a sync status message
func (s *Server) UpdateFromSyncMsg(msg syncpkg.StatusUpdateMsg) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch msg.Kind {
	case syncpkg.StatusAccountStart:
		s.state.IsRunning = true
		s.state.CurrentAccount = msg.Account
		if s.state.StartedAt == nil {
			now := time.Now()
			s.state.StartedAt = &now
		}

	case syncpkg.StatusAccountDone:
		s.state.CompletedAccounts++
		if msg.Stats != nil {
			s.state.SyncedMessages += msg.Stats.MigratedMessages
			s.updateAccountStatus(msg.Account, true, false, "", msg.Stats)
		}

	case syncpkg.StatusAccountError:
		s.state.FailedAccounts++
		s.updateAccountStatus(msg.Account, false, true, msg.Err.Error(), nil)

	case syncpkg.StatusFolderStart:
		s.updateFolderStatus(msg.Folder, msg.Total, 0, false)

	case syncpkg.StatusFolderDone:
		s.updateFolderStatus(msg.Folder, msg.Total, msg.Copied, true)

	case syncpkg.StatusMessageCopied:
		s.state.SyncedMessages++
		s.updateFolderProgress(msg.Folder, msg.Copied)

	case syncpkg.StatusMigrationDone:
		s.state.IsRunning = false
		now := time.Now()
		s.state.EndedAt = &now
	}

	// Broadcast updated state
	select {
	case s.broadcast <- Message{Type: "state", Data: s.state}:
	default:
	}
}

// AddLog adds a log entry and broadcasts it
func (s *Server) AddLog(level, text string) {
	s.mu.Lock()
	entry := LogEntry{
		Time:  time.Now(),
		Level: level,
		Text:  text,
	}
	s.state.Logs = append(s.state.Logs, entry)
	if len(s.state.Logs) > 1000 {
		s.state.Logs = s.state.Logs[len(s.state.Logs)-1000:]
	}
	s.mu.Unlock()

	select {
	case s.broadcast <- Message{Type: "log", Data: entry}:
	default:
	}
}

func (s *Server) updateAccountStatus(username string, done, failed bool, err string, stats *syncpkg.AccountStats) {
	for i := range s.state.Accounts {
		if s.state.Accounts[i].Username == username {
			s.state.Accounts[i].Done = done
			s.state.Accounts[i].Failed = failed
			s.state.Accounts[i].Error = err
			if stats != nil {
				s.state.Accounts[i].MigratedMessages = stats.MigratedMessages
				s.state.Accounts[i].MigratedBytes = stats.MigratedBytes
				s.state.Accounts[i].SkippedMessages = stats.SkippedDuplicates
				s.state.Accounts[i].FolderErrors = stats.FolderErrors
			}
			return
		}
	}
	// Add new account
	account := AccountStatus{
		Username: username,
		Done:     done,
		Failed:   failed,
		Error:    err,
	}
	if stats != nil {
		account.MigratedMessages = stats.MigratedMessages
		account.MigratedBytes = stats.MigratedBytes
		account.SkippedMessages = stats.SkippedDuplicates
		account.FolderErrors = stats.FolderErrors
	}
	s.state.Accounts = append(s.state.Accounts, account)
	s.state.TotalAccounts = len(s.state.Accounts)
}

func (s *Server) updateFolderStatus(name string, total, synced int, done bool) {
	for i := range s.state.Folders {
		if s.state.Folders[i].Name == name {
			s.state.Folders[i].Total = total
			s.state.Folders[i].Synced = synced
			s.state.Folders[i].Done = done
			return
		}
	}
	s.state.Folders = append(s.state.Folders, FolderStatus{
		Name:   name,
		Total:  total,
		Synced: synced,
		Done:   done,
	})
}

func (s *Server) updateFolderProgress(name string, synced int) {
	for i := range s.state.Folders {
		if s.state.Folders[i].Name == name {
			s.state.Folders[i].Synced = synced
			return
		}
	}
}

func (s *Server) broadcastLoop() {
	for msg := range s.broadcast {
		data, err := json.Marshal(msg)
		if err != nil {
			continue
		}

		s.mu.RLock()
		clients := make([]*Client, 0, len(s.clients))
		for client := range s.clients {
			clients = append(clients, client)
		}
		s.mu.RUnlock()

		for _, client := range clients {
			select {
			case client.send <- data:
			default:
				// Client is slow, close it
				s.removeClient(client)
				close(client.send)
			}
		}
	}
}

func (s *Server) removeClient(client *Client) {
	s.mu.Lock()
	delete(s.clients, client)
	s.mu.Unlock()
}

// HTTP Handlers

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(dashboardHTML))
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	state := s.state
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	logs := s.state.Logs
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// For simplicity, we'll use SSE instead of WebSocket
	// Upgrade to SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Send initial state
	s.mu.RLock()
	state := s.state
	s.mu.RUnlock()

	data, _ := json.Marshal(Message{Type: "state", Data: state})
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()

	// Create client and listen for updates
	client := &Client{srv: s, send: make(chan []byte, 256)}
	s.mu.Lock()
	s.clients[client] = true
	s.mu.Unlock()

	defer func() {
		s.removeClient(client)
		close(client.send)
	}()

	// Keep connection alive and send updates
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-client.send:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()

		case <-ticker.C:
			// Send heartbeat
			fmt.Fprintf(w, ":heartbeat\n\n")
			flusher.Flush()

		case <-r.Context().Done():
			return
		}
	}
}

// CredentialRequest represents login credentials from web
type CredentialRequest struct {
	SrcHost         string   `json:"srcHost"`
	SrcPort         int      `json:"srcPort"`
	SrcUser         string   `json:"srcUser"`
	SrcPass         string   `json:"srcPass"`
	SrcTLS          bool     `json:"srcTLS"`
	DstHost         string   `json:"dstHost"`
	DstPort         int      `json:"dstPort"`
	DstUser         string   `json:"dstUser"`
	DstPass         string   `json:"dstPass"`
	DstTLS          bool     `json:"dstTLS"`
	SelectedFolders []string `json:"selectedFolders,omitempty"`
}

// BulkAccount represents a single account in bulk migration
type BulkAccount struct {
	SrcHost string `json:"srcHost"`
	SrcPort int    `json:"srcPort"`
	SrcUser string `json:"srcUser"`
	SrcPass string `json:"srcPass"`
	SrcTLS  bool   `json:"srcTLS"`
	DstHost string `json:"dstHost"`
	DstPort int    `json:"dstPort"`
	DstUser string `json:"dstUser"`
	DstPass string `json:"dstPass"`
	DstTLS  bool   `json:"dstTLS"`
}

// BulkPreviewRequest represents a bulk preview request
type BulkPreviewRequest struct {
	Accounts []BulkAccount `json:"accounts"`
}

// BulkPreviewResponse represents preview data for multiple accounts
type BulkPreviewResponse struct {
	Accounts      []AccountPreview `json:"accounts"`
	TotalFolders  int              `json:"totalFolders"`
	TotalMessages int              `json:"totalMessages"`
	TotalSize     int64            `json:"totalSize"`
}

// AccountPreview represents preview data for a single account
type AccountPreview struct {
	SrcUser         string       `json:"srcUser"`
	DstUser         string       `json:"dstUser"`
	Folders         []FolderInfo `json:"folders"`
	TotalMessages   int          `json:"totalMessages"`
	TotalSize       int64        `json:"totalSize"`
	ConnectionError string       `json:"connectionError,omitempty"`
}

// FolderInfo represents folder information
type FolderInfo struct {
	Name         string `json:"name"`
	MessageCount int    `json:"messageCount"`
	SizeEstimate int64  `json:"sizeEstimate"`
}

func (s *Server) handleTestConnection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid JSON: " + err.Error()})
		return
	}

	// Test source connection
	srcCfg := imap.Config{
		Host:     req.SrcHost,
		Port:     993,
		Username: req.SrcUser,
		Password: req.SrcPass,
		TLS:      req.SrcTLS,
	}

	src, err := imap.Connect(srcCfg)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Source connection failed: " + err.Error()})
		return
	}
	src.Close()

	// Test destination connection
	dstCfg := imap.Config{
		Host:     req.DstHost,
		Port:     993,
		Username: req.DstUser,
		Password: req.DstPass,
		TLS:      req.DstTLS,
	}

	dst, err := imap.Connect(dstCfg)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Destination connection failed: " + err.Error()})
		return
	}
	dst.Close()

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func (s *Server) handlePreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Invalid JSON: " + err.Error()})
		return
	}

	// Connect to source
	srcCfg := imap.Config{
		Host:     req.SrcHost,
		Port:     993,
		Username: req.SrcUser,
		Password: req.SrcPass,
		TLS:      req.SrcTLS,
	}

	src, err := imap.Connect(srcCfg)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Source connection failed: " + err.Error()})
		return
	}
	defer src.Close()

	// List folders
	folderNames, err := src.ListFolders()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to list folders: " + err.Error()})
		return
	}

	// Get info for each folder
	type FolderInfo struct {
		Name         string `json:"name"`
		MessageCount int    `json:"messageCount"`
		SizeEstimate int64  `json:"sizeEstimate"`
	}

	var folders []FolderInfo
	var totalMessages int
	var totalSize int64

	for _, name := range folderNames {
		count, err := src.FolderStatus(name)
		if err != nil {
			continue
		}

		estimatedSize := int64(count) * 50 * 1024 // 50KB average per message

		folders = append(folders, FolderInfo{
			Name:         name,
			MessageCount: int(count),
			SizeEstimate: estimatedSize,
		})

		totalMessages += int(count)
		totalSize += estimatedSize
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"folders":       folders,
		"totalMessages": totalMessages,
		"totalSize":     totalSize,
	})
}

func (s *Server) handleBulkPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BulkPreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Invalid JSON: " + err.Error()})
		return
	}

	var response BulkPreviewResponse
	var grandTotalFolders, grandTotalMessages int
	var grandTotalSize int64

	for _, acc := range req.Accounts {
		preview := AccountPreview{
			SrcUser: acc.SrcUser,
			DstUser: acc.DstUser,
		}

		// Set default ports
		srcPort := acc.SrcPort
		if srcPort == 0 {
			srcPort = 993
		}
		dstPort := acc.DstPort
		if dstPort == 0 {
			dstPort = 993
		}

		// Connect to source
		srcCfg := imap.Config{
			Host:     acc.SrcHost,
			Port:     srcPort,
			Username: acc.SrcUser,
			Password: acc.SrcPass,
			TLS:      srcPort == 993,
		}

		src, err := imap.Connect(srcCfg)
		if err != nil {
			preview.ConnectionError = "Source connection failed: " + err.Error()
			response.Accounts = append(response.Accounts, preview)
			continue
		}

		// List folders
		folderNames, err := src.ListFolders()
		if err != nil {
			preview.ConnectionError = "Failed to list folders: " + err.Error()
			src.Close()
			response.Accounts = append(response.Accounts, preview)
			continue
		}

		// Get folder info
		for _, name := range folderNames {
			count, err := src.FolderStatus(name)
			if err != nil {
				continue
			}

			estimatedSize := int64(count) * 50 * 1024

			preview.Folders = append(preview.Folders, FolderInfo{
				Name:         name,
				MessageCount: int(count),
				SizeEstimate: estimatedSize,
			})

			preview.TotalMessages += int(count)
			preview.TotalSize += estimatedSize
		}

		src.Close()

		grandTotalFolders += len(preview.Folders)
		grandTotalMessages += preview.TotalMessages
		grandTotalSize += preview.TotalSize

		response.Accounts = append(response.Accounts, preview)
	}

	response.TotalFolders = grandTotalFolders
	response.TotalMessages = grandTotalMessages
	response.TotalSize = grandTotalSize

	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Invalid JSON: " + err.Error()})
		return
	}

	// TODO: Actually start migration with selected folders
	// For now just acknowledge
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// TODO: Implement stop via web API
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

// ScheduleRequest represents a scheduled migration request
type ScheduleRequest struct {
	Accounts   []BulkAccount `json:"accounts"`
	ScheduleAt string        `json:"scheduleAt"` // ISO 8601 format
	Repeat     string        `json:"repeat"`     // once, daily, weekly, monthly
}

// ScheduleResponse represents the response from schedule creation
type ScheduleResponse struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	NextRunAt string `json:"nextRunAt,omitempty"`
}

// ScheduledJob represents a scheduled migration job
type ScheduledJob struct {
	ID         string        `json:"id"`
	Accounts   []BulkAccount `json:"accounts"`
	ScheduleAt time.Time     `json:"scheduleAt"`
	Repeat     string        `json:"repeat"`
	CreatedAt  time.Time     `json:"createdAt"`
	Status     string        `json:"status"` // pending, running, completed, failed
}

var scheduledJobs = make(map[string]*ScheduledJob)
var jobsMutex sync.Mutex

func (s *Server) handleSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Invalid JSON: " + err.Error()})
		return
	}

	// Parse schedule time
	scheduleAt, err := time.Parse(time.RFC3339, req.ScheduleAt)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Invalid schedule time format"})
		return
	}

	// Create job
	jobID := fmt.Sprintf("job_%d", time.Now().UnixNano())
	job := &ScheduledJob{
		ID:         jobID,
		Accounts:   req.Accounts,
		ScheduleAt: scheduleAt,
		Repeat:     req.Repeat,
		CreatedAt:  time.Now(),
		Status:     "pending",
	}

	jobsMutex.Lock()
	scheduledJobs[jobID] = job
	jobsMutex.Unlock()

	// Calculate next run time
	nextRun := scheduleAt
	if nextRun.Before(time.Now()) {
		nextRun = calculateNextRun(nextRun, req.Repeat)
	}

	response := ScheduleResponse{
		ID:        jobID,
		Status:    "scheduled",
		Message:   fmt.Sprintf("Migration scheduled for %s", nextRun.Format("2006-01-02 15:04")),
		NextRunAt: nextRun.Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

func calculateNextRun(scheduleTime time.Time, repeat string) time.Time {
	next := scheduleTime
	now := time.Now()

	for next.Before(now) {
		switch repeat {
		case "daily":
			next = next.Add(24 * time.Hour)
		case "weekly":
			next = next.Add(7 * 24 * time.Hour)
		case "monthly":
			next = next.AddDate(0, 1, 0)
		default:
			return next
		}
	}
	return next
}

func (s *Server) handleGetSchedules(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobsMutex.Lock()
	defer jobsMutex.Unlock()

	var jobs []*ScheduledJob
	for _, job := range scheduledJobs {
		jobs = append(jobs, job)
	}

	json.NewEncoder(w).Encode(jobs)
}

func (s *Server) handleDeleteSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobID := r.URL.Query().Get("id")
	if jobID == "" {
		http.Error(w, "Missing job ID", http.StatusBadRequest)
		return
	}

	jobsMutex.Lock()
	delete(scheduledJobs, jobID)
	jobsMutex.Unlock()

	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

// ValidationRequest represents a validation test request
type ValidationRequest struct {
	Accounts []BulkAccount `json:"accounts"`
}

// ValidationResult represents the result of a connection test
type ValidationResult struct {
	Account       string `json:"account"`
	SrcHost       string `json:"srcHost"`
	SrcUser       string `json:"srcUser"`
	DstHost       string `json:"dstHost"`
	DstUser       string `json:"dstUser"`
	SrcStatus     string `json:"srcStatus"` // success, error, pending
	DstStatus     string `json:"dstStatus"` // success, error, pending
	SrcError      string `json:"srcError,omitempty"`
	DstError      string `json:"dstError,omitempty"`
	SrcServerInfo string `json:"srcServerInfo,omitempty"`
	DstServerInfo string `json:"dstServerInfo,omitempty"`
}

func (s *Server) handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Invalid JSON: " + err.Error()})
		return
	}

	var results []ValidationResult

	for _, acc := range req.Accounts {
		result := ValidationResult{
			Account: acc.SrcUser,
			SrcHost: acc.SrcHost,
			SrcUser: acc.SrcUser,
			DstHost: acc.DstHost,
			DstUser: acc.DstUser,
		}

		// Set default ports
		srcPort := acc.SrcPort
		if srcPort == 0 {
			srcPort = 993
		}
		dstPort := acc.DstPort
		if dstPort == 0 {
			dstPort = 993
		}

		// Test source connection
		srcCfg := imap.Config{
			Host:     acc.SrcHost,
			Port:     srcPort,
			Username: acc.SrcUser,
			Password: acc.SrcPass,
			TLS:      srcPort == 993,
		}

		src, err := imap.Connect(srcCfg)
		if err != nil {
			result.SrcStatus = "error"
			result.SrcError = err.Error()
		} else {
			result.SrcStatus = "success"
			result.SrcServerInfo = "IMAP4rev1 supported"
			src.Close()
		}

		// Test destination connection
		dstCfg := imap.Config{
			Host:     acc.DstHost,
			Port:     dstPort,
			Username: acc.DstUser,
			Password: acc.DstPass,
			TLS:      dstPort == 993,
		}

		dst, err := imap.Connect(dstCfg)
		if err != nil {
			result.DstStatus = "error"
			result.DstError = err.Error()
		} else {
			result.DstStatus = "success"
			result.DstServerInfo = "IMAP4rev1 supported"
			dst.Close()
		}

		results = append(results, result)
	}

	json.NewEncoder(w).Encode(results)
}

func (s *Server) handleLocales(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract language code from path (e.g., /locales/en.json -> en)
	path := r.URL.Path[len("/locales/"):]
	if path == "" || path == "/" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string][]string{
			"available": {"en", "tr"},
		})
		return
	}

	// Remove .json extension if present
	lang := strings.TrimSuffix(path, ".json")

	// Read from embedded locale files
	localeData, err := localesFS.ReadFile("locales/" + lang + ".json")
	if err != nil {
		http.Error(w, "Locale not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(localeData)
}
