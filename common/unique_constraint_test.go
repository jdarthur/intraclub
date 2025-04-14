package common

import (
	"fmt"
	"testing"
)

type testUnique struct {
	RecordId     RecordId
	ReferenceId1 RecordId
	ReferenceId2 RecordId
}

func (t *testUnique) GetOwner() RecordId {
	return InvalidRecordId
}

func (t *testUnique) UniquenessEquivalent(other *testUnique) error {
	if t.ReferenceId1 == other.ReferenceId1 && t.ReferenceId2 == other.ReferenceId2 {
		return fmt.Errorf("duplicate reference value pair")
	}
	return nil
}

func (t *testUnique) Type() string {
	return "test_unique"
}

func (t *testUnique) GetId() RecordId {
	return t.RecordId
}

func (t *testUnique) SetId(id RecordId) {
	t.RecordId = id
}

func (t *testUnique) EditableBy(db DatabaseProvider) []RecordId {
	return nil
}

func (t *testUnique) AccessibleTo(db DatabaseProvider) []RecordId {
	return nil
}

func (t *testUnique) SetOwner(recordId RecordId) {}

func (t *testUnique) StaticallyValid() error {
	return nil
}

func (t *testUnique) DynamicallyValid(db DatabaseProvider) error {
	return nil
}

func TestValidateUniqueConstraintOnCreate(t *testing.T) {
	db := NewUnitTestDBProvider()
	record1 := testUnique{
		RecordId:     1,
		ReferenceId1: 2,
		ReferenceId2: 3,
	}
	record2 := testUnique{
		RecordId:     2,
		ReferenceId1: 2,
		ReferenceId2: 3,
	}

	_, err := CreateOne(db, &record1)
	if err != nil {
		t.Fatal(err)
	}
	_, err = CreateOne(db, &record2)
	if err == nil {
		t.Fatal("Expected error when creating record which violates unique constraint")
	}
	fmt.Println(err)
}

func TestValidateUniqueConstraintOnUpdate(t *testing.T) {
	db := NewUnitTestDBProvider()
	record1 := testUnique{
		RecordId:     1,
		ReferenceId1: 2,
		ReferenceId2: 3,
	}
	record2 := testUnique{
		RecordId:     2,
		ReferenceId1: 3,
		ReferenceId2: 4,
	}

	_, err := CreateOne(db, &record1)
	if err != nil {
		t.Fatal(err)
	}
	v2, err := CreateOne(db, &record2)
	if err != nil {
		t.Fatal(err)
	}

	update := testUnique{
		RecordId:     v2.RecordId,
		ReferenceId1: 2,
		ReferenceId2: 3,
	}
	err = UpdateOne(db, &update)
	if err == nil {
		t.Fatal("Expected error when updating record which violates unique constraint")
	}

	fmt.Println(err)
}

func TestSelfUpdateWithUniqueConstraint(t *testing.T) {
	db := NewUnitTestDBProvider()
	record1 := testUnique{
		RecordId:     1,
		ReferenceId1: 2,
		ReferenceId2: 3,
	}
	v, err := CreateOne(db, &record1)
	if err != nil {
		t.Fatal(err)
	}

	update := testUnique{
		RecordId:     v.RecordId,
		ReferenceId1: 3,
		ReferenceId2: 4,
	}
	err = UpdateOne(db, &update)
	if err != nil {
		t.Fatal(err)
	}
}
