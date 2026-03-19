# MailMole

![MailMole Header](mailmole.jpeg)

MailMole is a terminal-first IMAP-to-IMAP migration tool written in Go.
It is designed for practical mailbox migration with a clear TUI, low-memory
message transfer, and compatibility with older Linux systems.

## Current capabilities

- **Interactive TUI** built with Bubble Tea and Lip Gloss
- **Preview Mode** - Review folders and message counts before migration
- **Web Dashboard** - Monitor migrations from any browser (optional)
- Intro/branding screen and mode selection
- Manual mode (single account pair)
- Bulk mode (multiple account pairs from file)
- Real IMAP authentication and folder discovery
- Folder creation on destination when needed
- UID-based message copy from source to destination
- Smart Retry with exponential backoff (`2s`, `5s`, `10s` + jitter)
- O(1) duplicate detection via in-memory `Message-ID` cache
- Batch metadata fetch for faster pre-filtering (`50` per batch)
- Parallel folder workers (up to `3` concurrently)
- Checkpoint persistence (`migration_state.json`) for bulk resume/skip
- Real-time activity log and progress updates
- Per-account fault isolation in bulk mode (continue on errors)
- `CGO_ENABLED=0` compatible build

## Why MailMole is different

### Smart Retry (resilience under real-world server pressure)

MailMole does not fail fast on transient IMAP/network errors. It automatically:

- Detects retryable errors (`timeout`, `server busy`, throttling patterns)
- Retries with exponential backoff (`2s`, `5s`, `10s`) + jitter
- Attempts socket/session recovery to avoid zombie connections
- Continues processing other folders/accounts when one unit fails

This is especially important on SmarterMail and enterprise servers with
throttling/rate-limit behavior.

### O(1) Caching (high-speed duplicate detection)

Instead of running a slow duplicate search per message, MailMole:

1. Loads destination folder `Message-ID` values once into memory
2. Keeps them in a `map[string]bool`
3. Uses O(1) lookup per source message

This removes repeated server-side search overhead and is one of the key reasons
for high throughput in large migrations.

## Requirements

- Go 1.24+
- IMAP access to both source and destination servers
- Network access to IMAP ports (usually `993` for TLS)

## Quick start

Download for Linux
```bash
wget https://github.com/kocdeniz/MailMole/releases/download/v1.0.0/mailmole_linux
```

Run from source:

```bash
go run .
```

Build a static binary:

```bash
CGO_ENABLED=0 go build -o mailmole .
```

Run the binary:

```bash
./mailmole
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

Press `Enter` on the last field to continue.

#### Preview Mode

After connecting, you'll see the **Preview Screen** showing:
- All folders from the source account
- Message count per folder
- Estimated total size
- Estimated migration duration

**Controls:**
- `↑/↓` - Navigate folders
- `Space` - Toggle folder selection
- `a` - Select all folders
- `n` - Select none
- `Enter` - Start migration with selected folders
- `Esc` - Go back

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
2. List source folders and process folders with a worker pool.
3. Ensure each folder exists on destination (`CREATE` if needed).
4. Preload destination `Message-ID` cache (single batched pass).
5. Fetch source metadata in batches (`UID`, `Message-ID`, size).
6. O(1) map lookup to skip duplicates before body transfer.
7. Transfer only required messages (`FETCH BODY[]` -> `APPEND`).
8. Emit live status/speed updates to the TUI via channel messages.

If a folder or account fails, the error is logged and migration continues with
the next item.

## Folder naming behavior

MailMole preserves server folder names exactly as returned by IMAP.

Example: if source uses `INBOX.Sent`, `INBOX.Drafts`, `INBOX.Archive`, those
exact names are created/copied on destination. This is normal IMAP behavior.

## Web Dashboard (Optional)

MailMole includes an optional web dashboard for monitoring migrations from any browser.

### Usage

Run with web dashboard enabled:
```bash
# TUI + Web Dashboard
./mailmole -web :8080

# Web Dashboard only (no TUI)
./mailmole -web :8080 -web-only
```

Then open `http://localhost:8080` in your browser.

### Features

- **Real-time monitoring** via Server-Sent Events (SSE)
- Live progress bar and statistics
- Per-folder and per-account status
- Activity log viewer
- Responsive design (works on mobile)
- Dark theme matching the TUI

The web dashboard updates automatically as the migration progresses, making it ideal for:
- Remote monitoring from another device
- Sharing progress with team members
- Running migrations on headless servers

## Security notes

- Credentials are entered directly in the terminal UI.
- For TLS connections made with a raw IP address, certificate verification is
  relaxed to avoid hostname mismatch failures.
- For hostname-based connections, normal TLS hostname verification applies.
- **Web Dashboard**: No authentication by default - only run on trusted networks
  or use firewall rules to restrict access.

## Project structure

```text
main.go                  # App entrypoint
internal/ui/             # Bubble Tea model, update loop, rendering
internal/sync/           # Queue parser and migration engine
internal/imap/           # IMAP client wrapper and transfer operations
```

## Known limitations

- No JSON/CSV export of migration reports yet (web dashboard available)
- Web dashboard has no built-in authentication

## Status

This project is under active development. Interfaces and behavior may evolve as
new migration and reporting features are added.
