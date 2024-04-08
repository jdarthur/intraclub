package common

import (
	"fmt"
)

type ListOfCrudRecords interface {
	Length() int
}

type CrudRecord interface {
	RecordType() string               // A string name for this record, used as the DB collection name
	OneRecord() CrudRecord            // One instance of the CrudRecord's type. Must be a pointer
	ListOfRecords() ListOfCrudRecords // List of instances of the CrudRecord's type. Must be a list of pointers
	SetId(id string)                  // Set's the ID field on the record. This function is why the OneRecord / ListOfRecords fields must return pointer types.
	GetId() string                    // Gets the ID field from the record
	Validatable                       // Validate the correctness of the record on DB create / update
}

func CheckExistenceOrError(provider DbProvider, record CrudRecord) error {
	_, exists, err := GetOne(provider, record)
	if err != nil {
		return err
	}

	if !exists {
		return RecordDoesNotExist(record)
	}

	return nil
}

func RecordDoesNotExist(record CrudRecord) error {
	return fmt.Errorf("%s with ID %s does not exist", record.RecordType(), record.GetId())
}
