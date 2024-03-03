package common

type CrudRecord interface {
	RecordType() string
	OneRecord() CrudRecord
	ListOfRecords() interface{}
	SetId(id string)
	GetId() string
	Validatable
}
