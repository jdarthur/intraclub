package common

import (
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

func (t *testRecord) EditableBy(db DatabaseProvider) []RecordId {
	return []RecordId{t.Owner}
}

func (t *testRecord) AccessibleTo(db DatabaseProvider) []RecordId {
	return AccessibleToEveryone
}

func (t *testRecord) SetOwner(recordId RecordId) {
	t.Owner = recordId
}

func (t *testRecord) StaticallyValid() error {
	return nil
}

func (t *testRecord) DynamicallyValid(db DatabaseProvider) error {
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

func TestCreateDataIsSetOnCreate(t *testing.T) {
	db := NewUnitTestDBProvider()
	v := newTestRecord()
	created, err := CreateOne(db, v)
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
	created, err := CreateOne(db, v)
	if err != nil {
		t.Fatal(err)
	}

	copied := created.Copy()
	copied.CreatedAt = time.Now()
	err = UpdateOne(db, copied)
	if err != nil {
		t.Fatal(err)
	}
	if copied.CreatedAt != created.CreatedAt {
		t.Error("Created timestamp is not immutable")
	}
}
