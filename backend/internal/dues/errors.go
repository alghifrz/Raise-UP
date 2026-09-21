package dues

import "errors"

var (
	ErrInvalidRequest       = errors.New("invalid request")
	ErrPeriodNotFound       = errors.New("dues period not found")
	ErrPeriodAlreadyExists  = errors.New("dues period already exists")
	ErrPaymentNotFound      = errors.New("dues payment not found")
	ErrPaymentAlreadyExists = errors.New("dues payment already exists")
	ErrInvalidPaymentAmount = errors.New("invalid payment amount")
	ErrResidentNotFound     = errors.New("resident not found")
	ErrReminderUnavailable  = errors.New("WhatsApp reminder service unavailable")
)
