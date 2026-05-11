package database

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestNewRecordIdGeneratesUniqueIds(t *testing.T) {
	ids := make(map[RecordId]bool)
	for i := 0; i < 1000; i++ {
		id := NewRecordId()
		if ids[id] {
			t.Fatalf("Duplicate RecordId generated: %d", id)
		}
		ids[id] = true
	}
}

func TestNewRecordIdNeverGeneratesUnavailableIds(t *testing.T) {
	for i := 0; i < 1000; i++ {
		id := NewRecordId()
		if id == InvalidRecordId {
			t.Fatal("NewRecordId generated InvalidRecordId")
		}
		if id == EveryoneRecordId {
			t.Fatal("NewRecordId generated EveryoneRecordId")
		}
		if id == SysAdminRecordId {
			t.Fatal("NewRecordId generated SysAdminRecordId")
		}
	}
}

func TestRecordIdValidRecordId(t *testing.T) {
	if InvalidRecordId.ValidRecordId() {
		t.Fatal("InvalidRecordId should not be valid")
	}
	if EveryoneRecordId.ValidRecordId() {
		t.Fatal("EveryoneRecordId should not be valid")
	}
	if SysAdminRecordId.ValidRecordId() {
		t.Fatal("SysAdminRecordId should not be valid")
	}
	id := NewRecordId()
	if !id.ValidRecordId() {
		t.Fatal("NewRecordId should be valid")
	}
}

func TestRecordIdString(t *testing.T) {
	id := RecordId(1)
	str := id.String()
	if str != "0000000000000001" {
		t.Fatalf("Expected '0000000000000001', got '%s'", str)
	}

	invalidId := InvalidRecordId
	invalidStr := invalidId.String()
	if invalidStr != "" {
		t.Fatalf("Expected empty string for InvalidRecordId, got '%s'", invalidStr)
	}

	id2 := RecordId(256)
	str2 := id2.String()
	if str2 != "0000000000000100" {
		t.Fatalf("Expected '0000000000000100', got '%s'", str2)
	}
}

func TestRecordIdFromString(t *testing.T) {
	id, err := RecordIdFromString("0000000000000001")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if id != RecordId(1) {
		t.Fatalf("Expected RecordId(1), got %d", id)
	}

	emptyId, err := RecordIdFromString("")
	if err != nil {
		t.Fatalf("Unexpected error for empty string: %v", err)
	}
	if emptyId != InvalidRecordId {
		t.Fatalf("Expected InvalidRecordId for empty string, got %d", emptyId)
	}

	_, err = RecordIdFromString("xyz")
	if err == nil {
		t.Fatal("Expected error for invalid hex string")
	}

	_, err = RecordIdFromString("000000000000000")
	if err == nil {
		t.Fatal("Expected error for short hex string")
	}

	_, err = RecordIdFromString("00000000000000000")
	if err == nil {
		t.Fatal("Expected error for long hex string")
	}
}

func TestRecordIdMarshalJSON(t *testing.T) {
	type testStruct struct {
		ID RecordId `json:"id"`
	}

	id := RecordId(256)
	data, err := json.Marshal(testStruct{ID: id})
	if err != nil {
		t.Fatalf("Unexpected error marshaling: %v", err)
	}
	expected := `{"id":"0000000000000100"}`
	if string(data) != expected {
		t.Fatalf("Expected '%s', got '%s'", expected, string(data))
	}
}

func TestRecordIdUnmarshalJSON(t *testing.T) {
	type testStruct struct {
		ID RecordId `json:"id"`
	}

	jsonStr := `{"id":"000000000000000a"}`
	var s testStruct
	err := json.Unmarshal([]byte(jsonStr), &s)
	if err != nil {
		t.Fatalf("Unexpected error unmarshaling: %v", err)
	}
	if s.ID != RecordId(10) {
		t.Fatalf("Expected RecordId(10), got %d", s.ID)
	}
}

func TestUnmarshalStringIdList(t *testing.T) {
	jsonStr := `["0000000000000001","0000000000000002","0000000000000003"]`
	ids, err := UnmarshalStringIdList([]byte(jsonStr))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(ids) != 3 {
		t.Fatalf("Expected 3 IDs, got %d", len(ids))
	}
	if ids[0] != RecordId(1) {
		t.Fatalf("Expected RecordId(1), got %d", ids[0])
	}
	if ids[1] != RecordId(2) {
		t.Fatalf("Expected RecordId(2), got %d", ids[1])
	}
	if ids[2] != RecordId(3) {
		t.Fatalf("Expected RecordId(3), got %d", ids[2])
	}
}

func TestUnmarshalStringIdListInvalid(t *testing.T) {
	_, err := UnmarshalStringIdList([]byte(`["invalid"]`))
	if err == nil {
		t.Fatal("Expected error for invalid ID string")
	}

	_, err = UnmarshalStringIdList([]byte(`not json`))
	if err == nil {
		t.Fatal("Expected error for invalid JSON")
	}
}

func TestSysAdminAndUsers(t *testing.T) {
	userId1 := RecordId(100)
	userId2 := RecordId(200)
	result := SysAdminAndUsers(userId1, userId2)
	if len(result) != 3 {
		t.Fatalf("Expected 3 IDs, got %d", len(result))
	}
	if result[0] != SysAdminRecordId {
		t.Fatalf("Expected SysAdminRecordId first, got %d", result[0])
	}
	if result[1] != userId1 {
		t.Fatalf("Expected userId1 second, got %d", result[1])
	}
	if result[2] != userId2 {
		t.Fatalf("Expected userId2 third, got %d", result[2])
	}
}

func TestSysAdminAndUsersEmpty(t *testing.T) {
	result := SysAdminAndUsers()
	if len(result) != 1 {
		t.Fatalf("Expected 1 ID (sys admin), got %d", len(result))
	}
	if result[0] != SysAdminRecordId {
		t.Fatalf("Expected SysAdminRecordId, got %d", result[0])
	}
}

func TestTimestampedRecordCreateTimestamp(t *testing.T) {
	db := NewUnitTestDBProvider()
	v := newTestRecord()
	created, err := CreateOne(context.Background(), db, v)
	if err != nil {
		t.Fatal(err)
	}
	if created.CreatedAt.IsZero() {
		t.Fatal("CreatedAt should be set after CreateOne")
	}
}

func TestTimestampedRecordUpdateTimestampOnUpdate(t *testing.T) {
	db := NewUnitTestDBProvider()
	v := newTestRecord()
	created, err := CreateOne(context.Background(), db, v)
	if err != nil {
		t.Fatal(err)
	}
	initialCreatedAt := created.CreatedAt
	initialUpdatedAt := created.UpdatedAt

	// Small delay to ensure update timestamp differs
	time.Sleep(10 * time.Millisecond)

	copied := created.Copy()
	copied.UpdatedAt = time.Time{}
	err = UpdateOne(context.Background(), db, copied)
	if err != nil {
		t.Fatal(err)
	}

	queried, exists, err := GetOneById(context.Background(), db, &testRecord{}, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("record should exist")
	}

	if queried.CreatedAt != initialCreatedAt {
		t.Fatal("CreatedAt should not change on update")
	}
	if queried.UpdatedAt.Equal(initialUpdatedAt) {
		t.Fatal("UpdatedAt should change on update")
	}
	if queried.UpdatedAt.Before(initialUpdatedAt) {
		t.Fatal("UpdatedAt should be after initialUpdatedAt")
	}
}

func TestRecordIdUint64(t *testing.T) {
	id := RecordId(42)
	if id.Uint64() != 42 {
		t.Fatalf("Expected 42, got %d", id.Uint64())
	}
}
