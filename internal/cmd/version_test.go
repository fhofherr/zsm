package cmd_test

import (
	"bytes"
	"testing"

	"github.com/fhofherr/zsm/internal/build"
	"github.com/fhofherr/zsm/internal/cmd"
	"github.com/fhofherr/zsm/internal/snapshot"
	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	tests := []cmd.TestCase{
		{
			Name: "print version",
			MakeArgs: func(t *testing.T) []string {
				return []string{"version"}
			},
			MakeMSM: func(t *testing.T) *snapshot.MockManager {
				return &snapshot.MockManager{}
			},
			AssertOutput: func(t *testing.T, stdout, _ string) {
				var buf bytes.Buffer

				err := build.WriteInfo(&buf)
				assert.NoError(t, err)
				assert.Equal(t, buf.String(), stdout)
			},
		},
	}

	cmd.RunTests(t, tests)
}
