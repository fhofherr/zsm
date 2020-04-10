package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fhofherr/zsm/internal/config"
	"github.com/fhofherr/zsm/internal/snapshot"
	"github.com/stretchr/testify/mock"
)

// TestCase tests the zsm command.
type TestCase struct {
	Name         string
	MakeArgs     func(*testing.T) []string
	MakeSMMock   func(*testing.T) *MockSnapshotManager
	AssertSMMock func(*testing.T, *MockSnapshotManager)
}

// Run runs the test case.
func (tt *TestCase) Run(t *testing.T) {
	t.Helper()

	if tt.MakeArgs == nil {
		t.Fatal("MakeArgs not set")
	}
	if tt.MakeSMMock == nil {
		t.Fatal("MakeSMMock not set")
	}

	msm := tt.MakeSMMock(t)
	msm.Test(t)

	smf := mockSnapshotManagerFactory(msm)
	zsmCmd := NewZSMCommand(WithSnapshotManagerFactory(smf))

	zsmCmd.SetArgs(tt.MakeArgs(t))
	if err := zsmCmd.Execute(); err != nil {
		t.Errorf("zsm failed: %v", err)
	}
	msm.AssertExpectations(t)

	if tt.AssertSMMock != nil {
		tt.AssertSMMock(t, msm)
	}
}

// RunTests runs all passed tests as sub-tests of t.
func RunTests(t *testing.T, tests []TestCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

// ConfigFile find a config file with name in the tests testdata directory.
func ConfigFile(t *testing.T, name string) string {
	cfgFile := filepath.Join("testdata", t.Name(), name)
	_, err := os.Stat(cfgFile)
	if os.IsNotExist(err) {
		t.Fatalf("%s does not exist", cfgFile)
	}
	return cfgFile
}

func mockSnapshotManagerFactory(msm *MockSnapshotManager) SnapshotManagerFactory {
	return func(cfg *zsmCommandConfig) (SnapshotManager, error) {
		msm.ZFS = cfg.V.GetString(config.ZFSCmd)
		return msm, nil
	}
}

// MockSnapshotManager is a mock implementation of the SnapshotManager interface.
type MockSnapshotManager struct {
	mock.Mock

	ZFS string
}

// CreateSnapshots registers a call to CreateSnapshots.
func (m *MockSnapshotManager) CreateSnapshots(opts ...snapshot.CreateOption) error {
	var args mock.Arguments
	if opts == nil {
		args = m.Called()
	} else {
		args = m.Called(opts)
	}
	return args.Error(0)
}

// CleanSnapshots registers a call to CleanSnapshots.
func (m *MockSnapshotManager) CleanSnapshots(cfg snapshot.BucketConfig) error {
	args := m.Called(cfg)
	return args.Error(0)
}
