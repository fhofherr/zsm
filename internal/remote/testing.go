package remote

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fhofherr/netutil"
	"github.com/gliderlabs/ssh"
	"github.com/stretchr/testify/assert"
	gossh "golang.org/x/crypto/ssh"
)

// RunTests runs all passed tests.
func RunTests(t *testing.T, tests []TestCase) {
	for _, tt := range tests {
		tt := tt

		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			tt.run(t)
		})
	}
}

// TestCase defines a test case for remote zsm invocation.
type TestCase struct {
	Name string

	// Call calls invokes a method on the remote Host. It should return any error
	// returned by the called host method. Additionally it may use the passed T
	// to perform further assertions.
	Call func(*testing.T, *Host) error

	Stdin  func(*testing.T) []byte // returns stdin expected to be sent to remote host
	Stdout func(*testing.T) []byte // returns stdout expected to be sent by remote host
	Stderr func(*testing.T) []byte // returns stderr expected to be sent by remote host

	ZSMExitCode int      // expected exit code of remote zsm
	ZSMCommand  []string // expected arguments for remote zsm

	// Abort waiting on channels after this time elapsed.
	// Default 10ms per channel
	ChannelTimeout time.Duration

	errC     chan error
	hostKeyC chan gossh.Signer
	addrC    chan string
	commandC chan []string
	stdinC   chan []byte

	server *SSHTestServer

	clientSigner gossh.Signer
	host         *Host
}

func (tt *TestCase) run(t *testing.T) {
	t.Helper()
	tt.init(t)

	if err := tt.Call(t, tt.host); err != nil {
		remoteErr := &Error{}
		if !errors.As(err, &remoteErr) {
			t.Fatalf("unexpected remote zsm error: %v", err)
		}
		assert.Equal(t, tt.ZSMExitCode, remoteErr.ExitCode)

		expectedStderr := []byte("")
		if tt.Stderr != nil {
			expectedStderr = tt.Stderr(t)
		}
		assert.Equal(t, expectedStderr, []byte(remoteErr.Stderr))
	}
	if len(tt.ZSMCommand) > 0 {
		select {
		case command := <-tt.commandC:
			assert.Equal(t, tt.ZSMCommand, command, "zsm was not called with the expected args")
		case <-time.After(tt.ChannelTimeout):
			t.Errorf("timed out waiting for zsm command after %v", tt.ChannelTimeout)
		}
	}
	if tt.Stdin != nil {
		expected := tt.Stdin(t)
		select {
		case actual := <-tt.stdinC:
			assert.Equal(t, expected, actual, "zsm did not receive the expected stdin")
		case <-time.After(tt.ChannelTimeout):
			t.Errorf("timed out waiting for zsm stdin after %v", tt.ChannelTimeout)
		}
	}

loop:
	for {
		select {
		case err := <-tt.errC:
			t.Error(err)
		default:
			break loop
		}
	}
}

func (tt *TestCase) init(t *testing.T) {
	tt.ChannelTimeout = 10 * time.Millisecond
	tt.errC = make(chan error, 1)
	tt.hostKeyC = make(chan gossh.Signer, 1)
	tt.addrC = make(chan string, 1)
	tt.commandC = make(chan []string, 1)
	tt.stdinC = make(chan []byte, 1)

	tt.makeClientKeys(t)
	tt.newSSHTestServer(t)
	go startSSHTestServer(tt.server, tt.addrC, tt.errC)

	tt.makeHost(t)
}

func (tt *TestCase) makeClientKeys(t *testing.T) {
	signer, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	clientSigner, err := gossh.NewSignerFromSigner(signer)
	if err != nil {
		t.Fatal(err)
	}
	tt.clientSigner = clientSigner
}

func (tt *TestCase) newSSHTestServer(t *testing.T) {
	var stdout, stderr []byte

	if tt.Stdout != nil {
		stdout = tt.Stdout(t)
	}
	if tt.Stderr != nil {
		stderr = tt.Stderr(t)
	}
	tt.server = &SSHTestServer{
		AuthorizedKeys: []ssh.PublicKey{tt.clientSigner.PublicKey()},
		HostKey:        tt.hostKeyC,
		Command:        tt.commandC,
		Stdin:          tt.stdinC,
		Stdout:         stdout,
		Stderr:         stderr,
		ExitCode:       tt.ZSMExitCode,
		Errors:         tt.errC,
	}
	t.Cleanup(func() {
		tt.server.Close()
	})
}

func (tt *TestCase) makeHost(t *testing.T) {
	var (
		hostKey   gossh.Signer
		remoteZSM string
	)
	addr := netutil.GetAddr(t, tt.addrC, tt.ChannelTimeout)

	select {
	case hostKey = <-tt.hostKeyC:
	case <-time.After(tt.ChannelTimeout):
		t.Fatalf("waiting for host key: timeout after: %v", tt.ChannelTimeout)
	}
	if len(tt.ZSMCommand) > 0 {
		remoteZSM = tt.ZSMCommand[0]
	}
	tt.host = &Host{
		Addr:      addr,
		AuthKey:   tt.clientSigner,
		HostKey:   hostKey.PublicKey(),
		RemoteZSM: remoteZSM,
	}
}

func startSSHTestServer(s *SSHTestServer, addrC chan<- string, errC chan<- error) {
	if err := netutil.ListenAndServe(s, netutil.NotifyAddr(addrC)); err != nil {
		errC <- err
	}
}

// SSHTestServer represents a simple SSH server for testing purposes.
type SSHTestServer struct {
	// AuthorizedKeys contains public SSH keys of clients allowed to connect to
	// the server. Changes made after calling Serve are ignored.
	AuthorizedKeys []ssh.PublicKey

	// Receives the host key of this server once it has been initialized.
	HostKey chan<- gossh.Signer

	// The server writes any command a client passed when initiating a session
	// into this channel. Setting Command after calling Serve has no effect.
	Command chan<- []string

	// The server writes anything read from stdin into this channel.
	// Setting Stdin after calling Serve has no effect.
	Stdin chan<- []byte

	// The server writes this to stdout.
	// Setting Stdout after calling Serve has no effect.
	Stdout []byte

	// The server writes this to stderr.
	// Setting Stderr after calling Serve has no effect.
	Stderr []byte

	// Any errors that occur while handling a request are written to this channel.
	// Setting Errors after calling Serve has no effect.
	Errors chan<- error

	// The server terminates all sessions with this exit code. Changing ExitCode
	// after calling Serve has no effect.
	ExitCode int

	server *ssh.Server

	initialized int32
	once        sync.Once
	initErr     error
}

// Serve accepts incoming connections to l.
func (s *SSHTestServer) Serve(l net.Listener) error {
	if err := s.init(); err != nil {
		return err
	}
	return s.server.Serve(l)
}

func (s *SSHTestServer) init() error {
	if atomic.LoadInt32(&s.initialized) != 0 && s.initErr == nil {
		return errors.New("ssh server: already initialized")
	}
	s.once.Do(func() {
		signer, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			s.initErr = fmt.Errorf("generate host key: %w", err)
			return
		}
		hostKey, err := gossh.NewSignerFromSigner(signer)
		if err != nil {
			s.initErr = fmt.Errorf("convert host key to gossh.Signer: %w", err)
			return
		}
		// Send the host key to the caller. Do this in a separate go routine
		// to ensure we don't block here, even if the caller did not pass an
		// unbuffered channel.
		go s.sendHostKey(hostKey)

		authorizedKeys := append([]ssh.PublicKey(nil), s.AuthorizedKeys...)
		s.server = &ssh.Server{
			HostSigners: []ssh.Signer{hostKey},
			PublicKeyHandler: func(_ ssh.Context, pubKey ssh.PublicKey) bool {
				for _, allowed := range authorizedKeys {
					if ssh.KeysEqual(allowed, pubKey) {
						return true
					}
				}
				return false
			},
			Handler: sshSessionHandler(s.ExitCode, s.Command, s.Stdin, s.Stdout, s.Stderr, s.Errors),
		}
		atomic.StoreInt32(&s.initialized, 1)
	})
	return s.initErr
}

func (s *SSHTestServer) sendHostKey(hostKey gossh.Signer) {
	if s.HostKey == nil {
		return
	}
	s.HostKey <- hostKey
}

// Close closes the server.
func (s *SSHTestServer) Close() error {
	if atomic.LoadInt32(&s.initialized) == 0 {
		return nil
	}
	return s.server.Close()
}

func sshSessionHandler(
	exitCode int,
	commandC chan<- []string,
	stdinC chan<- []byte,
	stdout, stderr []byte,
	errC chan<- error,
) ssh.Handler {
	sendErr := func(err error) {
		if errC != nil {
			errC <- err
		}
	}
	handleStdin := func(stdin io.Reader) {
		bs, err := ioutil.ReadAll(stdin)
		if err != nil {
			sendErr(err)
			return
		}
		if stdinC != nil {
			stdinC <- bs
		}
	}
	return func(sess ssh.Session) {
		if commandC != nil {
			commandC <- append([]string(nil), sess.Command()...)
		}
		go handleStdin(sess)
		if _, err := sess.Write(stdout); err != nil {
			sendErr(err)
			return
		}
		if _, err := sess.Stderr().Write(stderr); err != nil {
			sendErr(err)
			return
		}
		sess.Exit(exitCode) // nolint: errcheck
	}
}
