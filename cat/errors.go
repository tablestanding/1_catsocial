package cat

import "errors"

var (
	ErrCatNotFound                     = errors.New("cat not found")
	ErrCatSexEditedAfterMatchRequested = errors.New("cat sex edited after match has been requested")
)
