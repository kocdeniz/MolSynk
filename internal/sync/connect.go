package sync

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"imapsync/internal/imap"
)

// ---- Messages ----------------------------------------------------------------

// ConnectedMsg is sent when both IMAP servers have been successfully
// authenticated. It carries live clients so the sync phase can reuse them.
type ConnectedMsg struct {
	Src     *imap.Client
	Dst     *imap.Client
	Folders []FolderInfo // discovered from Src
}

// ConnErrMsg is sent when a connection or authentication attempt fails.
type ConnErrMsg struct {
	Err error
}

func (e ConnErrMsg) Error() string { return e.Err.Error() }

// ---- Commands ----------------------------------------------------------------

// Connect dials both servers, authenticates, and lists source folders.
// It is safe to call from a tea.Cmd — all blocking I/O happens inside the
// returned func and never touches the Bubble Tea event loop directly.
func Connect(srcCfg, dstCfg imap.Config) tea.Cmd {
	return func() tea.Msg {
		// Connect source
		src, err := imap.Connect(srcCfg)
		if err != nil {
			return ConnErrMsg{Err: fmt.Errorf("source: %w", err)}
		}

		// Connect destination
		dst, err := imap.Connect(dstCfg)
		if err != nil {
			src.Close()
			return ConnErrMsg{Err: fmt.Errorf("destination: %w", err)}
		}

		// Discover folders on source
		names, err := src.ListFolders()
		if err != nil {
			src.Close()
			dst.Close()
			return ConnErrMsg{Err: fmt.Errorf("listing folders: %w", err)}
		}

		folders := make([]FolderInfo, 0, len(names))
		for _, name := range names {
			count, err := src.FolderStatus(name)
			if err != nil {
				// Non-fatal: include with 0 count
				folders = append(folders, FolderInfo{Name: name, Count: 0})
				continue
			}
			folders = append(folders, FolderInfo{Name: name, Count: int(count)})
		}

		return ConnectedMsg{
			Src:     src,
			Dst:     dst,
			Folders: folders,
		}
	}
}
