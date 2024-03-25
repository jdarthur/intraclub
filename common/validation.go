package common

type Validatable interface {
	// ValidateStatic validates static constraints, e.g. number is in range, value
	// is non-empty, etc.
	//
	// This is designed to be a faster method than ValidateDynamic which can be called when
	// the data source is trusted, for example DB-to-DB calculation methods
	ValidateStatic() error

	// ValidateDynamic validates dynamic constraints, e.g. that the referenced UUID
	// exists in DB, API caller is the correct object owner, API caller is a relevant
	// captain or the commissioner, etc.
	//
	// This should always be called when accepting data from an untrusted source, for
	// example from a `POST` request on an API endpoint.
	ValidateDynamic(db DbProvider) error
}
