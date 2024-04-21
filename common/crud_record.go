package common

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ListOfCrudRecords interface {
	Length() int
	Get(index int) CrudRecord
}

type CrudRecord interface {
	RecordType() string               // A string name for this record, used as the DB collection name
	OneRecord() CrudRecord            // One instance of the CrudRecord's type. Must be a pointer
	ListOfRecords() ListOfCrudRecords // List of instances of the CrudRecord's type. Must be a list of pointers
	SetId(id primitive.ObjectID)      // Set's the ID field on the record. This function is why the OneRecord / ListOfRecords fields must return pointer types.
	GetId() primitive.ObjectID        // Gets the ID field from the record
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

func CheckExistenceOrErrorByStringId(provider DbProvider, record CrudRecord, id string) error {

	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid object id %s", id)
	}

	record.SetId(objId)

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
	return fmt.Errorf("%s with ID %s does not exist", record.RecordType(), record.GetId().Hex())
}
