package database

import (
	"context"
	"fmt"
	"testing"
)

type staticFailRecord struct {
	ID    RecordId
	Value string
}

func (s *staticFailRecord) GetOwner() RecordId {
	return InvalidRecordId
}

func (s *staticFailRecord) SetOwner(recordId RecordId) {}

func (s *staticFailRecord) Type() string {
	return "static_fail"
}

func (s *staticFailRecord) GetId() RecordId {
	return s.ID
}

func (s *staticFailRecord) SetId(id RecordId) {
	s.ID = id
}

func (s *staticFailRecord) EditableBy(_ context.Context, db DatabaseProvider) []RecordId {
	return nil
}

func (s *staticFailRecord) AccessibleTo(_ context.Context, db DatabaseProvider) []RecordId {
	return nil
}

func (s *staticFailRecord) StaticallyValid() error {
	if s.Value == "" {
		return fmt.Errorf("value is required")
	}
	return nil
}

func (s *staticFailRecord) DynamicallyValid(_ context.Context, db DatabaseProvider) error {
	return nil
}

func (s *staticFailRecord) BlankRecord() CrudRecord {
	return new(staticFailRecord)
}

type dynamicFailRecord struct {
	ID    RecordId
	Value string
}

func (d *dynamicFailRecord) GetOwner() RecordId {
	return InvalidRecordId
}

func (d *dynamicFailRecord) SetOwner(recordId RecordId) {}

func (d *dynamicFailRecord) Type() string {
	return "dynamic_fail"
}

func (d *dynamicFailRecord) GetId() RecordId {
	return d.ID
}

func (d *dynamicFailRecord) SetId(id RecordId) {
	d.ID = id
}

func (d *dynamicFailRecord) EditableBy(_ context.Context, db DatabaseProvider) []RecordId {
	return nil
}

func (d *dynamicFailRecord) AccessibleTo(_ context.Context, db DatabaseProvider) []RecordId {
	return nil
}

func (d *dynamicFailRecord) StaticallyValid() error {
	return nil
}

func (d *dynamicFailRecord) DynamicallyValid(_ context.Context, db DatabaseProvider) error {
	if d.Value == "invalid" {
		return fmt.Errorf("value cannot be 'invalid'")
	}
	return nil
}

func (d *dynamicFailRecord) BlankRecord() CrudRecord {
	return new(dynamicFailRecord)
}

func TestValidateStaticFails(t *testing.T) {
	db := NewUnitTestDBProvider()
	record := &staticFailRecord{
		Value: "",
	}

	err := Validate(context.Background(), db, record)
	if err == nil {
		t.Fatal("expected validation error for empty value")
	}
}

func TestValidateDynamicFails(t *testing.T) {
	db := NewUnitTestDBProvider()
	record := &dynamicFailRecord{
		Value: "invalid",
	}

	err := Validate(context.Background(), db, record)
	if err == nil {
		t.Fatal("expected validation error for 'invalid' value")
	}
}

func TestValidateBothPass(t *testing.T) {
	db := NewUnitTestDBProvider()
	record := &staticFailRecord{
		Value: "good",
	}

	err := Validate(context.Background(), db, record)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateDynamicFailsAfterStaticPasses(t *testing.T) {
	db := NewUnitTestDBProvider()
	record := &dynamicFailRecord{
		Value: "invalid",
	}

	err := Validate(context.Background(), db, record)
	if err == nil {
		t.Fatal("expected validation error for 'invalid' value")
	}
}
