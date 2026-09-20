package complaint

import (
	"context"

	db "github.com/diuk/raiseup/db/generated"
)

// AdaptResidentReader wraps a resident GetByID function and maps not-found errors.
func AdaptResidentReader(getByID func(context.Context, string) (db.Resident, error), isNotFound func(error) bool) ResidentReader {
	return residentAdapter{getByID: getByID, isNotFound: isNotFound}
}

type residentAdapter struct {
	getByID    func(context.Context, string) (db.Resident, error)
	isNotFound func(error) bool
}

func (a residentAdapter) GetByID(ctx context.Context, id string) (db.Resident, error) {
	item, err := a.getByID(ctx, id)
	if err != nil {
		if a.isNotFound != nil && a.isNotFound(err) {
			return db.Resident{}, ErrResidentNotFound
		}
		return db.Resident{}, err
	}
	return item, nil
}
