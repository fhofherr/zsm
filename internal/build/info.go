package build

import "time"

// Build coordinates.
var (
	Version = "dev"
	Commit  = "dev"
	Date    = time.Now().UTC().Format(time.RFC3339)
)
