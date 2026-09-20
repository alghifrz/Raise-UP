package chat

import "errors"

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrNotFound       = errors.New("conversation not found")
	ErrForbidden      = errors.New("forbidden")
	ErrUserNotFound   = errors.New("user not found")
)
