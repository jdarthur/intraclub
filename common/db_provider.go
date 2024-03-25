package common

var GlobalDbProvider DbProvider

type DbProvider interface {
	Connect() error
	Disconnect() error

	GetAll(record CrudRecord) (objects interface{}, err error)
	GetOne(record CrudRecord) (object CrudRecord, exists bool, err error)
	Create(object CrudRecord) (CrudRecord, error)
	Update(object CrudRecord) error
	Delete(record CrudRecord) error
}

func GetAll(db DbProvider, record CrudRecord) (objects interface{}, err error) {
	return db.GetAll(record)
}

func GetOne(db DbProvider, record CrudRecord) (object CrudRecord, exists bool, err error) {
	return db.GetOne(record)
}

func Create(db DbProvider, record CrudRecord) (object CrudRecord, err error) {

	err = record.ValidateStatic()
	if err != nil {
		return nil, err
	}

	err = record.ValidateDynamic(db)
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

	err = record.ValidateDynamic(db)
	if err != nil {
		return err
	}

	return db.Update(record)
}

func Delete(db DbProvider, record CrudRecord) (err error) {
	return db.Delete(record)
}
