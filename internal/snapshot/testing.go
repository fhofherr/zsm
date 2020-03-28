package snapshot

import (
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
