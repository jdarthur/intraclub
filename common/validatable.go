package common

type Validatable interface {
	StaticallyValid() error
}

type DatabaseValidatable interface {
	Validatable
	DynamicallyValid(db DatabaseProvider, existing DatabaseValidatable) error
}

func Validate(db DatabaseProvider, d DatabaseValidatable, existing DatabaseValidatable) error {
	err := d.StaticallyValid()
	if err != nil {
		return err
	}
	return d.DynamicallyValid(db, existing)
}
