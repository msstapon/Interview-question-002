package domain

import "errors"

// Sentinel errors that cross layer boundaries. Wrap with fmt.Errorf("...: %w", err)
// rather than redefining.
var (
	ErrNotFound           = errors.New("not found")
	ErrConflict           = errors.New("conflict")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
