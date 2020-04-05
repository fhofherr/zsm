package snapshot

import (
	"testing"
	"time"

	"github.com/fhofherr/zsm/internal/zfs"
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
