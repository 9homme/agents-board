// Package config resolves runtime configuration for the agent-board binaries.
// It is the ONLY production package introduced by REQ006 (architecture §3 D-012).
package config

import (
	"errors"
	"os"
)

// ResolveDBURL inspects the process environment and returns the database URL to use.
//
// Exactly one env var is accepted: DATABASE_URL. DB_URL is no longer supported and its
// presence causes a fatal, operator-actionable error (architecture §5.1 / D-006).
//
// The four env-state outcomes are:
//
//	DATABASE_URL set, DB_URL unset → return (url, nil)
//	Both set                       → return ("", error) — tells operator to remove DB_URL
//	Only DB_URL set                → return ("", error) — tells operator to rename to DATABASE_URL
//	Neither set                    → return ("", error) — tells operator DATABASE_URL is required
//
// The function never calls log.Fatal; the caller (main.go) decides exit behaviour.
func ResolveDBURL() (string, error) {
	dbURL := os.Getenv("DATABASE_URL")
	legacyURL := os.Getenv("DB_URL")
	switch {
	case dbURL != "" && legacyURL != "":
		return "", errors.New("DB_URL is set but no longer supported; remove DB_URL from your environment to disambiguate (DATABASE_URL is the sole accepted name as of REQ006/US010)")
	case dbURL != "":
		return dbURL, nil
	case legacyURL != "":
		return "", errors.New("DB_URL is no longer supported; rename to DATABASE_URL (REQ006/US010)")
	default:
		return "", errors.New("DATABASE_URL environment variable is required")
	}
}
