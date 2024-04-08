package test

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"intraclub/common"
)

// RecordCache is nested map of records from
// a common.CrudRecord record type to a map from
// object ID to object, e.g.
//
//		{
//			user:
//			{
//				b8c239bc-0c02-4b09-a5f0-89a8ca057bd8:
//				{
//					...
//				}
//			},
//		    player:
//			{
//				202697e3-dd10-4c46-9392-8975ccdc0bb4:
//				{
//					...
//				}
//			},
//	     ...
//		}
type RecordCache map[string]map[string]common.CrudRecord

type UnitTestDbProvider struct {
	Map RecordCache
}

type listOfAny []interface{}

func (l listOfAny) Length() int {
	return len(l)
}

func (u UnitTestDbProvider) GetAllWhere(record common.CrudRecord, filter map[string]interface{}) (objects common.ListOfCrudRecords, err error) {

	rootKey := record.RecordType()

	output := make(listOfAny, 0)

	v, ok := u.Map[rootKey]
	if !ok {
		return output, nil
	}

	for _, record := range v {

		m := make(map[string]interface{})
		b, err := json.Marshal(record)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(b, &m)
		if err != nil {
			return nil, err
		}

		for key, value := range filter {
			mValue, ok := m[key]
			if ok && mValue == value {
				output = append(output, m)
			}
		}
	}

	return output, nil
}

func (u UnitTestDbProvider) Connect() error {
	return nil
}

func (u UnitTestDbProvider) Disconnect() error {
	return nil
}

func (u UnitTestDbProvider) GetAll(record common.CrudRecord) (objects interface{}, err error) {
	rootKey := record.RecordType()

	output := make([]interface{}, 0)

	v, ok := u.Map[rootKey]
	if !ok {
		return output, nil
	}

	for _, value := range v {
		output = append(output, value)
	}

	return output, nil
}

func (u UnitTestDbProvider) GetOne(record common.CrudRecord) (object common.CrudRecord, exists bool, err error) {
	v := u.CreateRootMapIfNotExists(record)

	object, ok := v[record.GetId()]
	if !ok {
		return nil, false, nil
	}

	return object, true, nil
}

func (u UnitTestDbProvider) Create(object common.CrudRecord) (common.CrudRecord, error) {
	v := u.CreateRootMapIfNotExists(object)

	id := primitive.NewObjectID().String()
	object.SetId(id)

	v[id] = object
	u.Map[object.RecordType()] = v

	return object, nil
}

func (u UnitTestDbProvider) Update(object common.CrudRecord) error {
	v := u.CreateRootMapIfNotExists(object)

	id := primitive.NewObjectID().String()
	object.SetId(id)

	v[id] = object
	u.Map[object.RecordType()] = v

	return nil
}

func (u UnitTestDbProvider) Delete(record common.CrudRecord) error {
	v := u.CreateRootMapIfNotExists(record)

	_, ok := v[record.GetId()]
	if ok {
		delete(v, record.GetId())
	}

	u.Map[record.RecordType()] = v
	return nil
}

func NewUnitTestDbProvider() *UnitTestDbProvider {
	return &UnitTestDbProvider{Map: make(RecordCache)}
}

func (u UnitTestDbProvider) CreateRootMapIfNotExists(record common.CrudRecord) map[string]common.CrudRecord {
	v, ok := u.Map[record.RecordType()]
	if !ok {
		v = make(map[string]common.CrudRecord)
	}

	return v
}

func InitializeTestDbProvider() {
	p := NewUnitTestDbProvider()
	common.GlobalDbProvider = p
}
