package database

import (
	"context"
	"fmt"
	"testing"
)

type testUnique struct {
	RecordId     RecordId
	ReferenceId1 RecordId
	ReferenceId2 RecordId
}

func (t *testUnique) GetOwner() UserId {
	return InvalidUserId
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

func (t *testUnique) EditableBy(_ context.Context, db DatabaseProvider) []UserId {
	return nil
}

func (t *testUnique) AccessibleTo(_ context.Context, db DatabaseProvider) []UserId {
	return nil
}

func (t *testUnique) SetOwner(userId UserId) {}

func (t *testUnique) StaticallyValid() error {
	return nil
}

func (t *testUnique) DynamicallyValid(_ context.Context, db DatabaseProvider) error {
	return nil
}

func (t *testUnique) BlankRecord() CrudRecord {
	return new(testUnique)
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

	_, err := CreateOne(context.Background(), db, &record1)
	if err != nil {
		t.Fatal(err)
	}
	_, err = CreateOne(context.Background(), db, &record2)
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

	_, err := CreateOne(context.Background(), db, &record1)
	if err != nil {
		t.Fatal(err)
	}
	v2, err := CreateOne(context.Background(), db, &record2)
	if err != nil {
		t.Fatal(err)
	}

	update := testUnique{
		RecordId:     v2.RecordId,
		ReferenceId1: 2,
		ReferenceId2: 3,
	}
	err = UpdateOne(context.Background(), db, &update)
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
	v, err := CreateOne(context.Background(), db, &record1)
	if err != nil {
		t.Fatal(err)
	}

	update := testUnique{
		RecordId:     v.RecordId,
		ReferenceId1: 3,
		ReferenceId2: 4,
	}
	err = UpdateOne(context.Background(), db, &update)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUniqueConstraintNoExistingRecords(t *testing.T) {
	db := NewUnitTestDBProvider()

	record1 := testUnique{
		RecordId:     1,
		ReferenceId1: 2,
		ReferenceId2: 3,
	}
	_, err := CreateOne(context.Background(), db, &record1)
	if err != nil {
		t.Fatal("should succeed when no existing records")
	}
}

func TestUniqueConstraintSameRecordUpdate(t *testing.T) {
	db := NewUnitTestDBProvider()
	record1 := testUnique{
		RecordId:     1,
		ReferenceId1: 2,
		ReferenceId2: 3,
	}
	v, err := CreateOne(context.Background(), db, &record1)
	if err != nil {
		t.Fatal(err)
	}

	// Update with same values (should succeed since it's the same record)
	update := testUnique{
		RecordId:     v.RecordId,
		ReferenceId1: 2,
		ReferenceId2: 3,
	}
	err = UpdateOne(context.Background(), db, &update)
	if err != nil {
		t.Fatal("updating to same values should succeed")
	}
}

func TestUniqueConstraintPartialMatch(t *testing.T) {
	db := NewUnitTestDBProvider()

	record1 := testUnique{
		RecordId:     1,
		ReferenceId1: 2,
		ReferenceId2: 3,
	}
	_, err := CreateOne(context.Background(), db, &record1)
	if err != nil {
		t.Fatal(err)
	}

	// Only ReferenceId1 matches - should succeed
	record2 := testUnique{
		RecordId:     2,
		ReferenceId1: 2,
		ReferenceId2: 99,
	}
	_, err = CreateOne(context.Background(), db, &record2)
	if err != nil {
		t.Fatal("partial match should succeed")
	}

	// Only ReferenceId2 matches - should succeed
	record3 := testUnique{
		RecordId:     3,
		ReferenceId1: 99,
		ReferenceId2: 3,
	}
	_, err = CreateOne(context.Background(), db, &record3)
	if err != nil {
		t.Fatal("partial match should succeed")
	}
}

func TestUniqueConstraintMultipleDuplicates(t *testing.T) {
	db := NewUnitTestDBProvider()

	record1 := testUnique{
		RecordId:     1,
		ReferenceId1: 100,
		ReferenceId2: 200,
	}
	_, err := CreateOne(context.Background(), db, &record1)
	if err != nil {
		t.Fatal(err)
	}

	// First duplicate attempt
	record2 := testUnique{
		RecordId:     2,
		ReferenceId1: 100,
		ReferenceId2: 200,
	}
	_, err = CreateOne(context.Background(), db, &record2)
	if err == nil {
		t.Fatal("first duplicate should fail")
	}

	// Second duplicate attempt with different ID
	record3 := testUnique{
		RecordId:     3,
		ReferenceId1: 100,
		ReferenceId2: 200,
	}
	_, err = CreateOne(context.Background(), db, &record3)
	if err == nil {
		t.Fatal("second duplicate should fail")
	}
}

func TestUniqueConstraintUpdateToSelfUnique(t *testing.T) {
	db := NewUnitTestDBProvider()

	record1 := testUnique{
		RecordId:     1,
		ReferenceId1: 100,
		ReferenceId2: 200,
	}
	v1, err := CreateOne(context.Background(), db, &record1)
	if err != nil {
		t.Fatal(err)
	}

	record2 := testUnique{
		RecordId:     2,
		ReferenceId1: 300,
		ReferenceId2: 400,
	}
	v2, err := CreateOne(context.Background(), db, &record2)
	if err != nil {
		t.Fatal(err)
	}

	// Update record2 to match record1's values (should fail)
	update := testUnique{
		RecordId:     v2.RecordId,
		ReferenceId1: v1.ReferenceId1,
		ReferenceId2: v1.ReferenceId2,
	}
	err = UpdateOne(context.Background(), db, &update)
	if err == nil {
		t.Fatal("updating to match another record's unique values should fail")
	}
}
