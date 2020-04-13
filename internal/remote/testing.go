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

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

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
