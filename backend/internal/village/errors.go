package village

import "errors"

var (
	ErrInvalidRequest   = errors.New("invalid request")
	ErrProfileNotFound  = errors.New("village profile not found")
	ErrOfficialNotFound = errors.New("village official not found")
)
