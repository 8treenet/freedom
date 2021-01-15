package freedom

import (
	"os"
)

var (
	// EnvProfileDir is the name of the environment variable for the search
	// directory of the profile
	EnvProfileDir = "FREEDOM_PROJECT_CONFIG"

	// ProfileENV is the name of the profile directory in environment variable
	ProfileENV = EnvProfileDir

	// TODO(coco): this variable seems has no effect, considering remove it.
	// cachedProfileDir is the cache of the profile path
	cachedProfileDir string
)

// ProfileDirFromEnv reads from environment variable with EnvProfileDir
func ProfileDirFromEnv() string {
	cachedProfileDir = os.Getenv(EnvProfileDir)
	return cachedProfileDir
}