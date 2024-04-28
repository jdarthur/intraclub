package common

var GlobalDbProvider DbProvider

type DbProvider interface {
	Connect() error
	Disconnect() error

	GetAll(record CrudRecord) (objects ListOfCrudRecords, err error)
	GetAllWhere(record CrudRecord, filter map[string]interface{}) (objects ListOfCrudRecords, err error)
	GetOne(record CrudRecord) (object CrudRecord, exists bool, err error)
	Create(object CrudRecord) (CrudRecord, error)
	Update(object CrudRecord) error
	Delete(record CrudRecord) error
}

func GetAll(db DbProvider, record CrudRecord) (objects ListOfCrudRecords, err error) {
	return db.GetAll(record)

}

func GetAllWhere(db DbProvider, record CrudRecord, filter map[string]interface{}) (objects ListOfCrudRecords, err error) {
	return db.GetAllWhere(record, filter)
}

func GetOne(db DbProvider, record CrudRecord) (object CrudRecord, exists bool, err error) {
	return db.GetOne(record)
}

func GetOneByStringId(db DbProvider, record CrudRecord, id string) (object CrudRecord, err error) {

	objectId, err := TryParsingObjectId(id)
	if err != nil {
		return nil, err
	}

	// set the object ID on this record
	record.SetId(objectId)

	record, exists, err := GetOne(db, record)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, RecordDoesNotExist(record)
	}

	return record, nil
}

func Create(db DbProvider, record CrudRecord) (object CrudRecord, err error) {
	err = record.ValidateStatic()
	if err != nil {
		return nil, err
	}

	err = record.ValidateDynamic(db, false, nil)
	if err != nil {
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
