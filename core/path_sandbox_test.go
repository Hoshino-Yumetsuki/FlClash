package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/metacubex/mihomo/constant"
)

func TestEnsurePathUnderHome(t *testing.T) {
	home := t.TempDir()
	constant.SetHomeDir(home)

	inside := filepath.Join(home, "profiles", "a.yaml")
	if err := os.MkdirAll(filepath.Dir(inside), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := ensurePathUnderHome(inside); err != nil {
		t.Fatalf("inside path should be allowed: %v", err)
	}

	outside := filepath.Join(home, "..", "escape.txt")
	if _, err := ensurePathUnderHome(outside); err == nil {
		t.Fatal("path outside home should be rejected")
	}

	// classic traversal
	traversal := filepath.Join(home, "profiles", "..", "..", "etc", "passwd")
	if _, err := ensurePathUnderHome(traversal); err == nil {
		t.Fatal("traversal path should be rejected")
	}
}
