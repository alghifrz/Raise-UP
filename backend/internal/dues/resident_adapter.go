package dues

import (
	"context"

	db "github.com/diuk/raiseup/db/generated"
)

// ResidentReader loads residents to validate payment resident IDs.
type ResidentReader interface {
	GetByID(ctx context.Context, id string) (db.Resident, error)
}

type residentReaderFunc struct {
	getByID    func(context.Context, string) (db.Resident, error)
	isNotFound func(error) bool
}

// AdaptResidentReader wraps a resident GetByID function and maps not-found errors.
func AdaptResidentReader(getByID func(context.Context, string) (db.Resident, error), isNotFound func(error) bool) ResidentReader {
	return &residentReaderFunc{getByID: getByID, isNotFound: isNotFound}
}

func (r *residentReaderFunc) GetByID(ctx context.Context, id string) (db.Resident, error) {
	item, err := r.getByID(ctx, id)
	if err != nil {
		if r.isNotFound(err) {
			return db.Resident{}, ErrResidentNotFound
		}
		return db.Resident{}, err
	}
	return item, nil
}
