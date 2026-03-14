# MOLSYNK

![MOLSYNK Header](molsynk_header.png)

MOLSYNK is a terminal-first IMAP-to-IMAP migration tool written in Go.
It is designed for practical mailbox migration with a clear TUI, low-memory
message transfer, and compatibility with older Linux systems.

## Current capabilities

- Interactive TUI built with Bubble Tea and Lip Gloss
- Intro/branding screen and mode selection
- Manual mode (single account pair)
- Bulk mode (multiple account pairs from file)
- Real IMAP authentication and folder discovery
- Folder creation on destination when needed
- UID-based message copy from source to destination
- Real-time activity log and progress updates
- Per-account fault isolation in bulk mode (continue on errors)
- `CGO_ENABLED=0` compatible build

## Requirements

- Go 1.24+
- IMAP access to both source and destination servers
- Network access to IMAP ports (usually `993` for TLS)

## Quick start

Run from source:

```bash
go run .
```

Build a static binary:

```bash
CGO_ENABLED=0 go build -o molsynk .
```

Run the binary:

```bash
./molsynk
```

## Usage flow

1. Intro screen: press any key.
2. Select migration mode:
   - `1` Manual Entry
   - `2` Bulk Migration via File

### Manual mode

Fill these fields:

- Source Host:Port
- Source Username
- Source Password
- Destination Host:Port
- Destination Username
- Destination Password

Press `Enter` on the last field to connect.

If connection succeeds, dashboard opens and you can press `s` to start sync.
If connection fails, the app returns to the form and shows the exact error.

### Bulk mode

Fill these fields:

- Global Source Host:Port
- Global Destination Host:Port
- Accounts File

Accepted file extensions: `.csv`, `.txt`

File format (one account pair per line):

```text
src_user,src_pass,dst_user,dst_pass
```

Notes:

- Lines starting with `#` are treated as comments.
- Empty lines are ignored.
- On validation success, migration starts immediately.

## How migration works

For each account pair:

1. Connect and authenticate to source and destination IMAP servers.
2. List source folders.
3. Ensure each folder exists on destination (`CREATE` if needed).
4. `UID SEARCH` source folder.
5. For each UID:
   - `FETCH BODY[]` from source
   - `APPEND` to destination folder
6. Emit live status updates to the TUI via channel messages.

If a folder or account fails, the error is logged and migration continues with
the next item.

## Folder naming behavior

MOLSYNK preserves server folder names exactly as returned by IMAP.

Example: if source uses `INBOX.Sent`, `INBOX.Drafts`, `INBOX.Archive`, those
exact names are created/copied on destination. This is normal IMAP behavior.

## Security notes

- Credentials are entered directly in the terminal UI.
- For TLS connections made with a raw IP address, certificate verification is
  relaxed to avoid hostname mismatch failures.
- For hostname-based connections, normal TLS hostname verification applies.

## Project structure

```text
main.go                  # App entrypoint
internal/ui/             # Bubble Tea model, update loop, rendering
internal/sync/           # Queue parser and migration engine
internal/imap/           # IMAP client wrapper and transfer operations
```

## Known limitations

- No dry-run mode yet
- No final summary report export yet
- No retry/backoff policy yet (errors are logged and processing continues)

## Status

This project is under active development. Interfaces and behavior may evolve as
new migration and reporting features are added.
