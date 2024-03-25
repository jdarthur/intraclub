package common

import "fmt"

type CrudRecord interface {
	RecordType() string
	OneRecord() CrudRecord
	ListOfRecords() interface{}
	SetId(id string)
	GetId() string
	Validatable
}

func CheckExistenceOrError(record CrudRecord) error {
	_, exists, err := GetOne(GlobalDbProvider, record)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("%s with ID %s does not exist", record.RecordType(), record.GetId())
	}

	return nil
}
