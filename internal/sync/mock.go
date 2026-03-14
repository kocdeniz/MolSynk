// Package sync contains the background synchronisation logic.
// This file provides a mock implementation used for UI development and testing.
package sync

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ---- Messages ----------------------------------------------------------------

// FolderListMsg carries the list of discovered folders from the source server.
type FolderListMsg struct {
	Folders []FolderInfo
}

// FolderInfo describes a single mailbox folder on the source server.
type FolderInfo struct {
	Name  string
	Count int // approximate message count
}

// ProgressMsg reports that one message was successfully copied.
type ProgressMsg struct {
	Folder string
	Copied int
	Total  int
	Done   bool
}

// LogMsg carries a log line to be appended in the TUI.
type LogMsg struct {
	Level int // 0=info 1=warn 2=error 3=success
	Text  string
}

// SyncDoneMsg signals that the full migration has completed.
type SyncDoneMsg struct{}

// ---- Commands ----------------------------------------------------------------

// MockDiscoverFolders simulates listing folders on the source server.
func MockDiscoverFolders() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(600 * time.Millisecond)
		return FolderListMsg{
			Folders: []FolderInfo{
				{Name: "INBOX", Count: 120},
				{Name: "Sent", Count: 340},
				{Name: "Drafts", Count: 12},
				{Name: "Trash", Count: 55},
				{Name: "Spam", Count: 8},
				{Name: "Archive/2023", Count: 600},
				{Name: "Archive/2024", Count: 430},
			},
		}
	}
}

// MockSyncFolder simulates copying messages for one folder, emitting a
// ProgressMsg for every batch and a final ProgressMsg{Done:true} at the end.
func MockSyncFolder(folder FolderInfo) tea.Cmd {
	return func() tea.Msg {
		const batchSize = 10
		copied := 0
		for copied < folder.Count {
			time.Sleep(80 * time.Millisecond)
			remaining := folder.Count - copied
			batch := batchSize
			if remaining < batch {
				batch = remaining
			}
			copied += batch
		}
		return ProgressMsg{
			Folder: folder.Name,
			Copied: folder.Count,
			Total:  folder.Count,
			Done:   true,
		}
	}
}

// MockSyncFolderTick streams individual ProgressMsg ticks so the progress bar
// animates smoothly. It returns a tea.Cmd that starts the goroutine and sends
// messages via the channel approach using tea.Tick-style sequential commands.
func MockSyncFolderTick(folder FolderInfo, alreadyCopied int) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(120 * time.Millisecond)
		next := alreadyCopied + 10
		if next > folder.Count {
			next = folder.Count
		}
		done := next >= folder.Count
		return ProgressMsg{
			Folder: folder.Name,
			Copied: next,
			Total:  folder.Count,
			Done:   done,
		}
	}
}

// MockConnectSource simulates connecting to the source IMAP server.
func MockConnectSource() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(400 * time.Millisecond)
		return LogMsg{Level: 3, Text: "Source IMAP connected (mock)"}
	}
}

// MockConnectDest simulates connecting to the destination IMAP server.
func MockConnectDest() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(500 * time.Millisecond)
		return LogMsg{Level: 3, Text: fmt.Sprintf("Destination IMAP connected (mock)")}
	}
}
