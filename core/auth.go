//go:build !cgo

package main

import (
	"crypto/subtle"
	"fmt"
	"os"
	"strings"
	"sync"
)

const ipcTokenEnv = "FLCLASH_IPC_TOKEN"

var (
	sessionToken   string
	sessionAuthed  bool
	sessionAuthMu  sync.Mutex
	requireSession bool
)

func initSessionAuth() error {
	sessionToken = strings.TrimSpace(os.Getenv(ipcTokenEnv))
	// Desktop (!cgo) always requires a session token so unauthenticated
	// local peers cannot drive privileged actions.
	if sessionToken == "" {
		return fmt.Errorf("missing %s: refuse to start without session token", ipcTokenEnv)
	}
	if len(sessionToken) < 32 {
		return fmt.Errorf("%s too short", ipcTokenEnv)
	}
	requireSession = true
	return nil
}

func isSessionAuthed() bool {
	sessionAuthMu.Lock()
	defer sessionAuthMu.Unlock()
	return sessionAuthed
}

func authenticateSession(token string) error {
	sessionAuthMu.Lock()
	defer sessionAuthMu.Unlock()
	if !requireSession {
		sessionAuthed = true
		return nil
	}
	if subtle.ConstantTimeCompare([]byte(token), []byte(sessionToken)) != 1 {
		return fmt.Errorf("invalid session token")
	}
	sessionAuthed = true
	return nil
}
