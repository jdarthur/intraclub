package database

import (
	"context"
	"fmt"
)

type UniquenessConstraint[T CrudRecord] interface {
	UniquenessEquivalent(other T) error
	CrudRecord
}

func ValidateUniqueConstraint[T CrudRecord](ctx context.Context, db Provider, c T) error {
	u, ok := any(c).(UniquenessConstraint[T])
	if ok {
		otherRecords, err := GetAllWhere[T](ctx, db, func(_ context.Context, c2 T) bool {
			return c.GetId() != c2.GetId()
		})

		for _, other := range otherRecords {
			err = u.UniquenessEquivalent(other)
			if err != nil {
				return fmt.Errorf("uniqueness constraint violated for %T: %s", c, err)
			}
		}
		return err
	}
	return nil
}
