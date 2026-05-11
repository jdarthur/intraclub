package database

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type testRecord struct {
	ID        RecordId
	Owner     RecordId
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (t *testRecord) GetOwner() RecordId {
	return t.Owner
}

func newTestRecord() *testRecord {
	return &testRecord{
		ID: NewRecordId(),
	}
}

func (t *testRecord) Copy() *testRecord {
	return &testRecord{
		ID:        t.ID,
		Owner:     t.Owner,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func (t *testRecord) Type() string {
	return "test record"
}

func (t *testRecord) GetId() RecordId {
	return t.ID
}

func (t *testRecord) SetId(id RecordId) {
	t.ID = id
}

func (t *testRecord) EditableBy(_ context.Context, db DatabaseProvider) []RecordId {
	return []RecordId{t.Owner}
}

func (t *testRecord) AccessibleTo(_ context.Context, db DatabaseProvider) []RecordId {
	return AccessibleToEveryone
}

func (t *testRecord) SetOwner(recordId RecordId) {
	t.Owner = recordId
}

func (t *testRecord) StaticallyValid() error {
	return nil
}

func (t *testRecord) DynamicallyValid(_ context.Context, db DatabaseProvider) error {
	return nil
}

func (t *testRecord) GetTimeStamps() (created, updated time.Time) {
	return t.CreatedAt, t.UpdatedAt
}

func (t *testRecord) SetCreateTimestamp(time time.Time) time.Time {
	oldValue := t.CreatedAt
	t.CreatedAt = time
	return oldValue
}

func (t *testRecord) SetUpdateTimestamp(time time.Time) time.Time {
	oldValue := t.UpdatedAt
	t.UpdatedAt = time
	return oldValue
}

func (t *testRecord) BlankRecord() CrudRecord {
	return new(testRecord)
}

func TestCreateDataIsSetOnCreate(t *testing.T) {
	db := NewUnitTestDBProvider()
	v := newTestRecord()
	created, err := CreateOne(context.Background(), db, v)
	if err != nil {
		t.Fatal(err)
	}

	if created.CreatedAt.IsZero() {
		t.Error("Created timestamp is zero")
	}
	fmt.Println(created)
}

func TestCreateDateIsImmutable(t *testing.T) {
	db := NewUnitTestDBProvider()
	v := newTestRecord()
	created, err := CreateOne(context.Background(), db, v)
	if err != nil {
		t.Fatal(err)
	}

	copied := created.Copy()
	copied.CreatedAt = time.Now()
	err = UpdateOne(context.Background(), db, copied)
	if err != nil {
		t.Fatal(err)
	}
	if copied.CreatedAt != created.CreatedAt {
		t.Error("Created timestamp is not immutable")
	}
}

func TestGetOneByIdSuccess(t *testing.T) {
	db := NewUnitTestDBProvider()
	v := newTestRecord()
	created, err := CreateOne(context.Background(), db, v)
	if err != nil {
		t.Fatal(err)
	}

	result, exists, err := GetOneById(context.Background(), db, &testRecord{}, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("record should exist")
	}
	if result.ID != created.ID {
		t.Fatalf("expected ID %s, got %s", created.ID, result.ID)
	}
}

func TestGetOneByIdNotFound(t *testing.T) {
	db := NewUnitTestDBProvider()
	nonExistentId := NewRecordId()

	result, exists, err := GetOneById(context.Background(), db, &testRecord{}, nonExistentId)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("record should not exist")
	}
	if result != nil {
		t.Fatal("result should be nil for non-existent record")
	}
}

func TestGetExistingRecordByIdSuccess(t *testing.T) {
	db := NewUnitTestDBProvider()
	v := newTestRecord()
	created, err := CreateOne(context.Background(), db, v)
	if err != nil {
		t.Fatal(err)
	}

	result, err := GetExistingRecordById(context.Background(), db, &testRecord{}, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != created.ID {
		t.Fatalf("expected ID %s, got %s", created.ID, result.ID)
	}
}

func TestGetExistingRecordByIdNotFound(t *testing.T) {
	db := NewUnitTestDBProvider()
	nonExistentId := NewRecordId()

	_, err := GetExistingRecordById(context.Background(), db, &testRecord{}, nonExistentId)
	if err == nil {
		t.Fatal("expected error for non-existent record")
	}
}

func TestExistsByIdExists(t *testing.T) {
	db := NewUnitTestDBProvider()
	v := newTestRecord()
	created, err := CreateOne(context.Background(), db, v)
	if err != nil {
		t.Fatal(err)
	}

	err = ExistsById(context.Background(), db, &testRecord{}, created.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestExistsByIdNotFound(t *testing.T) {
	db := NewUnitTestDBProvider()
	nonExistentId := NewRecordId()

	err := ExistsById(context.Background(), db, &testRecord{}, nonExistentId)
	if err == nil {
		t.Fatal("expected error for non-existent record")
	}
}

func TestGetAllEmpty(t *testing.T) {
	db := NewUnitTestDBProvider()

	results, err := GetAll[*testRecord](context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 records, got %d", len(results))
	}
}

func TestGetAllWithRecords(t *testing.T) {
	db := NewUnitTestDBProvider()

	for i := 0; i < 5; i++ {
		_, err := CreateOne(context.Background(), db, newTestRecord())
		if err != nil {
			t.Fatal(err)
		}
	}

	results, err := GetAll[*testRecord](context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 5 {
		t.Fatalf("expected 5 records, got %d", len(results))
	}
}

func TestGetAllWhereWithFilter(t *testing.T) {
	db := NewUnitTestDBProvider()

	ownerId1 := NewRecordId()
	ownerId2 := NewRecordId()

	v1 := newTestRecord()
	v1.Owner = ownerId1
	_, err := CreateOne(context.Background(), db, v1)
	if err != nil {
		t.Fatal(err)
	}

	v2 := newTestRecord()
	v2.Owner = ownerId2
	_, err = CreateOne(context.Background(), db, v2)
	if err != nil {
		t.Fatal(err)
	}

	results, err := GetAllWhere[*testRecord](context.Background(), db, func(_ context.Context, c *testRecord) bool {
		return c.Owner == ownerId1
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 record, got %d", len(results))
	}
	if results[0].Owner != ownerId1 {
		t.Fatal("expected record with ownerId1")
	}
}

func TestGetAllWhereNoResults(t *testing.T) {
	db := NewUnitTestDBProvider()

	_, err := CreateOne(context.Background(), db, newTestRecord())
	if err != nil {
		t.Fatal(err)
	}

	results, err := GetAllWhere[*testRecord](context.Background(), db, func(_ context.Context, c *testRecord) bool {
		return false
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 records, got %d", len(results))
	}
}

func TestGetAllWhereNilFilter(t *testing.T) {
	db := NewUnitTestDBProvider()

	for i := 0; i < 3; i++ {
		_, err := CreateOne(context.Background(), db, newTestRecord())
		if err != nil {
			t.Fatal(err)
		}
	}

	results, err := GetAllWhere[*testRecord](context.Background(), db, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 records, got %d", len(results))
	}
}

func TestDeleteOneByIdSuccess(t *testing.T) {
	db := NewUnitTestDBProvider()
	v := newTestRecord()
	created, err := CreateOne(context.Background(), db, v)
	if err != nil {
		t.Fatal(err)
	}

	result, exists, err := DeleteOneById(context.Background(), db, &testRecord{}, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("record should exist")
	}
	if result.ID != created.ID {
		t.Fatalf("expected ID %s, got %s", created.ID, result.ID)
	}

	// Verify it's actually deleted
	_, exists2, _ := GetOneById(context.Background(), db, &testRecord{}, created.ID)
	if exists2 {
		t.Fatal("record should be deleted")
	}
}

func TestDeleteOneByIdNotFound(t *testing.T) {
	db := NewUnitTestDBProvider()
	nonExistentId := NewRecordId()

	result, exists, err := DeleteOneById(context.Background(), db, &testRecord{}, nonExistentId)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("record should not exist")
	}
	if result != nil {
		t.Fatal("result should be nil for non-existent record")
	}
}

func TestDeleteOneByIdDoubleDelete(t *testing.T) {
	db := NewUnitTestDBProvider()
	v := newTestRecord()
	created, err := CreateOne(context.Background(), db, v)
	if err != nil {
		t.Fatal(err)
	}

	// First delete
	_, exists1, err := DeleteOneById(context.Background(), db, &testRecord{}, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !exists1 {
		t.Fatal("first delete should find record")
	}

	// Second delete (should be idempotent)
	_, exists2, err := DeleteOneById(context.Background(), db, &testRecord{}, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if exists2 {
		t.Fatal("second delete should not find record")
	}
}

func TestUpdateOneBasic(t *testing.T) {
	db := NewUnitTestDBProvider()
	v := newTestRecord()
	v.Owner = NewRecordId()
	created, err := CreateOne(context.Background(), db, v)
	if err != nil {
		t.Fatal(err)
	}

	updated := &testRecord{
		ID:    created.ID,
		Owner: NewRecordId(),
	}
	err = UpdateOne(context.Background(), db, updated)
	if err != nil {
		t.Fatal(err)
	}

	retrieved, exists, err := GetOneById(context.Background(), db, &testRecord{}, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("record should exist after update")
	}
	if retrieved.Owner != updated.Owner {
		t.Fatal("Owner should be updated")
	}
}

func TestUpdateOneNotFound(t *testing.T) {
	db := NewUnitTestDBProvider()

	nonExistent := &testRecord{
		ID:    NewRecordId(),
		Owner: NewRecordId(),
	}
	err := UpdateOne(context.Background(), db, nonExistent)
	if err == nil {
		t.Fatal("expected error updating non-existent record")
	}
}

func TestCreateOneWithPreSetId(t *testing.T) {
	db := NewUnitTestDBProvider()
	v := newTestRecord()
	v.ID = NewRecordId()
	created, err := CreateOne(context.Background(), db, v)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != v.ID {
		t.Fatalf("expected ID %s, got %s", v.ID, created.ID)
	}
}

func TestCreateOneGeneratesNewId(t *testing.T) {
	db := NewUnitTestDBProvider()
	id := NewRecordId()

	v := newTestRecord()
	v.ID = id
	created, err := CreateOne(context.Background(), db, v)
	if err != nil {
		t.Fatal(err)
	}
	// CreateOne always generates a new ID, even if one was already set
	if created.ID == id {
		t.Fatal("CreateOne should generate a new ID even if one was already set")
	}
}

func TestGetAllDifferentTypes(t *testing.T) {
	db := NewUnitTestDBProvider()

	_, err := CreateOne(context.Background(), db, newTestRecord())
	if err != nil {
		t.Fatal(err)
	}

	testUnique1 := &testUnique{
		RecordId:     1,
		ReferenceId1: 2,
		ReferenceId2: 3,
	}
	_, err = CreateOne(context.Background(), db, testUnique1)
	if err != nil {
		t.Fatal(err)
	}

	testRecords, err := GetAll[*testRecord](context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if len(testRecords) != 1 {
		t.Fatalf("expected 1 testRecord, got %d", len(testRecords))
	}

	uniqueRecords, err := GetAll[*testUnique](context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if len(uniqueRecords) != 1 {
		t.Fatalf("expected 1 testUnique, got %d", len(uniqueRecords))
	}
}

func TestCrudRecordBlankRecord(t *testing.T) {
	v := newTestRecord()
	blank := v.BlankRecord()
	if blank == nil {
		t.Fatal("BlankRecord should not return nil")
	}
	blankTestRecord, ok := blank.(*testRecord)
	if !ok {
		t.Fatal("BlankRecord should return *testRecord")
	}
	if blankTestRecord.ID != InvalidRecordId {
		t.Fatal("BlankRecord should have invalid ID")
	}
}
