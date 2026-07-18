//go:build !cgo && !windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// refuseDirectSetuidLaunch blocks the LPE path where an unprivileged user
// executes a setuid-root FlClashCore with an attacker-controlled IPC address.
// When effective uid is root but real uid is not, require the parent process
// executable to be the installed FlClash UI next to this core binary.
func refuseDirectSetuidLaunch() error {
	euid := os.Geteuid()
	ruid := os.Getuid()
	if euid != 0 || ruid == 0 {
		return nil
	}
	ppid := os.Getppid()
	if ppid <= 1 {
		return fmt.Errorf("setuid core: refusing launch without trusted parent")
	}
	parentPath, err := parentProcessPath(ppid)
	if err != nil {
		return fmt.Errorf("setuid core: cannot verify parent: %w", err)
	}
	parentPath, err = filepath.EvalSymlinks(parentPath)
	if err != nil {
		// If parent path cannot be resolved, fail closed under setuid.
		return fmt.Errorf("setuid core: cannot resolve parent path: %w", err)
	}
	parentPath = filepath.Clean(parentPath)

	coreExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("setuid core: cannot resolve self: %w", err)
	}
	coreExe, err = filepath.EvalSymlinks(coreExe)
	if err != nil {
		coreExe = filepath.Clean(coreExe)
	} else {
		coreExe = filepath.Clean(coreExe)
	}
	dir := filepath.Dir(coreExe)

	// Only allow exact UI binaries shipped next to FlClashCore.
	allowed := []string{
		filepath.Join(dir, "FlClash"),
		filepath.Join(dir, "flclash"),
	}
	// macOS app bundle: .../FlClash.app/Contents/MacOS/FlClash
	if base := filepath.Base(dir); base == "MacOS" {
		allowed = append(allowed, filepath.Join(dir, "FlClash"))
	}

	for _, want := range allowed {
		want = filepath.Clean(want)
		if parentPath == want {
			return nil
		}
		// Compare basenames only when full path equals after EvalSymlinks of want if it exists.
		if st, err := os.Stat(want); err == nil && !st.IsDir() {
			if resolved, err := filepath.EvalSymlinks(want); err == nil && filepath.Clean(resolved) == parentPath {
				return nil
			}
		}
	}
	return fmt.Errorf("setuid core: parent %q is not installed FlClash UI", parentPath)
}

func parentProcessPath(ppid int) (string, error) {
	// Linux: /proc/<ppid>/exe
	if target, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", ppid)); err == nil {
		return target, nil
	}
	// macOS: ps -o comm= may be only basename; prefer full path via ps -o command=
	if out, err := exec.Command("ps", "-p", strconv.Itoa(ppid), "-o", "command=").Output(); err == nil {
		fields := strings.Fields(strings.TrimSpace(string(out)))
		if len(fields) > 0 {
			// First field is executable path when launched with absolute path.
			return fields[0], nil
		}
	}
	if out, err := exec.Command("ps", "-p", strconv.Itoa(ppid), "-o", "comm=").Output(); err == nil {
		name := strings.TrimSpace(string(out))
		if name != "" {
			return name, nil
		}
	}
	return "", fmt.Errorf("empty parent path")
}
