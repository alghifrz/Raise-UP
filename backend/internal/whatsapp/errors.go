package whatsapp

import "errors"

var (
	ErrNotConfigured  = errors.New("whatsapp not configured")
	ErrInvalidRequest = errors.New("invalid request")
	ErrSendFailed     = errors.New("whatsapp send failed")
	ErrForbidden      = errors.New("forbidden")
)
