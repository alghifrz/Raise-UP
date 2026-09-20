package uuidutil

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ToString converts a pgtype.UUID to its canonical string form.
func ToString(id pgtype.UUID) (string, error) {
	if !id.Valid {
		return "", fmt.Errorf("invalid uuid")
	}
	parsed, err := uuid.FromBytes(id.Bytes[:])
	if err != nil {
		return "", err
	}
	return parsed.String(), nil
}

// FromString parses a UUID string into pgtype.UUID.
func FromString(value string) (pgtype.UUID, error) {
	parsed, err := uuid.Parse(value)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: parsed, Valid: true}, nil
}
