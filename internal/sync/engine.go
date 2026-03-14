// Package sync contains MOLSYNK's migration engine.
// RunMigration runs the full account queue in a goroutine and sends
// StatusUpdateMsg values back to the Bubble Tea event loop via a channel
// wrapped in tea.Cmd. This keeps the UI fully responsive at all times.
package sync

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"imapsync/internal/imap"
)

// ---- Status messages ---------------------------------------------------------

// StatusKind classifies each update sent to the UI.
type StatusKind int

const (
	StatusAccountStart  StatusKind = iota // starting a new account
	StatusAccountDone                     // account finished successfully
	StatusAccountError                    // account failed, moving to next
	StatusFolderStart                     // entering a new folder
	StatusFolderDone                      // folder finished
	StatusMessageCopied                   // one message transferred
	StatusMigrationDone                   // entire queue finished
)

// StatusUpdateMsg is posted to the Bubble Tea event loop from the goroutine.
type StatusUpdateMsg struct {
	Kind    StatusKind
	Account string // src username — identifies the account
	Folder  string
	Copied  int // messages copied so far in the current folder
	Total   int // total messages in the current folder
	Err     error
}

// ---- tea.Cmd wiring ----------------------------------------------------------

// RunMigration launches the migration engine in a goroutine and returns both:
//   - A tea.Cmd that reads the first StatusUpdateMsg from the channel.
//   - The channel itself, so the caller can store it on the model and call
//     WaitForNext to schedule subsequent reads.
func RunMigration(pairs []AccountPair) (tea.Cmd, <-chan StatusUpdateMsg) {
	ch := make(chan StatusUpdateMsg, 64)
	go runEngine(pairs, ch)
	return drainCmd(ch), ch
}

// drainCmd returns a tea.Cmd that reads one message from ch, then re-schedules
// itself if the channel is still open.
func drainCmd(ch <-chan StatusUpdateMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			// Channel closed — signal completion
			return StatusUpdateMsg{Kind: StatusMigrationDone}
		}
		return msg
	}
}

// WaitForNext is used by the Update function after receiving a StatusUpdateMsg
// to schedule reading the next one from the same channel.
func WaitForNext(ch <-chan StatusUpdateMsg) tea.Cmd {
	return drainCmd(ch)
}

// ---- Engine ------------------------------------------------------------------

// runEngine iterates over every AccountPair and migrates each one.
// Errors on individual accounts are non-fatal: they are sent as
// StatusAccountError and the engine moves to the next account.
func runEngine(pairs []AccountPair, ch chan<- StatusUpdateMsg) {
	defer close(ch)

	for _, pair := range pairs {
		label := pair.SrcCfg.Username

		ch <- StatusUpdateMsg{Kind: StatusAccountStart, Account: label}

		if err := migrateAccount(pair, label, ch); err != nil {
			ch <- StatusUpdateMsg{
				Kind:    StatusAccountError,
				Account: label,
				Err:     err,
			}
			// Non-fatal: continue with next account
			continue
		}

		ch <- StatusUpdateMsg{Kind: StatusAccountDone, Account: label}
	}
}

// migrateAccount performs the full folder-by-folder migration for one pair.
func migrateAccount(pair AccountPair, label string, ch chan<- StatusUpdateMsg) error {
	// Connect source
	src, err := imap.Connect(pair.SrcCfg)
	if err != nil {
		return fmt.Errorf("connect source: %w", err)
	}
	defer src.Close()

	// Connect destination
	dst, err := imap.Connect(pair.DstCfg)
	if err != nil {
		return fmt.Errorf("connect destination: %w", err)
	}
	defer dst.Close()

	// Discover source folders
	folders, err := src.ListFolders()
	if err != nil {
		return fmt.Errorf("list folders: %w", err)
	}

	for _, folder := range folders {
		if err := migrateFolder(src, dst, folder, label, ch); err != nil {
			// Per-folder error is non-fatal: log and skip to next folder
			ch <- StatusUpdateMsg{
				Kind:    StatusAccountError,
				Account: label,
				Folder:  folder,
				Err:     fmt.Errorf("folder %s: %w", folder, err),
			}
		}
	}
	return nil
}

// migrateFolder copies all messages from one source folder to the equivalent
// destination folder. It transfers messages one at a time to keep memory usage
// bounded — peak RAM = one message body at any instant.
func migrateFolder(src, dst *imap.Client, folder, account string, ch chan<- StatusUpdateMsg) error {
	// Ensure the folder exists on the destination
	if err := dst.EnsureFolder(folder); err != nil {
		return fmt.Errorf("ensure folder: %w", err)
	}

	// Fetch all UIDs from source
	uids, err := src.FetchUIDs(folder)
	if err != nil {
		return fmt.Errorf("fetch UIDs: %w", err)
	}
	total := len(uids)
	if total == 0 {
		return nil // nothing to do
	}

	ch <- StatusUpdateMsg{
		Kind:    StatusFolderStart,
		Account: account,
		Folder:  folder,
		Copied:  0,
		Total:   total,
	}

	copied := 0
	for _, uid := range uids {
		if err := src.TransferMessage(uid, dst, folder); err != nil {
			// Non-fatal per message: emit a warning and continue
			ch <- StatusUpdateMsg{
				Kind:    StatusAccountError,
				Account: account,
				Folder:  folder,
				Err:     fmt.Errorf("uid %d: %w", uid, err),
			}
			continue
		}
		copied++
		ch <- StatusUpdateMsg{
			Kind:    StatusMessageCopied,
			Account: account,
			Folder:  folder,
			Copied:  copied,
			Total:   total,
		}
	}

	ch <- StatusUpdateMsg{
		Kind:    StatusFolderDone,
		Account: account,
		Folder:  folder,
		Copied:  copied,
		Total:   total,
	}
	return nil
}
