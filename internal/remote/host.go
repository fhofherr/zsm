package remote

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/fhofherr/zsm/internal/snapshot"
	gossh "golang.org/x/crypto/ssh"
)

// Host represents a remote host on which ZSM is installed.
type Host struct {
	User    string
	Addr    string
	AuthKey gossh.Signer
	HostKey gossh.PublicKey

	RemoteZSM string

	client *gossh.Client
	mu     sync.Mutex // protects client
}

// Dial creates a SSH connection to the remote host.
func (h *Host) Dial() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.client != nil {
		// already connected
		return nil
	}
	config := &gossh.ClientConfig{
		User: h.User,
		Auth: []gossh.AuthMethod{
			gossh.PublicKeys(h.AuthKey),
		},
		HostKeyCallback: gossh.FixedHostKey(h.HostKey),
	}
	client, err := gossh.Dial("tcp", h.Addr, config)
	if err != nil {
		return fmt.Errorf("dial ssh: %w", err)
	}
	h.client = client
	return nil
}

func (h *Host) newSession() (*gossh.Session, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.client == nil {
		return nil, fmt.Errorf("not connected")
	}
	sess, err := h.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return sess, nil
}

// Close closes the connection to the remote host.
func (h *Host) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.client == nil {
		return nil
	}
	client := h.client
	h.client = nil

	return client.Close()
}

// ListSnapshots lists all snapshots available on the remote host.
func (h *Host) ListSnapshots() ([]snapshot.Name, error) {
	var (
		stdout   bytes.Buffer
		parseBuf bytes.Buffer
	)

	zsmListCmd := fmt.Sprintf("%s list -o jsonl", h.RemoteZSM)
	if err := h.runRemoteZSM(zsmListCmd, &stdout, nil); err != nil {
		return nil, err
	}
	// Bail out if the remote side has no snapshots.
	if stdout.Len() == 0 {
		return nil, nil
	}

	// Pre-allocate for 100 snapshots. This covers daily snapshots for about 3
	// months plus a few hourly snapshots. This should be enough for most cases.
	names := make([]snapshot.Name, 0, 100)
	for _, b := range stdout.Bytes() {
		parseBuf.WriteByte(b)
		if b != '\n' {
			continue
		}
		name, err := snapshot.ParseNameJSON(parseBuf.Bytes())
		if err != nil {
			return nil, fmt.Errorf("remote zsm: invalid snapshot name: %s", parseBuf.String())
		}
		parseBuf.Reset()
		names = append(names, name)
	}
	return names, nil
}

func (h *Host) runRemoteZSM(cmd string, stdout io.Writer, stdin io.Reader) error {
	var stderr bytes.Buffer

	sess, err := h.newSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	sess.Stdin = stdin
	sess.Stdout = stdout
	sess.Stderr = &stderr

	if err := sess.Run(cmd); err != nil {
		exitErr := &gossh.ExitError{}
		if !errors.As(err, &exitErr) {
			return err
		}
		return &Error{
			SubCommand: "list",
			ExitCode:   exitErr.ExitStatus(),
			Stderr:     stderr.String(),
		}
	}
	return nil
}
