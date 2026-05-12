package database

import (
	"context"
)

type Validatable interface {
	StaticallyValid() error
}

type DatabaseValidatable interface {
	Validatable
	DynamicallyValid(ctx context.Context, db Provider) error
}

func Validate(ctx context.Context, db Provider, d DatabaseValidatable) error {
	err := d.StaticallyValid()
	if err != nil {
		return err
	}

	return d.DynamicallyValid(ctx, db)
}
