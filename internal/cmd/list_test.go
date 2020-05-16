package cmd_test

import (
	"strings"
	"testing"

	"github.com/fhofherr/zsm/internal/cmd"
	"github.com/fhofherr/zsm/internal/snapshot"
	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	tests := []cmd.TestCase{
		{
			Name: "plain text",
			MakeArgs: func(t *testing.T) []string {
				return []string{"list"}
			},
			MakeMSM: func(t *testing.T) *snapshot.MockManager {
				sns := []snapshot.Name{
					snapshot.MustParseName(t, "zfs_test@2020-04-10T09:45:58.564585005Z"),
					snapshot.MustParseName(t, "zfs_test@2020-04-10T09:44:58.564585005Z"),
				}
				sm := &snapshot.MockManager{}
				sm.On("ListSnapshots").Return(sns, nil)

				return sm
			},
			AssertOutput: func(t *testing.T, stdout, stderr string) {
				expected := []string{
					"zfs_test@2020-04-10T09:45:58.564585005Z",
					"zfs_test@2020-04-10T09:44:58.564585005Z",
				}
				actual := strings.Split(strings.TrimSpace(stdout), "\n")
				assert.Equal(t, expected, actual)
				assert.Empty(t, stderr)
			},
		},
		{
			Name: "json output",
			MakeArgs: func(t *testing.T) []string {
				return []string{"list", "-o", "jsonl"}
			},
			MakeMSM: func(t *testing.T) *snapshot.MockManager {
				sns := []snapshot.Name{
					snapshot.MustParseName(t, "zfs_test@2020-04-10T09:45:58.564585005Z"),
					snapshot.MustParseName(t, "zfs_test@2020-04-10T09:44:58.564585005Z"),
				}
				sm := &snapshot.MockManager{}
				sm.On("ListSnapshots").Return(sns, nil)

				return sm
			},
			AssertOutput: func(t *testing.T, stdout, stderr string) {
				expected := []snapshot.Name{
					snapshot.MustParseName(t, "zfs_test@2020-04-10T09:45:58.564585005Z"),
					snapshot.MustParseName(t, "zfs_test@2020-04-10T09:44:58.564585005Z"),
				}
				var actual []snapshot.Name
				for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
					name, err := snapshot.ParseNameJSON([]byte(line))
					if err != nil {
						t.Error(err)
						continue
					}
					actual = append(actual, name)
				}
				assert.Equal(t, expected, actual)
				assert.Empty(t, stderr)
			},
		},
	}
	cmd.RunTests(t, tests)
}
