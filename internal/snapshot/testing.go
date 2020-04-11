package snapshot

import (
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
