package database

import (
	"context"
	"fmt"
	"sync"
)

type RecordCache map[RecordId]CrudRecord // Map from a RecordId to the CrudRecord with that ID

// UnitTestDbProvider is an in-memory, map-backed database.Provider. It is used
// by the test suite and as the ephemeral "memory" provider for the running
// server. All access is guarded by a mutex so concurrent gin handlers do not
// race on the underlying maps (a pre-#53 stopgap; SQLite replaces it).
type UnitTestDbProvider struct {
	Caches map[string]RecordCache // map from CrudRecord.Type to a RecordCache
	mu     sync.Mutex
}

func NewUnitTestDBProvider() *UnitTestDbProvider {
	return &UnitTestDbProvider{
		Caches: make(map[string]RecordCache),
	}
}

// getOrCreateRecordCache returns the RecordCache for the given record type,
// creating it if necessary. The caller must hold u.mu.
func (u *UnitTestDbProvider) getOrCreateRecordCache(recordType CrudRecord) RecordCache {
	v, ok := u.Caches[recordType.Type()]
	if ok {
		return v
	}
	u.Caches[recordType.Type()] = make(RecordCache)
	return u.Caches[recordType.Type()]
}

// getOne is the unlocked implementation of GetOne. The caller must hold u.mu.
func (u *UnitTestDbProvider) getOne(record CrudRecord) (CrudRecord, bool) {
	cache := u.getOrCreateRecordCache(record)
	v, ok := cache[record.GetId()]
	return v, ok
}

func (u *UnitTestDbProvider) GetOne(ctx context.Context, record CrudRecord) (CrudRecord, bool, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	v, ok := u.getOne(record)
	return v, ok, nil
}

func (u *UnitTestDbProvider) GetAll(ctx context.Context, recordType CrudRecord) ([]CrudRecord, error) {
	return u.GetAllWhere(ctx, recordType, nil)
}

func (u *UnitTestDbProvider) GetAllWhere(ctx context.Context, recordType CrudRecord, where WhereFunc) ([]CrudRecord, error) {
	// Snapshot the records under the lock, then evaluate the where predicate
	// outside it. The predicate may itself re-enter the provider — e.g. a
	// record's AccessibleTo/EditableBy resolving its members via another
	// GetAllWhere/GetOne (Team → TeamAssignment). Because u.mu is a plain,
	// non-reentrant Mutex, holding it while running the predicate would
	// deadlock on that nested read. The snapshot is stable, so the predicate
	// is safe to run unlocked.
	u.mu.Lock()
	cache := u.getOrCreateRecordCache(recordType)
	records := make([]CrudRecord, 0, len(cache))
	for _, record := range cache {
		records = append(records, record)
	}
	u.mu.Unlock()

	output := make([]CrudRecord, 0, len(records))
	for _, record := range records {
		if where == nil || where(ctx, record) {
			output = append(output, record)
		}
	}
	return output, nil
}

func (u *UnitTestDbProvider) Create(ctx context.Context, record CrudRecord) (CrudRecord, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	// get the RecordCache for this type, creating it if necessary
	cache := u.getOrCreateRecordCache(record)

	// set a RecordId if not already set
	if !record.GetId().ValidRecordId() {
		record.SetId(NewRecordId())
	}

	// check that a record doesn't already exist with the given RecordId
	_, exists := cache[record.GetId()]
	if exists {
		return nil, fmt.Errorf("A %s record with ID %s already exists", record.Type(), record.GetId())
	}

	// save the record to the RecordCache
	cache[record.GetId()] = record
	return record, nil
}

func (u *UnitTestDbProvider) Update(ctx context.Context, record CrudRecord) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	_, exists := u.getOne(record)
	if !exists {
		return fmt.Errorf("%s with ID  %s does not exist", record.Type(), record.GetId())
	}
	cache := u.getOrCreateRecordCache(record)
	cache[record.GetId()] = record
	return nil
}

func (u *UnitTestDbProvider) Delete(ctx context.Context, record CrudRecord) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	_, exists := u.getOne(record)
	if !exists {
		return nil
	}
	cache := u.getOrCreateRecordCache(record)
	delete(cache, record.GetId())
	return nil
}

func (u *UnitTestDbProvider) Dump() {
	u.mu.Lock()
	defer u.mu.Unlock()
	for tableName, cache := range u.Caches {
		for id, record := range cache {
			fmt.Printf("%s %s\n", tableName, id)
			fmt.Printf("  ---> %+v\n", record)
		}
	}
}
