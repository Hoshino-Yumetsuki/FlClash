package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/metacubex/mihomo/constant"
)

// ensurePathUnderHome resolves path (including symlinks when present) and
// requires the final path to stay under mihomo HomeDir.
func ensurePathUnderHome(path string) (string, error) {
	home := constant.Path.HomeDir()
	if home == "" {
		return "", fmt.Errorf("home dir not initialized")
	}
	absHome, err := filepath.Abs(home)
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(absHome); err == nil {
		absHome = resolved
	}
	absHome = filepath.Clean(absHome)

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	// Resolve symlinks when the path (or a prefix) exists.
	if resolved, err := filepath.EvalSymlinks(absPath); err == nil {
		absPath = resolved
	} else {
		// Path may not exist yet (delete of missing file). Resolve existing prefix.
		absPath = cleanExistingPrefix(absPath)
	}
	absPath = filepath.Clean(absPath)

	sep := string(os.PathSeparator)
	if absPath != absHome && !strings.HasPrefix(absPath, absHome+sep) {
		return "", fmt.Errorf("path outside home dir")
	}
	return absPath, nil
}

func cleanExistingPrefix(path string) string {
	cur := path
	for {
		if resolved, err := filepath.EvalSymlinks(cur); err == nil {
			// Re-join unresolved suffix.
			suffix, err := filepath.Rel(cur, path)
			if err != nil || suffix == "." {
				return resolved
			}
			return filepath.Join(resolved, suffix)
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return path
		}
		cur = parent
	}
}
