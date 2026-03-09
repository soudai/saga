package version

import "fmt"

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func String() string {
	return fmt.Sprintf("saga version=%s commit=%s date=%s", Version, Commit, BuildDate)
}
