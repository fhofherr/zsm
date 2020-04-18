package remote_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/fhofherr/zsm/internal/remote"
	"github.com/fhofherr/zsm/internal/snapshot"
	"github.com/stretchr/testify/assert"
)

func TestHost_Dial_RemoteServerNotAvailable(t *testing.T) {
	host := &remote.Host{
		// Port does not really matter here, there should be nothing listening
		// on it.
		Addr: "127.0.0.1:1234",
	}
	err := host.Dial()
	assert.EqualError(t, err, "dial ssh: dial tcp 127.0.0.1:1234: connect: connection refused")
}

func TestHost_Dial(t *testing.T) {
	tests := []remote.TestCase{
		{
			Name: "not connected",
			Call: func(t *testing.T, host *remote.Host) error {
				// The actual method being called does not matter here. What's
				// important is, that we did not call host.Dial before
				_, err := host.ListSnapshots()
				if !assert.EqualError(t, err, "not connected") {
					return err
				}
				return nil
			},
		},
		{
			Name: "already connected",
			Call: func(t *testing.T, host *remote.Host) error {
				if err := host.Dial(); err != nil {
					t.Fatal(err)
				}
				defer host.Close()

				// This should return nil
				return host.Dial()
			},
		},
	}

	remote.RunTests(t, tests)
}

func TestHost_ListSnapshots(t *testing.T) {
	snapshots := snapshot.FakeNames(t, snapshot.Name{
		FileSystem: "zsm_test/fs_1",
		Timestamp:  time.Now().UTC(),
	}, snapshot.Minute, 5)

	tests := []remote.TestCase{
		{
			Name: "remote host has no snapshots",
			Call: func(t *testing.T, host *remote.Host) error {
				if err := host.Dial(); err != nil {
					t.Fatal(err)
				}
				defer host.Close()

				actual, err := host.ListSnapshots()
				assert.NoError(t, err)
				assert.Empty(t, actual)

				return err
			},
			ZSMCommand: []string{"/path/to/remote/zsm", "list", "-o", "jsonl"},
		},
		{
			Name: "remote host has snapshots",
			Call: func(t *testing.T, host *remote.Host) error {
				if err := host.Dial(); err != nil {
					t.Fatal(err)
				}
				defer host.Close()

				actual, err := host.ListSnapshots()
				assert.NoError(t, err)
				assert.Equal(t, snapshots, actual)

				return err
			},
			ZSMCommand: []string{"/path/to/remote/zsm", "list", "-o", "jsonl"},
			Stdout: func(t *testing.T) []byte {
				var bs bytes.Buffer

				for _, name := range snapshots {
					if err := name.ToJSONW(&bs); err != nil {
						t.Fatal(err)
					}
				}
				return bs.Bytes()
			},
		},
		{
			Name: "remote host returns error",
			Call: func(t *testing.T, host *remote.Host) error {
				if err := host.Dial(); err != nil {
					t.Fatal(err)
				}
				defer host.Close()

				_, err := host.ListSnapshots()
				return err
			},
			Stderr: func(t *testing.T) []byte {
				return []byte("remote zsm list wrote this to stderr")
			},
			ZSMExitCode: 10,
			ZSMCommand:  []string{"/path/to/remote/zsm", "list", "-o", "jsonl"},
		},
	}
	remote.RunTests(t, tests)
}

func TestHost_ReceiveSnapshot(t *testing.T) {
	tests := []remote.TestCase{
		{
			Name: "receive snapshot data",
			Call: func(t *testing.T, host *remote.Host) error {
				if err := host.Dial(); err != nil {
					t.Fatal(err)
				}
				defer host.Close()

				name := snapshot.MustParseName(t, "zsm_test@2020-04-10T09:45:58.564585005Z")
				data := bytes.NewReader([]byte("this is the snapshot data"))
				return host.ReceiveSnapshot("target_fs", name, data)
			},
			ZSMCommand: []string{
				"/path/to/remote/zsm",
				"receive",
				"target_fs",
				"zsm_test@2020-04-10T09:45:58.564585005Z",
			},
			Stdin: func(t *testing.T) []byte {
				return []byte("this is the snapshot data")
			},
		},
		{
			Name: "remote host returns error",
			Call: func(t *testing.T, host *remote.Host) error {
				if err := host.Dial(); err != nil {
					t.Fatal(err)
				}
				defer host.Close()

				name := snapshot.MustParseName(t, "zsm_test@2020-04-10T09:45:58.564585005Z")
				data := bytes.NewReader([]byte("this is the snapshot data"))
				return host.ReceiveSnapshot("target_fs", name, data)
			},
			ZSMExitCode: 10,
			Stderr: func(t *testing.T) []byte {
				return []byte("remote zsm receive wrote this to stderr")
			},
			Stdin: func(t *testing.T) []byte {
				return []byte("this is the snapshot data")
			},
		},
	}

	remote.RunTests(t, tests)
}
