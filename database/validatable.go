package database

import (
	"context"
)

type Validatable interface {
	StaticallyValid() error
}

type DatabaseValidatable interface {
	Validatable
	DynamicallyValid(ctx context.Context, db DatabaseProvider) error
}

func Validate(ctx context.Context, db DatabaseProvider, d DatabaseValidatable) error {
	err := d.StaticallyValid()
	if err != nil {
		return err
	}

	return d.DynamicallyValid(ctx, db)
}
