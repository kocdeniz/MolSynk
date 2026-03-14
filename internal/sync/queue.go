package sync

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"imapsync/internal/imap"
)

// AccountPair describes one source→destination account migration.
type AccountPair struct {
	SrcCfg imap.Config
	DstCfg imap.Config
}

// ParseQueueFile reads a .csv or .txt file and returns a slice of AccountPair.
//
// Each non-blank, non-comment line must have exactly 4 comma-separated fields:
//
//	src_user, src_pass, dst_user, dst_pass
//
// The global source and destination hosts/ports/TLS settings are injected from
// the caller (entered on the Bulk setup screen).
func ParseQueueFile(
	path string,
	srcHost string, srcPort int, srcTLS bool,
	dstHost string, dstPort int, dstTLS bool,
) ([]AccountPair, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	var pairs []AccountPair
	lineNo := 0
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		lineNo++
		raw := strings.TrimSpace(scanner.Text())

		// Skip blank lines and comments
		if raw == "" || strings.HasPrefix(raw, "#") {
			continue
		}

		fields := splitCSVLine(raw)
		if len(fields) != 4 {
			return nil, fmt.Errorf("line %d: expected 4 fields (src_user, src_pass, dst_user, dst_pass), got %d", lineNo, len(fields))
		}

		srcUser := strings.TrimSpace(fields[0])
		srcPass := strings.TrimSpace(fields[1])
		dstUser := strings.TrimSpace(fields[2])
		dstPass := strings.TrimSpace(fields[3])

		if srcUser == "" || srcPass == "" || dstUser == "" || dstPass == "" {
			return nil, fmt.Errorf("line %d: one or more fields are empty", lineNo)
		}

		pairs = append(pairs, AccountPair{
			SrcCfg: imap.Config{
				Host:     srcHost,
				Port:     srcPort,
				Username: srcUser,
				Password: srcPass,
				TLS:      srcTLS,
			},
			DstCfg: imap.Config{
				Host:     dstHost,
				Port:     dstPort,
				Username: dstUser,
				Password: dstPass,
				TLS:      dstTLS,
			},
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if len(pairs) == 0 {
		return nil, fmt.Errorf("no valid account lines found in %s", path)
	}
	return pairs, nil
}

// splitCSVLine splits a single CSV line on commas, respecting basic quoted fields.
func splitCSVLine(line string) []string {
	var fields []string
	var cur strings.Builder
	inQuote := false

	for i := 0; i < len(line); i++ {
		c := line[i]
		switch {
		case c == '"':
			inQuote = !inQuote
		case c == ',' && !inQuote:
			fields = append(fields, cur.String())
			cur.Reset()
		default:
			cur.WriteByte(c)
		}
	}
	fields = append(fields, cur.String())
	return fields
}
