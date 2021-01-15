package freedom

import (
	"path"
)

// JoinPath returns a string that joins any number of path elements into a
// single path, separating them with slashes.
func JoinPath(elems ...string) string {
	return path.Join(elems...)
}
