package freedom

import (
	"os"
)

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
