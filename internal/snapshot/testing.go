package snapshot

import (
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/fhofherr/zsm/internal/zfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockZFSAdapter mocks calls to the ZFS executable installed on the system.
type MockZFSAdapter struct {
	mock.Mock
}

// CreateSnapshot registers a mock call to zfs snapshot
func (m *MockZFSAdapter) CreateSnapshot(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

// List registers a mock call to zfs list
func (m *MockZFSAdapter) List(typ zfs.ListType) ([]string, error) {
	args := m.Called(typ)
	return args.Get(0).([]string), args.Error(1)
}

// Destroy registers a call to zfs destroy.
func (m *MockZFSAdapter) Destroy(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

// Receive registers a call to zfs receive.
func (m *MockZFSAdapter) Receive(name string, r io.Reader) error {
	args := m.Called(name, r)
	return args.Error(0)
}

// Send registers a call to zfs send.
func (m *MockZFSAdapter) Send(name, ref string, w io.Writer) error {
	args := m.Called(name, ref, w)
	return args.Error(0)
}

// AssertNameFormat asserts that the passed snapName has the expected format
// for a snapshot of a filesystem with name fsName.
func AssertNameFormat(t *testing.T, fsName, snapName string) bool {
	name, ok := ParseName(snapName)
	if !ok {
		t.Errorf("unexpected snapshot name: %s", snapName)
	}
	if name.Timestamp.Location() != time.UTC {
		t.Errorf("timestamp not UTC: %v", name.Timestamp.Location())
		ok = false
	}
	return ok
}

// FakeNames creates n fake snapshots Names. Two consecutive names are delta
// apart. The last snapshot Name is end.
func FakeNames(t *testing.T, end Name, delta Interval, n int) []Name {
	names := make([]Name, n)
	ts := end.Timestamp
	for i := n - 1; i >= 0; i-- {
		names[i] = Name{
			FileSystem: end.FileSystem,
			Timestamp:  ts,
		}
		switch delta {
		case Minute:
			ts = ts.Add(-time.Minute)
		case Hour:
			ts = ts.Add(-time.Hour)
		case Day:
			ts = ts.AddDate(0, 0, -1)
		case Week:
			ts = ts.AddDate(0, 0, -7)
		case Month:
			ts = ts.AddDate(0, -1, 0)
		case Year:
			ts = ts.AddDate(-1, 0, 0)
		default:
			t.Fatalf("unsupported interval: %s", delta)
		}
	}

	return names
}

// EqualCreateOptions returns a function that checks if the passed CreateOptions
// match the expected create options.
//
// EqualCreateOptions is mainly intended for use with mock.MatchedBy.
func EqualCreateOptions(t *testing.T, expectedOpts ...CreateOption) func([]CreateOption) bool {
	expected := createOpts{}
	for _, opt := range expectedOpts {
		opt(&expected)
	}

	return func(actualOpts []CreateOption) bool {
		actual := createOpts{}
		for _, opt := range actualOpts {
			opt(&actual)
		}
		// Use assert.Equal instead of cmp.Equal as this fails the test with
		// additional output if expected and actual don't match.
		return assert.Equal(t, expected, actual)
	}
}

// ShuffleNamesC returns a shuffled copy of names.
func ShuffleNamesC(names []Name) []Name {
	shuffled := append([]Name{}, names...)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled
}

// MustParseName parses the passed snapshot name using ParseName.
// If ParseName returns false for its second argument, MustParseName fails the
// test.
func MustParseName(t *testing.T, name string) Name {
	t.Helper()

	parsed, ok := ParseName(name)
	if !ok {
		t.Fatalf("Can't parse %s", name)
	}
	return parsed
}

// MustParseTime parses the passed value as a time.Time according to layout.
// It fails the test if the value cannot be parsed.
func MustParseTime(t *testing.T, layout, value string) time.Time {
	t.Helper()

	ts, err := time.Parse(layout, value)
	if err != nil {
		t.Fatalf("Parse time %s: %v", value, err)
	}
	return ts
}

// MockManager is a mock for Manager and provides the same methods.
//
// MockManager is useful across various packages. It is therefore defined
// once in the snapshot package and not in the packages, that use the
// functionality.
type MockManager struct {
	mock.Mock

	ZFS string

	expectedCreateOpts createOpts
	actualCreateOpts   createOpts

	expectedSendOpts sendOpts
	actualSendOpts   sendOpts
}

// CreateSnapshots registers a call to CreateSnapshots.
func (m *MockManager) CreateSnapshots(opts ...CreateOption) error {
	callArgs := make([]interface{}, len(opts))
	for i, opt := range opts {
		callArgs[i] = opt
		opt(&m.actualCreateOpts)
	}
	args := m.Called(callArgs...)
	return args.Error(0)
}

// ExpectCreateOptions sets the CreateOptions expected when CreateSnapshot is called.
func (m *MockManager) ExpectCreateOptions(opts ...CreateOption) {
	for _, opt := range opts {
		opt(&m.expectedCreateOpts)
	}
}

// AssertCreateOptions asserts that the expected send options were actually passed.
func (m *MockManager) AssertCreateOptions(t *testing.T) bool {
	return assert.Equal(t, m.expectedCreateOpts, m.actualCreateOpts)
}

// CleanSnapshots registers a call to CleanSnapshots.
func (m *MockManager) CleanSnapshots(cfg BucketConfig) error {
	args := m.Called(cfg)
	return args.Error(0)
}

// ListSnapshots registers a call to ListSnapshots.
func (m *MockManager) ListSnapshots() ([]Name, error) {
	args := m.Called()
	return args.Get(0).([]Name), args.Error(1)
}

// ReceiveSnapshot registers a call to ReceiveSnapshot.
func (m *MockManager) ReceiveSnapshot(targetFS string, name Name, r io.Reader) error {
	args := m.Called(targetFS, name, r)
	return args.Error(0)
}

// SendSnapshot registers a call to SendSnapshot.
func (m *MockManager) SendSnapshot(name Name, w io.Writer, opts ...SendOption) error {
	callArgs := []interface{}{name, w}
	for _, opt := range opts {
		callArgs = append(callArgs, opt)
		opt(&m.actualSendOpts)
	}
	args := m.Called(callArgs...)
	return args.Error(0)
}

// ExpectSendOptions sets the SendOptions expected when SendSnapshot is called.
func (m *MockManager) ExpectSendOptions(opts ...SendOption) {
	for _, opt := range opts {
		opt(&m.expectedSendOpts)
	}
}

// AssertSendOptions asserts that the expected send options were actually passed.
func (m *MockManager) AssertSendOptions(t *testing.T) bool {
	return assert.Equal(t, m.expectedSendOpts, m.actualSendOpts)
}
