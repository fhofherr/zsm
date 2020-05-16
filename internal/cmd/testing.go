package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/fhofherr/zsm/internal/config"
	"github.com/fhofherr/zsm/internal/snapshot"
)

// TestCase tests the zsm command.
type TestCase struct {
	Name         string
	MakeArgs     func(t *testing.T) []string
	MakeMSM      func(t *testing.T) *snapshot.MockManager
	AssertMSM    func(t *testing.T, msm *snapshot.MockManager)
	AssertOutput func(t *testing.T, stdout, stderr string)
}

func (tt *TestCase) run(t *testing.T) {
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
	msm.AssertCreateOptions(t)
	msm.AssertSendOptions(t)

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
		tt := tt
		t.Run(tt.Name, tt.run)
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

func mockSnapshotManagerFactory(msm *snapshot.MockManager) SnapshotManagerFactory {
	return func(cfg *zsmCommandConfig) (SnapshotManager, error) {
		msm.ZFS = cfg.V.GetString(config.ZFSCmd)
		return msm, nil
	}
}
