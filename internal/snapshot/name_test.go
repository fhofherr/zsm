package snapshot_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/fhofherr/zsm/internal/snapshot"
	"github.com/stretchr/testify/assert"
)

func TestParseName(t *testing.T) {
	nowUTC := time.Now().UTC()
	tests := []struct {
		name        string
		snapNameStr string
		expected    snapshot.Name
		expectedOk  bool
	}{
		{
			name:        "valid snapshot",
			snapNameStr: fmt.Sprintf("%s@%s", "zsm_test/fs_1", nowUTC.Format(snapshot.TimestampFormat)),
			expected: snapshot.Name{
				FileSystem: "zsm_test/fs_1",
				Timestamp:  nowUTC,
			},
			expectedOk: true,
		},
		{
			name:        "invalid timestamp",
			snapNameStr: "zsm_test@tuesday",
		},
		{
			name:        "missing file system",
			snapNameStr: fmt.Sprintf("@%s", nowUTC.Format(snapshot.TimestampFormat)),
		},
		{
			name:        "invalid snapshot name",
			snapNameStr: "invalid_name",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			name, ok := snapshot.ParseName(tt.snapNameStr)
			assert.Equal(t, tt.expectedOk, ok)
			assert.Equal(t, tt.expected, name)
		})
	}
}

func TestParseNameJSON(t *testing.T) {
	name := snapshot.Name{FileSystem: "zsm_test", Timestamp: time.Now().UTC()}
	nameJSON, err := name.ToJSON()
	if !assert.NoError(t, err) {
		return
	}
	parsed, err := snapshot.ParseNameJSON(nameJSON)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, name, parsed)
}

func TestName_String(t *testing.T) {
	nowUTC := time.Now().UTC()
	fs := "zsm_test/fs_1"
	name := snapshot.Name{FileSystem: fs, Timestamp: nowUTC}

	nameStr := name.String()
	assert.Equal(t, nameStr, fmt.Sprintf("%s@%s", fs, nowUTC.Format(snapshot.TimestampFormat)))

	parsed, ok := snapshot.ParseName(nameStr)
	assert.True(t, ok, "nameStr cannot be parsed")
	assert.Equal(t, name, parsed)
}
