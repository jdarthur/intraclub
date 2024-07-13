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
	ValidateDynamic(db DbProvider, isUpdate bool, previousState CrudRecord) error
}

func ValueMustBeGloballyUnique(db DbProvider, record CrudRecord, key string, value interface{}) error {

	filter := map[string]interface{}{
		key: value,
	}

	records, err := db.GetAllWhere(record, filter)
	if err != nil {
		return err
	}

	if records.Length() != 0 {
		return ApiError{
			References: []interface{}{record.RecordType(), key, value},
			Code:       FieldMustBeGloballyUnique,
		}
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
		return ApiError{
			References: []any{field},
			Code:       FieldNotUpdatable,
		}
	}

	return nil
}

func GetOneByIdAndValidate(db DbProvider, c CrudRecord, id string) error {

	record, err := GetOneByStringId(db, c, id)
	if err != nil {
		return err
	}

	return ValidateStaticAndDynamic(db, record)
}

func ValidateStaticAndDynamic(db DbProvider, record Validatable) error {
	err := record.ValidateStatic()
	if err != nil {
		return err
	}

	return record.ValidateDynamic(db, false, nil)
}
