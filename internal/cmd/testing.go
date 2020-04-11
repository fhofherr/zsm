package cmd

import (
	"bytes"
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
	MakeArgs     func(t *testing.T) []string
	MakeMSM      func(t *testing.T) *MockSnapshotManager
	AssertMSM    func(t *testing.T, msm *MockSnapshotManager)
	AssertOutput func(t *testing.T, stdout, stderr string)
}

// Run runs the test case.
func (tt *TestCase) Run(t *testing.T) {
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	t.Helper()

	if tt.MakeArgs == nil {
		t.Fatal("MakeArgs not set")
	}
	if tt.MakeMSM == nil {
		t.Fatal("MakeSMMock not set")
	}

	msm := tt.MakeMSM(t)
	msm.Test(t)

	smf := mockSnapshotManagerFactory(msm)
	zsmCmd := NewZSMCommand(
		WithSnapshotManagerFactory(smf),
		WithStdout(&stdout),
		WithStderr(&stderr),
	)

	zsmCmd.SetArgs(tt.MakeArgs(t))
	if err := zsmCmd.Execute(); err != nil {
		t.Errorf("zsm failed: %v", err)
	}
	msm.AssertExpectations(t)

	if tt.AssertMSM != nil {
		tt.AssertMSM(t, msm)
	}
	if tt.AssertOutput != nil {
		tt.AssertOutput(t, stdout.String(), stderr.String())
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

// ListSnapshots mocks the ListSnapshots method.
func (m *MockSnapshotManager) ListSnapshots() ([]snapshot.Name, error) {
	args := m.Called()
	return args.Get(0).([]snapshot.Name), args.Error(1)
}
