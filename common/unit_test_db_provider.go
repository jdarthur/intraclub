package common

import "fmt"

type RecordCache map[RecordId]CrudRecord // Map from a RecordId to the CrudRecord with that ID
type UnitTestDBProvider struct {
	Caches map[string]RecordCache // map from CrudRecord.Type to a RecordCache
}

func NewUnitTestDBProvider() *UnitTestDBProvider {
	return &UnitTestDBProvider{
		Caches: make(map[string]RecordCache),
	}
}

func (u *UnitTestDBProvider) getOrCreateRecordCache(recordType CrudRecord) RecordCache {
	v, ok := u.Caches[recordType.Type()]
	if ok {
		return v
	}
	u.Caches[recordType.Type()] = make(RecordCache)
	return u.Caches[recordType.Type()]
}

func (u *UnitTestDBProvider) GetOne(record CrudRecord) (CrudRecord, bool, error) {
	cache := u.getOrCreateRecordCache(record)
	v, ok := cache[record.GetId()]
	return v, ok, nil
}

func (u *UnitTestDBProvider) GetAll(recordType CrudRecord) ([]CrudRecord, error) {
	return u.GetAllWhere(recordType, nil)
}

func (u *UnitTestDBProvider) GetAllWhere(recordType CrudRecord, where WhereFunc) ([]CrudRecord, error) {

	output := make([]CrudRecord, 0)
	cache := u.getOrCreateRecordCache(recordType)
	for _, record := range cache {
		if where == nil || where(record) {
			output = append(output, record)
		}
	}
	//u.Dump()
	return output, nil
}

func (u *UnitTestDBProvider) Create(record CrudRecord) (CrudRecord, error) {
	// get the RecordCache for this type, creating it if necessary
	cache := u.getOrCreateRecordCache(record)

	// set a RecordId if not already set
	if !record.GetId().ValidRecordId() {
		record.SetId(NewRecordId())
	}

	// check that a record doesn't already exist with the given RecordIs
	_, exists := cache[record.GetId()]
	if exists {
		return nil, fmt.Errorf("A %s record with ID %s already exists", record.Type(), record.GetId())
	}

	// save the record to the RecordCache
	cache[record.GetId()] = record
	return record, nil
}

func (u *UnitTestDBProvider) Update(record CrudRecord) error {
	_, exists, _ := u.GetOne(record)
	if !exists {
		return fmt.Errorf("%s with ID  %s does not exist", record.Type(), record.GetId())
	}
	cache := u.getOrCreateRecordCache(record)
	cache[record.GetId()] = record
	return nil
}

func (u *UnitTestDBProvider) Delete(record CrudRecord) error {
	_, exists, _ := u.GetOne(record)
	if !exists {
		return nil
	}
	cache := u.getOrCreateRecordCache(record)
	delete(cache, record.GetId())
	return nil
}

func (u *UnitTestDBProvider) Dump() {
	for tableName, cache := range u.Caches {
		for id, record := range cache {
			fmt.Printf("%s %s\n", tableName, id)
			fmt.Printf("  ---> %+v\n", record)
		}
	}
}
