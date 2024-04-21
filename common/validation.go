package common

import (
	"fmt"
)

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
	ValidateDynamic(db DbProvider, isUpdate bool, previousState CrudRecord) error
}

func ValueMustBeGloballyUnique(db DbProvider, record CrudRecord, key string, value interface{}) error {

	filter := map[string]interface{}{
		key: value,
	}

	records, err := db.GetAllWhere(record, filter)
	if err != nil {
		panic(err)
		return err
	}

	if records.Length() != 0 {
		return fmt.Errorf("%s with %s %v already exists", record.RecordType(), key, value)
	}

	return nil
}

type HasNonUpdatable interface {
	VerifyUpdatable(c CrudRecord) (illegalUpdate bool, field string)
	CrudRecord
}

func CheckNonUpdatableFields(request HasNonUpdatable, db DbProvider) error {

	recordInDb, exists, err := GetOne(db, request)
	if err != nil {
		return err
	}
	if !exists {
		return RecordDoesNotExist(request)
	}

	illegalUpdate, field := request.VerifyUpdatable(recordInDb)
	if illegalUpdate {
		return fmt.Errorf("field '%s' may not be changed after creation", field)
	}

	return nil
}
