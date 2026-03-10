package version

import "fmt"

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func String() string {
	return fmt.Sprintf("sg version=%s commit=%s date=%s", Version, Commit, BuildDate)
}
