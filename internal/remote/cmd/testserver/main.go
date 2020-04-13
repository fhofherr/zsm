// testserver provides a command to start the remote.SSHTestServer. Its main
// purpose is to make sure, that the server works as expected before using it
// in a test (and afterwards should we believe there is a flaw).
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/fhofherr/netutil"
	"github.com/fhofherr/zsm/internal/remote"
	"github.com/gliderlabs/ssh"
	"github.com/spf13/pflag"
)

var (
	authorizedKeyFile string
	port              int
)

func init() {
	pflag.StringVarP(&authorizedKeyFile, "auth-key-file", "a", "~/.ssh/id_rsa.pub",
		"File containing the only authorized public key.")
	pflag.IntVarP(&port, "port", "p", 2222, "Port the server listens on.")
}

func findHomeDir() string {
	usr, err := user.Current()
	if err != nil {
		fmt.Fprintf(os.Stderr, "determine current user: %v", err)
		os.Exit(1)
	}
	return usr.HomeDir
}

func parseAuthorizedKey() ssh.PublicKey {
	bs, err := ioutil.ReadFile(authorizedKeyFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read authorized key file: %v", err)
		os.Exit(1)
	}
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(bs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse authorized key file: %v", err)
		os.Exit(1)
	}
	return pubKey
}

func printResults(commandC <-chan []string, stdinC <-chan []byte, errC <-chan error) {
	for {
		select {
		case cmd := <-commandC:
			fmt.Printf("Received command: %s\n", strings.Join(cmd, " "))
		case stdin := <-stdinC:
			fmt.Printf("Received stdin: %s\n", stdin)
		case err := <-errC:
			fmt.Fprintf(os.Stderr, "Received error: %v\n", err)
		}
	}
}

func main() {
	var (
		commandC = make(chan []string, 1)
		stdinC   = make(chan []byte, 1)
		errC     = make(chan error, 1)
	)
	pflag.Parse()

	if strings.HasPrefix(authorizedKeyFile, "~") {
		homeDir := findHomeDir()
		authorizedKeyFile = filepath.Join(homeDir, authorizedKeyFile[1:])
	}
	pubKey := parseAuthorizedKey()
	server := &remote.SSHTestServer{
		AuthorizedKeys: []ssh.PublicKey{pubKey},
		Command:        commandC,
		Stdin:          stdinC,
		Stdout:         []byte("Hello from SSHTestServer\n"),
		Stderr:         []byte("Hello from SSHTestServer on stderr\n"),
		Errors:         errC,
	}
	go printResults(commandC, stdinC, errC)

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	if err := netutil.ListenAndServe(server, netutil.WithAddr(addr)); err != nil {
		fmt.Fprintf(os.Stderr, "listen and serve: %v", err)
		os.Exit(1)
	}
}
