package snapshot

import (
	"fmt"
	"strings"
	"time"
)

// TimestampFormat represents the format of timestamps used by zsm to identify
// snapshots.
const TimestampFormat = time.RFC3339Nano

// Name represents a named snapshot created by zsm.
type Name struct {
	FileSystem string
	Timestamp  time.Time
}

// ParseName parses a string representing a snapshot into a Name.
func ParseName(name string) (Name, bool) {
	parts := strings.Split(name, "@")
	if len(parts) != 2 {
		return Name{}, false
	}
	if parts[0] == "" {
		return Name{}, false
	}
	ts, err := time.Parse(TimestampFormat, parts[1])
	if err != nil {
		return Name{}, false
	}
	return Name{FileSystem: parts[0], Timestamp: ts}, true
}

func (n Name) String() string {
	return fmt.Sprintf("%s@%s", n.FileSystem, n.Timestamp.Format(TimestampFormat))
}
