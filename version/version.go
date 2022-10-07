package version

import "fmt"

const (
	RepoUrl = "https://github.com/cloudnativedaysjp/seaman"
)

var (
	Version = "REPLACEMENT"
	Commit  = "REPLACEMENT"
)

func Information() string {
	return fmt.Sprintf("Version %s (Commit: %s)\nRepoUrl: %s", Version, Commit, RepoUrl)
}
