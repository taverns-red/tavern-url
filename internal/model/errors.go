package model

import "errors"

// ErrNotFound is a generic not-found error for domain lookups.
var ErrNotFound = errors.New("not found")
