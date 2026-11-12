package foldercheck

import (
	"path/filepath"
	"runtime"
	"strings"
)

// within reports whether target equals dir or lies below it, on a
// path-separator boundary. Windows paths compare case-insensitively.
func within(target, dir string) bool {
	if samePath(target, dir) {
		return true
	}
	if runtime.GOOS == "windows" {
		target, dir = strings.ToLower(target), strings.ToLower(dir)
	}
	if !strings.HasSuffix(dir, string(filepath.Separator)) {
		dir += string(filepath.Separator)
	}
	return strings.HasPrefix(target, dir)
}

func samePath(a, b string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}
