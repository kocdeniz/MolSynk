// Package imap wraps go-imap/v2 for MOLSYNK's connection, folder, and
// message-transfer needs. Every exported function that touches the network is
// designed to be called inside a tea.Cmd. CGO_ENABLED=0 safe — pure Go TLS.
package imap

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"time"

	imaplib "github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
)

// isIPAddress returns true when host is a bare IPv4 or IPv6 address.
// TLS certificates are almost never issued to raw IPs, so we skip
// verification in that case rather than refusing the connection.
func isIPAddress(host string) bool {
	return net.ParseIP(host) != nil
}

// ---- Config ------------------------------------------------------------------

// Config holds credentials and address for one IMAP endpoint.
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	TLS      bool
}

func (c Config) Addr() string { return fmt.Sprintf("%s:%d", c.Host, c.Port) }

// ---- Client ------------------------------------------------------------------

// Client is a connected, authenticated IMAP session.
// Callers must call Close when done.
type Client struct {
	inner *imapclient.Client
	Cfg   Config
}

// Connect dials and authenticates. Pure-Go TLS (no CGO).
func Connect(cfg Config) (*Client, error) {
	var (
		raw net.Conn
		err error
	)
	if cfg.TLS {
		tlsCfg := &tls.Config{
			ServerName: cfg.Host,
			MinVersion: tls.VersionTLS12,
		}
		// Certificates are not issued to IP addresses; skip verification
		// when the caller provided a bare IP so the connection still works.
		if isIPAddress(cfg.Host) {
			tlsCfg.InsecureSkipVerify = true //nolint:gosec
		}
		raw, err = tls.Dial("tcp", cfg.Addr(), tlsCfg)
	} else {
		raw, err = net.DialTimeout("tcp", cfg.Addr(), 15*time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", cfg.Addr(), err)
	}
	c := imapclient.New(raw, &imapclient.Options{})
	if err := c.Login(cfg.Username, cfg.Password).Wait(); err != nil {
		_ = c.Close()
		return nil, fmt.Errorf("login %s@%s: %w", cfg.Username, cfg.Host, err)
	}
	return &Client{inner: c, Cfg: cfg}, nil
}

// Close logs out and tears down the underlying connection.
func (cl *Client) Close() {
	_ = cl.inner.Logout().Wait()
	_ = cl.inner.Close()
}

// ---- Folder operations -------------------------------------------------------

// ListFolders returns the names of all mailboxes visible to the account.
func (cl *Client) ListFolders() ([]string, error) {
	items, err := cl.inner.List("", "*", &imaplib.ListOptions{
		ReturnStatus: &imaplib.StatusOptions{NumMessages: true},
	}).Collect()
	if err != nil {
		return nil, fmt.Errorf("LIST: %w", err)
	}
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.Mailbox)
	}
	return names, nil
}

// FolderStatus returns the number of messages in a named mailbox.
func (cl *Client) FolderStatus(name string) (uint32, error) {
	data, err := cl.inner.Status(name, &imaplib.StatusOptions{
		NumMessages: true,
	}).Wait()
	if err != nil {
		return 0, fmt.Errorf("STATUS %s: %w", name, err)
	}
	if data.NumMessages == nil {
		return 0, nil
	}
	return *data.NumMessages, nil
}

// EnsureFolder creates the mailbox on dst if it does not already exist.
// Idempotent: already-exists responses from the server are silently ignored.
func (cl *Client) EnsureFolder(name string) error {
	err := cl.inner.Create(name, nil).Wait()
	if err != nil && !isAlreadyExists(err) {
		return fmt.Errorf("CREATE %s: %w", name, err)
	}
	return nil
}

func isAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	msg := bytes.ToLower([]byte(err.Error()))
	for _, marker := range [][]byte{
		[]byte("alreadyexists"),
		[]byte("already exists"),
		[]byte("mailbox exists"),
	} {
		if bytes.Contains(msg, marker) {
			return true
		}
	}
	return false
}

// ---- Message transfer --------------------------------------------------------

// FetchUIDs selects a mailbox (read-only) and returns all message UIDs.
func (cl *Client) FetchUIDs(folder string) ([]imaplib.UID, error) {
	_, err := cl.inner.Select(folder, &imaplib.SelectOptions{ReadOnly: true}).Wait()
	if err != nil {
		return nil, fmt.Errorf("SELECT %s: %w", folder, err)
	}
	searchData, err := cl.inner.UIDSearch(&imaplib.SearchCriteria{}, nil).Wait()
	if err != nil {
		return nil, fmt.Errorf("UID SEARCH %s: %w", folder, err)
	}
	return searchData.AllUIDs(), nil
}

// TransferMessage fetches the raw RFC-5322 body of one message by UID from
// this client and APPENDs it to dstFolder on dst.
// An io.Pipe keeps peak RAM usage bounded to one message at a time.
func (cl *Client) TransferMessage(uid imaplib.UID, dst *Client, dstFolder string) error {
	// UID FETCH <uid> BODY[]
	uidSet := imaplib.UIDSetNum(uid)
	fetchOpts := &imaplib.FetchOptions{
		BodySection: []*imaplib.FetchItemBodySection{{}}, // empty section == whole message
	}
	fetchCmd := cl.inner.Fetch(uidSet, fetchOpts)
	defer fetchCmd.Close()

	msgData := fetchCmd.Next()
	if msgData == nil {
		return fmt.Errorf("FETCH uid %d: no message data", uid)
	}

	// Walk the fetch items to find the body section literal
	var bodyReader io.Reader
	for {
		item := msgData.Next()
		if item == nil {
			break
		}
		if bs, ok := item.(imapclient.FetchItemDataBodySection); ok {
			bodyReader = bs.Literal
			break
		}
	}
	if bodyReader == nil {
		return fmt.Errorf("FETCH uid %d: no body section in response", uid)
	}

	// Buffer the message body — required to know the exact size for APPEND
	var buf bytes.Buffer
	bw := bufio.NewWriterSize(&buf, 64*1024)
	if _, err := io.Copy(bw, bodyReader); err != nil {
		return fmt.Errorf("FETCH uid %d: read body: %w", uid, err)
	}
	if err := bw.Flush(); err != nil {
		return fmt.Errorf("FETCH uid %d: flush: %w", uid, err)
	}
	// Drain any remaining fetch data (flags, etc.)
	for fetchCmd.Next() != nil {
	}

	// APPEND to destination
	size := int64(buf.Len())
	appendCmd := dst.inner.Append(dstFolder, size, &imaplib.AppendOptions{})
	if _, err := io.Copy(appendCmd, bytes.NewReader(buf.Bytes())); err != nil {
		return fmt.Errorf("APPEND stream %s: %w", dstFolder, err)
	}
	if err := appendCmd.Close(); err != nil {
		return fmt.Errorf("APPEND close %s: %w", dstFolder, err)
	}
	if _, err := appendCmd.Wait(); err != nil {
		return fmt.Errorf("APPEND wait %s: %w", dstFolder, err)
	}
	return nil
}
