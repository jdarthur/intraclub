package common

type Validatable interface {
	StaticallyValid() error
}

type DatabaseValidatable interface {
	Validatable
	DynamicallyValid(db DatabaseProvider) error
}

func Validate(db DatabaseProvider, d DatabaseValidatable) error {
	err := d.StaticallyValid()
	if err != nil {
		return err
	}
	return d.DynamicallyValid(db)
}
