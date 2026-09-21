package gallery

import "errors"

var (
	ErrInvalidRequest     = errors.New("invalid request")
	ErrNotFound           = errors.New("gallery item not found")
	ErrStorageUnavailable = errors.New("gallery storage is not configured")
)
