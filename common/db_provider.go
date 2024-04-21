package common

import "strings"

var GlobalDbProvider DbProvider

type DbProvider interface {
	Connect() error
	Disconnect() error

	GetAll(record CrudRecord) (objects interface{}, err error)
	GetAllWhere(record CrudRecord, filter map[string]interface{}) (objects ListOfCrudRecords, err error)
	GetOne(record CrudRecord) (object CrudRecord, exists bool, err error)
	Create(object CrudRecord) (CrudRecord, error)
	Update(object CrudRecord) error
	Delete(record CrudRecord) error
}

func GetAll(db DbProvider, record CrudRecord) (objects interface{}, err error) {
	return db.GetAll(record)
}

func GetAllWhere(db DbProvider, record CrudRecord, filter map[string]interface{}) (objects ListOfCrudRecords, err error) {
	return db.GetAllWhere(record, filter)
}

func GetOne(db DbProvider, record CrudRecord) (object CrudRecord, exists bool, err error) {
	return db.GetOne(record)
}

func Create(db DbProvider, record CrudRecord) (object CrudRecord, err error) {

	err = record.ValidateStatic()
	if err != nil {
		if strings.Contains(err.Error(), "unmarshal") {
			panic(err)
		}
		return nil, err
	}

	err = record.ValidateDynamic(db, false, nil)
	if err != nil {
		if strings.Contains(err.Error(), "unmarshal") {
			panic(err)
		}
		return nil, err
	}

	return db.Create(record)
}

func Update(db DbProvider, record CrudRecord) (err error) {
	err = record.ValidateStatic()
	if err != nil {
		return err
	}

	err = record.ValidateDynamic(db, false, nil)
	if err != nil {
		return err
	}

	// check if the record in question has non-updatable fields
	// configured and whether any of those fields has been changed
	// in the request vs. the existing record
	h, ok := record.(HasNonUpdatable)
	if ok {
		err = CheckNonUpdatableFields(h, db)
		if err != nil {
			return err
		}
	}

	return db.Update(record)
}

func Delete(db DbProvider, record CrudRecord) (err error) {
	return db.Delete(record)
}
