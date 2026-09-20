package complaint

import "errors"

var (
	ErrInvalidRequest          = errors.New("invalid request")
	ErrNotFound                = errors.New("complaint not found")
	ErrResidentNotFound        = errors.New("resident not found")
	ErrInvalidStatusTransition = errors.New("invalid complaint status transition")
)
