package announcement

import "errors"

var (
	ErrInvalidRequest          = errors.New("invalid request")
	ErrNotFound                = errors.New("announcement not found")
	ErrResidentNotFound        = errors.New("resident not found")
	ErrRecipientRequired       = errors.New("private announcement requires at least one recipient")
	ErrInvalidStatusTransition = errors.New("invalid announcement status transition")
)
