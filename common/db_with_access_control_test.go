package common

import (
	"fmt"
	"testing"
)

type PrivateTestRecord struct {
	ID       RecordId
	Owner    RecordId
	SharedTo []RecordId
	Value    string
}

func (p *PrivateTestRecord) SetOwner(recordId RecordId) {
	p.Owner = recordId
}

func NewPrivateTestRecord() *PrivateTestRecord {
	return &PrivateTestRecord{}
}

func (p *PrivateTestRecord) Type() string {
	return "private_record"
}

func (p *PrivateTestRecord) GetId() RecordId {
	return p.ID
}

func (p *PrivateTestRecord) SetId(id RecordId) {
	p.ID = id
}

func (p *PrivateTestRecord) EditableBy(db DatabaseProvider) []RecordId {
	return []RecordId{p.Owner, SysAdminRecordId}
}

func (p *PrivateTestRecord) AccessibleTo(db DatabaseProvider) []RecordId {
	v := make([]RecordId, 0, 1+len(p.SharedTo))
	v = append(v, p.Owner)
	v = append(v, p.SharedTo...)
	return v
}

func (p *PrivateTestRecord) StaticallyValid() error {
	return nil
}

func (p *PrivateTestRecord) DynamicallyValid(db DatabaseProvider) error {
	return nil
}

func (p *PrivateTestRecord) ShareTo(db DatabaseProvider, shareToUserId, updateUserId RecordId) error {
	for _, s := range p.SharedTo {
		if shareToUserId == s {
			return nil
		}
	}
	p.SharedTo = append(p.SharedTo, shareToUserId)

	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: updateUserId}
	return wac.UpdateOneById(p)
}

func newStoredPrivateTestRecord(t *testing.T, db DatabaseProvider, owner RecordId) *PrivateTestRecord {
	r := NewPrivateTestRecord()
	r.Owner = owner
	v, err := CreateOne(db, r)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestGetOneViaOwner(t *testing.T) {
	db := NewUnitTestDBProvider()
	userId := NewRecordId()
	r := newStoredPrivateTestRecord(t, db, userId)

	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: userId}
	v, exists, err := wac.GetOneById(r, r.GetId())
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("record should exist in GetOneById w/ owner ID as accessor")
	}
	fmt.Printf("%+v\n", v)
}

func TestGetOneViaSharedTo(t *testing.T) {
	db := NewUnitTestDBProvider()
	ownerId := NewRecordId()
	r := newStoredPrivateTestRecord(t, db, ownerId)

	sharedToUserId := NewRecordId()
	err := r.ShareTo(db, sharedToUserId, ownerId)
	if err != nil {
		t.Fatal(err)
	}

	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: sharedToUserId}
	v, exists, err := wac.GetOneById(r, r.GetId())
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("record should exist in GetOneById w/ shared user ID as accessor")
	}
	fmt.Printf("%+v\n", v)
}

func TestGetOneUnauthorized(t *testing.T) {
	db := NewUnitTestDBProvider()
	ownerId := NewRecordId()
	r := newStoredPrivateTestRecord(t, db, ownerId)

	// attempt to get the record via another user ID
	otherUserId := NewRecordId()
	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: otherUserId}
	_, exists, err := wac.GetOneById(r, r.GetId())
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("record should not exist in GetOneById w/ other user ID as accessor")
	}
}

func TestDeleteOneByOwner(t *testing.T) {
	db := NewUnitTestDBProvider()
	ownerId := NewRecordId()
	r := newStoredPrivateTestRecord(t, db, ownerId)

	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: ownerId}
	v, exists, err := wac.DeleteOneById(r, r.GetId())
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("record should exist in GetOneById w/ other user ID as accessor")
	}
	fmt.Printf("%+v\n", v)
}

func TestDeleteOneByUnauthorized(t *testing.T) {
	db := NewUnitTestDBProvider()
	ownerId := NewRecordId()
	r := newStoredPrivateTestRecord(t, db, ownerId)

	otherUserId := NewRecordId()
	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: otherUserId}
	_, exists, err := wac.DeleteOneById(r, r.GetId())
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("Expected record not to exist when deleting from other user")
	}
}

func TestUpdateOneByOwner(t *testing.T) {
	db := NewUnitTestDBProvider()
	ownerId := NewRecordId()
	r := newStoredPrivateTestRecord(t, db, ownerId)

	v, err := updateRecordAndReQuery(t, db, r, "new value", ownerId)
	if err != nil {
		t.Fatal(err)
	}
	if v.Value != "new value" {
		t.Fatal("Expected new value but got " + v.Value)
	}
}

func updateRecordIntoCopy(r *PrivateTestRecord, newValue string) *PrivateTestRecord {
	// copy record and update it
	copyOfRecord := NewPrivateTestRecord()
	copyOfRecord.ID = r.ID
	copyOfRecord.Owner = r.Owner
	copyOfRecord.SharedTo = append(copyOfRecord.SharedTo, r.SharedTo...)
	copyOfRecord.Value = newValue
	return copyOfRecord
}

func updateRecordAndReQuery(t *testing.T, db DatabaseProvider, r *PrivateTestRecord, newValue string, asUser RecordId) (*PrivateTestRecord, error) {
	copied := updateRecordIntoCopy(r, newValue)

	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: asUser}
	err := wac.UpdateOneById(copied)

	v, exists, err2 := GetOneById(db, r, r.GetId())
	if err2 != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("record should exist in GetOneById w/o access control")
	}
	return v, err
}

func TestUpdateOneByUnauthorized(t *testing.T) {
	//create a record in the database
	db := NewUnitTestDBProvider()
	ownerId := NewRecordId()
	r := newStoredPrivateTestRecord(t, db, ownerId)

	// attempt to update via another user ID
	otherUserId := NewRecordId()
	v, err := updateRecordAndReQuery(t, db, r, "new value", otherUserId)
	if err == nil {
		t.Fatal("expected an error updating by unauthorized user")
	} else if v.Value != "" {
		t.Fatalf("Expected value to be unset, got %s", v.Value)
	}
}

func TestAccessibleByEveryone(t *testing.T) {
	//create a record in the database
	db := NewUnitTestDBProvider()
	ownerId := NewRecordId()
	r := newStoredPrivateTestRecord(t, db, ownerId)
	err := r.ShareTo(db, EveryoneRecordId, ownerId)
	if err != nil {
		t.Fatal(err)
	}

	// attempt to get via another user ID
	otherUserId := NewRecordId()
	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: otherUserId}
	v, exists, err := wac.GetOneById(r, r.GetId())
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("public record should exist in GetOneById w/ other user ID as accessor")
	}
	fmt.Printf("%+v\n", v)
}

func TestAccessibleBySysAdmin(t *testing.T) {
	db := NewUnitTestDBProvider()
	ownerId := NewRecordId()
	r := newStoredPrivateTestRecord(t, db, ownerId)

	sysAdminId := NewRecordId()
	SysAdminCheck = func(db DatabaseProvider, c RecordId) (bool, error) {
		return c == sysAdminId, nil
	}
	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: sysAdminId}
	v, exists, err := wac.GetOneById(r, r.GetId())
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("public record should exist in GetOneById w/ sys admin user ID as accessor")
	}
	fmt.Printf("%+v\n", v)
}

func TestEditableBySysAdmin(t *testing.T) {
	db := NewUnitTestDBProvider()
	ownerId := NewRecordId()
	r := newStoredPrivateTestRecord(t, db, ownerId)
	fmt.Printf("%+v\n", r)

	sysAdminId := NewRecordId()
	SysAdminCheck = func(db DatabaseProvider, c RecordId) (bool, error) {
		return c == sysAdminId, nil
	}
	v, err := updateRecordAndReQuery(t, db, r, "new value", sysAdminId)
	if err != nil {
		t.Fatal(err)
	}
	if v.Value != "new value" {
		t.Fatal("Expected new value but got " + v.Value)
	}
	fmt.Printf("%+v\n", v)
}
