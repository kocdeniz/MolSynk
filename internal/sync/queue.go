package sync

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/kocdeniz/mailmole/internal/imap"
)

// AccountPair describes one source→destination account migration.
type AccountPair struct {
	SrcCfg imap.Config
	DstCfg imap.Config
}

// ParseQueueFile reads a .csv or .txt file and returns a slice of AccountPair.
//
// Supports two formats:
// 1. Simple (4 fields): src_user, src_pass, dst_user, dst_pass
//   - Uses global source and destination hosts/ports from parameters
//
// 2. Advanced (6 fields): src_host, src_user, src_pass, dst_host, dst_user, dst_pass
//   - Each account can have different source and destination servers
//
// Examples:
//
//	Simple:  eren@old.com,pass123,eren@new.com,pass456
//	Advanced: mail.old.com,eren@old.com,pass123,mail.new.com,eren@new.com,pass456
func ParseQueueFile(
	path string,
	defaultSrcHost string, defaultSrcPort int, defaultSrcTLS bool,
	defaultDstHost string, defaultDstPort int, defaultDstTLS bool,
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

		// Support both 4-field (simple) and 6-field (advanced) formats
		if len(fields) != 4 && len(fields) != 6 {
			return nil, fmt.Errorf("line %d: expected 4 fields (simple) or 6 fields (advanced), got %d", lineNo, len(fields))
		}

		var srcHost, srcUser, srcPass, dstHost, dstUser, dstPass string
		var srcPort, dstPort int
		var srcTLS, dstTLS bool

		if len(fields) == 4 {
			// Simple format: src_user, src_pass, dst_user, dst_pass
			srcHost = defaultSrcHost
			srcPort = defaultSrcPort
			srcTLS = defaultSrcTLS
			srcUser = strings.TrimSpace(fields[0])
			srcPass = strings.TrimSpace(fields[1])
			dstHost = defaultDstHost
			dstPort = defaultDstPort
			dstTLS = defaultDstTLS
			dstUser = strings.TrimSpace(fields[2])
			dstPass = strings.TrimSpace(fields[3])
		} else {
			// Advanced format: src_host, src_user, src_pass, dst_host, dst_user, dst_pass
			srcHost, srcPort, srcTLS = parseHostConfig(strings.TrimSpace(fields[0]), defaultSrcPort, defaultSrcTLS)
			srcUser = strings.TrimSpace(fields[1])
			srcPass = strings.TrimSpace(fields[2])
			dstHost, dstPort, dstTLS = parseHostConfig(strings.TrimSpace(fields[3]), defaultDstPort, defaultDstTLS)
			dstUser = strings.TrimSpace(fields[4])
			dstPass = strings.TrimSpace(fields[5])
		}

		if srcHost == "" || srcUser == "" || srcPass == "" || dstHost == "" || dstUser == "" || dstPass == "" {
			return nil, fmt.Errorf("line %d: one or more required fields are empty", lineNo)
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

// parseHostConfig parses host string which can be:
// - hostname (e.g., "mail.example.com") - uses default port
// - hostname:port (e.g., "mail.example.com:993")
func parseHostConfig(hostStr string, defaultPort int, defaultTLS bool) (string, int, bool) {
	hostStr = strings.TrimSpace(hostStr)
	if hostStr == "" {
		return "", defaultPort, defaultTLS
	}

	// Check if port is specified
	if idx := strings.LastIndex(hostStr, ":"); idx > 0 {
		host := hostStr[:idx]
		portStr := hostStr[idx+1:]
		if port, err := parsePort(portStr); err == nil {
			return host, port, port == 993
		}
	}

	// No port specified, use defaults
	return hostStr, defaultPort, defaultTLS
}

func parsePort(portStr string) (int, error) {
	var port int
	_, err := fmt.Sscanf(portStr, "%d", &port)
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("invalid port")
	}
	return port, nil
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
