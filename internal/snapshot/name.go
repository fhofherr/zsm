package snapshot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// TimestampFormat represents the format of timestamps used by zsm to identify
// snapshots.
const TimestampFormat = time.RFC3339Nano

// Name represents a named snapshot created by zsm.
type Name struct {
	FileSystem string    `json:"fileSystem"`
	Timestamp  time.Time `json:"timestamp"`
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

// ParseNameJSON parses a JSON representation of a name.
func ParseNameJSON(data []byte) (Name, error) {
	var n Name

	if err := json.Unmarshal(data, &n); err != nil {
		return n, fmt.Errorf("name from JSON: %w", err)
	}
	return n, nil
}

func (n Name) String() string {
	return fmt.Sprintf("%s@%s", n.FileSystem, n.Timestamp.Format(TimestampFormat))
}

// ToJSON converts the name to a JSON representation.
func (n Name) ToJSON() ([]byte, error) {
	var bs bytes.Buffer
	err := n.ToJSONW(&bs)
	return bs.Bytes(), err
}

// ToJSONW converts the name to a JSON representation and writes it to w.
func (n Name) ToJSONW(w io.Writer) error {
	if err := json.NewEncoder(w).Encode(n); err != nil {
		return fmt.Errorf("name to JSON: %w", err)
	}
	return nil
}
