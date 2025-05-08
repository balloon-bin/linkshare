package version

import "fmt"

var (
	Version        = "dev"
	GitCommit      = "unknown"
	CommitDateTime = "unknown"
	SchemaVersion  = 1
)

// PrintVersionInfo prints formatted version information to stdout
func Print() {
	fmt.Printf("Version:      %s\n", Version)
	fmt.Printf("Git commit:   %s %s\n", GitCommit, CommitDateTime)
	fmt.Printf("Schema:       v%d\n", SchemaVersion)
}
