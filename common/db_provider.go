package common

var GlobalDbProvider DbProvider

type DbProvider interface {
	Connect() error
	Disconnect() error

	GetAll(record CrudRecord) (objects interface{}, err error)
	GetOne(record CrudRecord, id string) (object interface{}, exists bool, err error)
	Create(record CrudRecord, object interface{}) (interface{}, error)
	Update(record CrudRecord, object interface{}) error
	Delete(record CrudRecord, id string) error
}
