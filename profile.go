package freedom

import (
	"fmt"
	"os"
	"path"

	"github.com/BurntSushi/toml"
)

var (
	// profileFallbackSearchDirs is a series of directory that is used to search
	// profile file if a profile file has not been found in other directory.
	profileFallbackSearchDirs = []string{"./conf", "./server/conf"}

	// configurator is a instance of Configurator. It is never nil.
	configurator Configurator

	// EnvProfileDir is the name of the environment variable for the search
	// directory of the profile.
	EnvProfileDir = "FREEDOM_PROJECT_CONFIG"

	// ProfileENV is the name of the profile directory in environment variable.
	ProfileENV = EnvProfileDir

	// cachedProfileDir is the cache of the profile's path.
	// TODO(coco): this variable seems has no effect, considering remove it.
	cachedProfileDir string
)

var _ Configurator = (*fallbackConfigurator)(nil)

// Configurer .
type Configurer = Configurator

// Configurator .
type Configurator interface {
	Configure(obj interface{}, file string, metaData ...interface{}) error
}

func initConfigurator() {
	SetConfigurator(newFallbackConfigurator())
}

// SetConfigurator assigns a Configurator to global configurator
func SetConfigurator(c Configurator) {
	configurator = c
}

// SetConfigurer assigns a Configurator to global configurator
func SetConfigurer(c Configurer) {
	SetConfigurator(c)
}

// Configure .
func Configure(obj interface{}, file string, metaData ...interface{}) error {
	return configurator.Configure(obj, file)
}

// detectProfileInFallbackSearchDirs accepts a string with the name of a
// profile file, and search the file in profileFallbackSearchDirs. It returns
// (a string with the path of the profile file, true) if the profile file has
// been found, and ("", false) otherwise.
func detectProfileInFallbackSearchDirs(file string) (string, bool) {
	for _, dir := range profileFallbackSearchDirs {
		filePath := JoinPath(dir, file)
		if IsDir(dir) && IsFile(filePath) {
			return filePath, true
		}
	}

	return "", false
}

// detectProfilePath accepts a string with the name of a profile file, and
// search the file in the directory which specified in environment variable.
// If the file has not been found, continue search the file by
// detectProfileInFallbackSearchDirs. It returns (a string with the path of
// the profile file, true) if the profile file has found, and ("", false)
// otherwise.
func detectProfilePath(file string) (string, bool) {
	dir := ProfileDirFromEnv()

	filePath := JoinPath(dir, file)
	if IsFile(filePath) {
		return filePath, true
	}

	return detectProfileInFallbackSearchDirs(file)
}

// ReadProfile accepts a string with the name of a profile file, and search
// the file by detectProfilePath. It will fill v with the configuration by
// parsing the profile into toml format, and returns nil if the file has
// found. It returns error, if the file has not been found or any error
// encountered.
func ReadProfile(file string, v interface{}) error {
	filePath, isFilePathExist := detectProfilePath(file)

	if !isFilePathExist {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	_, err := toml.DecodeFile(filePath, v)
	if err != nil {
		Logger().Errorf("[Freedom] configuration decode error: %s", err.Error())
		return err
	}

	Logger().Infof("[Freedom] configuration was found: %s", filePath)
	return nil
}

// fallbackConfigurator is used to act as a fallback if no any configurator
// are applied. It implements Configurator.
type fallbackConfigurator struct{}

// newFallbackConfigurator creates a fallbackConfigurator
func newFallbackConfigurator() *fallbackConfigurator {
	return &fallbackConfigurator{}
}

// Configured proxy a call to ReadProfile
func (*fallbackConfigurator) Configure(obj interface{}, file string, metaData ...interface{}) error {
	return ReadProfile(file, obj)
}

// JoinPath returns a string that joins any number of path elements into a
// single path, separating them with slashes.
func JoinPath(elems ...string) string {
	return path.Join(elems...)
}

// IsDir accepts a string with a directory path and tests the path. It returns
// true if the path exists and it is a directory, and false otherwise.
func IsDir(dir string) bool {
	stat, err := os.Stat(dir)
	if err != nil {
		return false
	}

	return stat.IsDir()
}

// IsFile accepts a string with a file path and tests the path. It returns
// true if the path exists and it is a file, and false otherwise.
func IsFile(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}

	return !stat.IsDir()
}

// ProfileDirFromEnv reads from environment variable with EnvProfileDir
func ProfileDirFromEnv() string {
	cachedProfileDir = os.Getenv(EnvProfileDir)
	return cachedProfileDir
}
