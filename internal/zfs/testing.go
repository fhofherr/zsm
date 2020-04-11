package zfs

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCase represents a test case for a call to the zfs adapter.
type TestCase struct {
	Name string

	// Call calls the respective adapter function; should return any error
	// returned by the zfs adapter. May use the passed *testing.T to do further
	// assertions on the result of calling zsf.
	Call func(*testing.T, Adapter) error

	ZFSArgs     []string                  // expected arguments
	ZFSExitCode int                       // expected exit code of zfs
	Stdin       func(t *testing.T) []byte // returns stdin expected to be sent to zfs
	Stdout      func(t *testing.T) []byte // returns stdout expected to be sent by zfs
	Stderr      func(t *testing.T) []byte // returns stderr expected to be sent by zfs
}

func (tt *TestCase) run(t *testing.T, fake *fakeZFS) {
	var zfsArgs []string

	t.Helper()
	cmdFunc, zfsDir := fake.CmdFunc(t, &zfsArgs, tt)

	err := tt.Call(t, Adapter(cmdFunc))
	if err != nil {
		zfsErr := &Error{}
		if !errors.As(err, &zfsErr) {
			t.Errorf("unexpected zfs error: %v", err)
		}
		assert.Equal(t, tt.ZFSExitCode, zfsErr.ExitCode, "zfs did not return the expected exit code")

		expectedStderr := []byte("")
		if tt.Stderr != nil {
			expectedStderr = tt.Stderr(t)
		}
		assert.Equal(t, expectedStderr, []byte(zfsErr.Stderr), "zfs wrote something unexpected to stderr")
	}

	assert.Equal(t, tt.ZFSArgs, zfsArgs, "zfs was not called with the expected args")

	// Check that zfs received what we meant to be sent via stdin
	if tt.Stdin != nil {
		expected := tt.Stdin(t)
		actual, err := ioutil.ReadFile(filepath.Join(zfsDir, "stdin"))
		if err != nil {
			t.Fatalf("read zfs stdin: %v", err)
		}
		assert.Equal(t, expected, actual, "zfs did not receive the expected stdin")
	}
}

// RunTests runs all tests. If parallel is true t.Parallel is called for each
// test.
func RunTests(t *testing.T, tests []TestCase, parallel bool) {
	fake := &fakeZFS{}
	fake.run(t)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			if parallel {
				t.Parallel()
			}
			tt.run(t, fake)
		})
	}
}

const (
	keyFakeZFSDir      = "ZSM_TEST_FAKE_ZFS_DIR"
	keyFakeZFSExitCode = "ZSM_TEST_FAKE_ZFS_EXIT_CODE"
	fakeZFSFailed      = 254
)

type fakeZFS struct {
	TestName string
}

func (f *fakeZFS) CmdFunc(t *testing.T, args *[]string, tt *TestCase) (CmdFunc, string) {
	zfsDir, err := ioutil.TempDir("", "zsm-test-zfs-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.RemoveAll(zfsDir)
	})

	if tt.Stdout != nil {
		if err := ioutil.WriteFile(filepath.Join(zfsDir, "stdout"), tt.Stdout(t), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if tt.Stderr != nil {
		if err := ioutil.WriteFile(filepath.Join(zfsDir, "stderr"), tt.Stderr(t), 0600); err != nil {
			t.Fatal(err)
		}
	}

	cmdFunc := NewCmdFunc(os.Args[0], fmt.Sprintf("-test.run=%s", f.TestName))
	cmdFunc = WithEnv(cmdFunc, map[string]string{
		keyFakeZFSExitCode: strconv.Itoa(tt.ZFSExitCode),
		keyFakeZFSDir:      zfsDir,
	})
	cmdFunc = SwallowFurtherArgs(cmdFunc, args)
	return cmdFunc, zfsDir
}

func (f *fakeZFS) run(t *testing.T) {
	if strings.ContainsRune(t.Name(), '/') {
		t.Fatal("Can't run fake zfs from a sub-test")
	}
	f.TestName = t.Name()

	fakeZFSDir := os.Getenv(keyFakeZFSDir)
	if fakeZFSDir == "" {
		return
	}

	if err := f.saveInput(filepath.Join(fakeZFSDir, "stdin")); err != nil {
		fmt.Fprintf(os.Stderr, "save stdin: %v", err)
		os.Exit(fakeZFSFailed)
	}
	if err := f.sendOutput(filepath.Join(fakeZFSDir, "stdout"), os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "send stdout: %v", err)
		os.Exit(fakeZFSFailed)
	}
	if err := f.sendOutput(filepath.Join(fakeZFSDir, "stderr"), os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "send stderr: %v", err)
		os.Exit(fakeZFSFailed)
	}

	ecStr := os.Getenv(keyFakeZFSExitCode)
	if ecStr == "" {
		ecStr = "0"
	}
	ec, err := strconv.Atoi(ecStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid exit code: %s", ecStr)
		ec = fakeZFSFailed
	}
	os.Exit(ec)
}

func (f *fakeZFS) sendOutput(path string, w io.Writer) error {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("fake zfs: read file %s: %w", path, err)
	}
	if _, err := w.Write(bs); err != nil {
		return fmt.Errorf("fake zfs: send output %s: %w", path, err)
	}
	return nil
}

func (f *fakeZFS) saveInput(path string) error {
	bs, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("fake zfs: read stdin: %w", err)
	}
	if err := ioutil.WriteFile(path, bs, 0600); err != nil {
		return fmt.Errorf("fake zfs: save input %s: %w", path, err)
	}
	return nil
}
