package zfs

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

const (
	// FakeZFSExitTestFailed signals that FakeZFS exited due to a configuration
	// error.
	FakeZFSExitTestFailed = 254

	// KeyIsFakeZFSCmd is the name of the environment variable that enables the
	// fake zfs command.
	KeyIsFakeZFSCmd = "IS_FAKE_ZFS_CMD"

	// KeyFakeZFSExitCode is the name of the environment variable that contains
	// the exit code the fake zfs should return. If not set fake zfs will return
	// 0.
	KeyFakeZFSExitCode = "FAKE_ZFS_EXIT_CODE"

	// KeyFakeZFSOutFile is the path to the file containing the output fake zfs
	// should write to stdout.
	KeyFakeZFSOutFile = "FAKE_ZFS_OUT_FILE"

	// KeyFakeZFSErrFile is the path to the file containing the output fake zfs
	// should write to stderr.
	KeyFakeZFSErrFile = "FAKE_ZFS_ERR_FILE"
)

// Fake mocks the zfs program.
//
// Fake calls os.Exit!
func Fake() {
	if os.Getenv(KeyIsFakeZFSCmd) != "1" {
		return
	}
	zfsExitCode := os.Getenv(KeyFakeZFSExitCode)
	if zfsExitCode == "" {
		zfsExitCode = "0"
	}
	ec, err := strconv.Atoi(zfsExitCode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid value for %s: %s", KeyFakeZFSExitCode, zfsExitCode)
		os.Exit(FakeZFSExitTestFailed)
	}

	zfsOutFile := os.Getenv(KeyFakeZFSOutFile)
	if zfsOutFile != "" {
		stdout, err := ioutil.ReadFile(zfsOutFile)
		if err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "read fake zfs out file %s: %v", zfsOutFile, err)
			os.Exit(FakeZFSExitTestFailed)
		}
		fmt.Print(string(stdout))
	}

	zfsErrFile := os.Getenv(KeyFakeZFSErrFile)
	if zfsErrFile != "" {
		stderr, err := ioutil.ReadFile(zfsErrFile)
		if err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "read fake zfs out file %s: %v", zfsOutFile, err)
			os.Exit(FakeZFSExitTestFailed)
		}
		fmt.Fprint(os.Stderr, string(stderr))
	}

	os.Exit(ec)
}
