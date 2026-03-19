package sync

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	imapconn "github.com/kocdeniz/mailmole/internal/imap"
)

// PreviewData represents migration preview information
type PreviewData struct {
	SourceAccount string
	DestAccount   string
	SourceHost    string
	DestHost      string
	Folders       []FolderPreviewInfo
}

// FolderPreviewInfo holds preview data for a single folder
type FolderPreviewInfo struct {
	Name         string
	MessageCount int
	SizeEstimate int64
}

// PreviewMsg is sent when preview data is ready
type PreviewMsg struct {
	Data  PreviewData
	Error error
}

// FetchPreview gathers preview information from source IMAP server
func FetchPreview(srcCfg, dstCfg imapconn.Config) tea.Cmd {
	return func() tea.Msg {
		// Connect to source
		src, err := imapconn.Connect(srcCfg)
		if err != nil {
			return PreviewMsg{Error: fmt.Errorf("source connection: %w", err)}
		}
		defer src.Close()

		// List folders
		folderNames, err := src.ListFolders()
		if err != nil {
			return PreviewMsg{Error: fmt.Errorf("listing folders: %w", err)}
		}

		// Get message count for each folder
		var folders []FolderPreviewInfo
		var totalMessages int
		var totalSize int64

		for _, name := range folderNames {
			count, err := src.FolderStatus(name)
			if err != nil {
				// Skip folders we can't access
				continue
			}

			// Estimate size: average 50KB per message (conservative)
			estimatedSize := int64(count) * 50 * 1024

			folders = append(folders, FolderPreviewInfo{
				Name:         name,
				MessageCount: int(count),
				SizeEstimate: estimatedSize,
			})

			totalMessages += int(count)
			totalSize += estimatedSize
		}

		return PreviewMsg{
			Data: PreviewData{
				SourceAccount: srcCfg.Username,
				DestAccount:   dstCfg.Username,
				SourceHost:    fmt.Sprintf("%s:%d", srcCfg.Host, srcCfg.Port),
				DestHost:      fmt.Sprintf("%s:%d", dstCfg.Host, dstCfg.Port),
				Folders:       folders,
			},
		}
	}
}

// CalculateEstimatedDuration estimates migration time based on message count
// Assumes 2 messages per second average with retry logic
func CalculateEstimatedDuration(messageCount int) time.Duration {
	seconds := float64(messageCount) / 2.0
	if seconds < 60 {
		return time.Duration(seconds) * time.Second
	}
	return time.Duration(seconds) * time.Second
}

// FormatSize formats byte size to human readable string
func FormatSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// FormatDuration formats duration to human readable string
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		if seconds > 0 {
			return fmt.Sprintf("%d min %d sec", minutes, seconds)
		}
		return fmt.Sprintf("%d minutes", minutes)
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if minutes > 0 {
		return fmt.Sprintf("%d hours %d min", hours, minutes)
	}
	return fmt.Sprintf("%d hours", hours)
}
