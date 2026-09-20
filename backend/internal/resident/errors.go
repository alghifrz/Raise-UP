package resident

import "errors"

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrNotFound       = errors.New("resident not found")
	ErrDeleteConflict = errors.New("resident delete conflict")
)
