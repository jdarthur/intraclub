package database

import (
	"context"
	"fmt"
	"testing"
)

type PrivateTestRecord struct {
	ID       RecordId `json:"id"`
	Owner    UserId
	SharedTo []UserId
	Value    string
}

func (p *PrivateTestRecord) GetOwner() UserId {
	return p.Owner
}

func (p *PrivateTestRecord) SetOwner(userId UserId) {
	p.Owner = userId
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

func (p *PrivateTestRecord) EditableBy(_ context.Context, db Provider) []UserId {
	return []UserId{p.Owner, SysAdminUserId}
}

func (p *PrivateTestRecord) AccessibleTo(_ context.Context, db Provider) []UserId {
	v := make([]UserId, 0, 1+len(p.SharedTo))
	v = append(v, p.Owner)
	v = append(v, p.SharedTo...)
	return v
}

func (p *PrivateTestRecord) StaticallyValid() error {
	return nil
}

func (p *PrivateTestRecord) DynamicallyValid(_ context.Context, db Provider) error {
	return nil
}

func (p *PrivateTestRecord) NewRecord() CrudRecord {
	return new(PrivateTestRecord)
}

func (p *PrivateTestRecord) ShareTo(ctx context.Context, db Provider, shareToUserId, updateUserId UserId) error {
	for _, s := range p.SharedTo {
		if shareToUserId == s {
			return nil
		}
	}
	p.SharedTo = append(p.SharedTo, shareToUserId)

	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: updateUserId}
	return wac.UpdateOneById(ctx, p)
}

func newStoredPrivateTestRecord(t *testing.T, db Provider, owner UserId) *PrivateTestRecord {
	r := NewPrivateTestRecord()
	r.Owner = owner
	v, err := CreateOne(context.Background(), db, r)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func testGetOneViaOwner(t *testing.T, db Provider) {
	userId := UserId(NewRecordId())
	r := newStoredPrivateTestRecord(t, db, userId)

	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: userId}
	v, exists, err := wac.GetOneById(context.Background(), r, r.GetId())
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("record should exist in GetOneById w/ owner ID as accessor")
	}
	fmt.Printf("%+v\n", v)
}

func testGetOneViaSharedTo(t *testing.T, db Provider) {
	ownerId := UserId(NewRecordId())
	r := newStoredPrivateTestRecord(t, db, ownerId)

	sharedToUserId := UserId(NewRecordId())
	err := r.ShareTo(context.Background(), db, sharedToUserId, ownerId)
	if err != nil {
		t.Fatal(err)
	}

	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: sharedToUserId}
	v, exists, err := wac.GetOneById(context.Background(), r, r.GetId())
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("record should exist in GetOneById w/ shared user ID as accessor")
	}
	fmt.Printf("%+v\n", v)
}

func testGetOneUnauthorized(t *testing.T, db Provider) {
	ownerId := UserId(NewRecordId())
	r := newStoredPrivateTestRecord(t, db, ownerId)

	// attempt to get the record via another user ID
	otherUserId := UserId(NewRecordId())
	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: otherUserId}
	_, exists, err := wac.GetOneById(context.Background(), r, r.GetId())
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("record should not exist in GetOneById w/ other user ID as accessor")
	}
}

func testDeleteOneByOwner(t *testing.T, db Provider) {
	ownerId := UserId(NewRecordId())
	r := newStoredPrivateTestRecord(t, db, ownerId)

	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: ownerId}
	v, exists, err := wac.DeleteOneById(context.Background(), r, r.GetId())
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("record should exist in GetOneById w/ other user ID as accessor")
	}
	fmt.Printf("%+v\n", v)
}

func testDeleteOneByUnauthorized(t *testing.T, db Provider) {
	ownerId := UserId(NewRecordId())
	r := newStoredPrivateTestRecord(t, db, ownerId)

	otherUserId := UserId(NewRecordId())
	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: otherUserId}
	_, exists, err := wac.DeleteOneById(context.Background(), r, r.GetId())
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("Expected record not to exist when deleting from other user")
	}
}

func testUpdateOneByOwner(t *testing.T, db Provider) {
	ownerId := UserId(NewRecordId())
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

func updateRecordAndReQuery(t *testing.T, db Provider, r *PrivateTestRecord, newValue string, asUser UserId) (*PrivateTestRecord, error) {
	copied := updateRecordIntoCopy(r, newValue)

	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: asUser}
	err := wac.UpdateOneById(context.Background(), copied)

	v, exists, err2 := GetOneById(context.Background(), db, r, r.GetId())
	if err2 != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("record should exist in GetOneById w/o access control")
	}
	return v, err
}

func testUpdateOneByUnauthorized(t *testing.T, db Provider) {
	// create a record in the database
	ownerId := UserId(NewRecordId())
	r := newStoredPrivateTestRecord(t, db, ownerId)

	// attempt to update via another user ID
	otherUserId := UserId(NewRecordId())
	v, err := updateRecordAndReQuery(t, db, r, "new value", otherUserId)
	if err == nil {
		t.Fatal("expected an error updating by unauthorized user")
	} else if v.Value != "" {
		t.Fatalf("Expected value to be unset, got %s", v.Value)
	}
}

func testAccessibleByEveryone(t *testing.T, db Provider) {
	// create a record in the database
	ownerId := UserId(NewRecordId())
	r := newStoredPrivateTestRecord(t, db, ownerId)
	err := r.ShareTo(context.Background(), db, EveryoneUserId, ownerId)
	if err != nil {
		t.Fatal(err)
	}

	// attempt to get via another user ID
	otherUserId := UserId(NewRecordId())
	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: otherUserId}
	v, exists, err := wac.GetOneById(context.Background(), r, r.GetId())
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("public record should exist in GetOneById w/ other user ID as accessor")
	}
	fmt.Printf("%+v\n", v)
}

func testAccessibleBySysAdmin(t *testing.T, db Provider) {
	ownerId := UserId(NewRecordId())
	r := newStoredPrivateTestRecord(t, db, ownerId)

	sysAdminId := UserId(NewRecordId())
	SysAdminCheck = func(ctx context.Context, db Provider, userId UserId) (bool, error) {
		return userId == sysAdminId, nil
	}
	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: sysAdminId}
	v, exists, err := wac.GetOneById(context.Background(), r, r.GetId())
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("public record should exist in GetOneById w/ sys admin user ID as accessor")
	}
	fmt.Printf("%+v\n", v)
}

func testEditableBySysAdmin(t *testing.T, db Provider) {
	ownerId := UserId(NewRecordId())
	r := newStoredPrivateTestRecord(t, db, ownerId)
	fmt.Printf("%+v\n", r)

	sysAdminId := UserId(NewRecordId())
	SysAdminCheck = func(ctx context.Context, db Provider, userId UserId) (bool, error) {
		return userId == sysAdminId, nil
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

func testGetAllByOwner(t *testing.T, db Provider) {
	ownerId := UserId(NewRecordId())
	otherId := UserId(NewRecordId())

	// Create a record owned by ownerId
	_ = newStoredPrivateTestRecord(t, db, ownerId)
	_ = newStoredPrivateTestRecord(t, db, ownerId)
	// Create a record owned by someone else
	newStoredPrivateTestRecord(t, db, otherId)

	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: ownerId}
	results, err := wac.GetAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 records, got %d", len(results))
	}
}

func testGetAllByUnauthorizedUser(t *testing.T, db Provider) {
	ownerId := UserId(NewRecordId())
	unauthorizedId := UserId(NewRecordId())

	newStoredPrivateTestRecord(t, db, ownerId)

	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: unauthorizedId}
	results, err := wac.GetAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 records for unauthorized user, got %d", len(results))
	}
}

func testGetAllBySysAdmin(t *testing.T, db Provider) {
	ownerId := UserId(NewRecordId())
	adminId := UserId(NewRecordId())

	newStoredPrivateTestRecord(t, db, ownerId)

	SysAdminCheck = func(ctx context.Context, db Provider, userId UserId) (bool, error) {
		return userId == adminId, nil
	}

	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: adminId}
	results, err := wac.GetAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 record for sys admin, got %d", len(results))
	}
}

func testDeleteOneByIdNotFoundNonOwner(t *testing.T, db Provider) {
	ownerId := UserId(NewRecordId())
	newStoredPrivateTestRecord(t, db, ownerId)

	nonExistentId := NewRecordId()
	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: ownerId}
	_, exists, err := wac.DeleteOneById(context.Background(), &PrivateTestRecord{}, nonExistentId)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("should not exist for non-existent record")
	}
}

func testUpdateOneByIdNotFound(t *testing.T, db Provider) {
	ownerId := UserId(NewRecordId())
	newStoredPrivateTestRecord(t, db, ownerId)

	nonExistent := &PrivateTestRecord{
		ID:    NewRecordId(),
		Owner: ownerId,
	}
	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: ownerId}
	err := wac.UpdateOneById(context.Background(), nonExistent)
	if err == nil {
		t.Fatal("expected error updating non-existent record")
	}
}

func testDeleteBySharedToUser(t *testing.T, db Provider) {
	ownerId := UserId(NewRecordId())
	r := newStoredPrivateTestRecord(t, db, ownerId)

	sharedToId := UserId(NewRecordId())
	err := r.ShareTo(context.Background(), db, sharedToId, ownerId)
	if err != nil {
		t.Fatal(err)
	}

	// Shared user should be able to read but not delete (only owner/sysadmin can edit)
	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: sharedToId}
	v, exists, err := wac.GetOneById(context.Background(), &PrivateTestRecord{}, r.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("shared user should be able to read")
	}
	fmt.Printf("%+v\n", v)

	// Shared user should NOT be able to delete
	_, exists, err = wac.DeleteOneById(context.Background(), &PrivateTestRecord{}, r.ID)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("shared user should not be able to delete")
	}
}

func testUpdateBySharedToUser(t *testing.T, db Provider) {
	ownerId := UserId(NewRecordId())
	r := newStoredPrivateTestRecord(t, db, ownerId)

	sharedToId := UserId(NewRecordId())
	err := r.ShareTo(context.Background(), db, sharedToId, ownerId)
	if err != nil {
		t.Fatal(err)
	}

	// Shared user should NOT be able to update
	updated := &PrivateTestRecord{
		ID:    r.ID,
		Owner: ownerId,
		Value: "new value",
	}
	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: sharedToId}
	err = wac.UpdateOneById(context.Background(), updated)
	if err == nil {
		t.Fatal("shared user should not be able to update")
	}
}

func testGetOneBySharedToUser(t *testing.T, db Provider) {
	ownerId := UserId(NewRecordId())
	r := newStoredPrivateTestRecord(t, db, ownerId)

	sharedToId := UserId(NewRecordId())
	err := r.ShareTo(context.Background(), db, sharedToId, ownerId)
	if err != nil {
		t.Fatal(err)
	}

	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: sharedToId}
	v, exists, err := wac.GetOneById(context.Background(), &PrivateTestRecord{}, r.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("shared user should be able to read")
	}
	fmt.Printf("%+v\n", v)
}

func testUpdateOneBySharedToUser(t *testing.T, db Provider) {
	ownerId := UserId(NewRecordId())
	r := newStoredPrivateTestRecord(t, db, ownerId)

	// Share the record
	sharedToId := UserId(NewRecordId())
	err := r.ShareTo(context.Background(), db, sharedToId, ownerId)
	if err != nil {
		t.Fatal(err)
	}

	// Shared user tries to update - should fail
	updated := &PrivateTestRecord{
		ID:    r.ID,
		Owner: ownerId,
		Value: "updated by shared user",
	}
	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: sharedToId}
	err = wac.UpdateOneById(context.Background(), updated)
	if err == nil {
		t.Fatal("shared user should not be able to update")
	}
}

func testGetOneIdDoesNotExist(t *testing.T, db Provider) {
	ownerId := UserId(NewRecordId())
	newStoredPrivateTestRecord(t, db, ownerId)

	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: ownerId}
	_, exists, err := wac.GetOneById(context.Background(), &PrivateTestRecord{}, NewRecordId())
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("should not exist for non-existent ID")
	}
}

func testGetAllWithMultipleOwners(t *testing.T, db Provider) {
	owner1 := UserId(NewRecordId())
	owner2 := UserId(NewRecordId())
	owner3 := UserId(NewRecordId())

	newStoredPrivateTestRecord(t, db, owner1)
	newStoredPrivateTestRecord(t, db, owner1)
	newStoredPrivateTestRecord(t, db, owner2)
	newStoredPrivateTestRecord(t, db, owner3)

	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: owner1}
	results, err := wac.GetAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 records for owner1, got %d", len(results))
	}
}

func testDeleteBySysAdmin(t *testing.T, db Provider) {
	ownerId := UserId(NewRecordId())
	r := newStoredPrivateTestRecord(t, db, ownerId)

	sysAdminId := UserId(NewRecordId())
	SysAdminCheck = func(ctx context.Context, db Provider, userId UserId) (bool, error) {
		return userId == sysAdminId, nil
	}

	wac := WithAccessControl[*PrivateTestRecord]{Database: db, AccessControlUser: sysAdminId}
	v, exists, err := wac.DeleteOneById(context.Background(), &PrivateTestRecord{}, r.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("sys admin should be able to delete")
	}
	fmt.Printf("%+v\n", v)
}

// TestDbWithAccessControlContract runs every TestDbWithAccessControlContract test body against every database.Provider kind
// (memory, SQLite in-memory, SQLite on-disk) via runContractTests (issue #62).
func TestDbWithAccessControlContract(t *testing.T) {
	runContractTests(t, []contractTest{
		{name: "TestGetOneViaOwner", fn: testGetOneViaOwner},
		{name: "TestGetOneViaSharedTo", fn: testGetOneViaSharedTo},
		{name: "TestGetOneUnauthorized", fn: testGetOneUnauthorized},
		{name: "TestDeleteOneByOwner", fn: testDeleteOneByOwner},
		{name: "TestDeleteOneByUnauthorized", fn: testDeleteOneByUnauthorized},
		{name: "TestUpdateOneByOwner", fn: testUpdateOneByOwner},
		{name: "TestUpdateOneByUnauthorized", fn: testUpdateOneByUnauthorized},
		{name: "TestAccessibleByEveryone", fn: testAccessibleByEveryone},
		{name: "TestAccessibleBySysAdmin", fn: testAccessibleBySysAdmin},
		{name: "TestEditableBySysAdmin", fn: testEditableBySysAdmin},
		{name: "TestGetAllByOwner", fn: testGetAllByOwner},
		{name: "TestGetAllByUnauthorizedUser", fn: testGetAllByUnauthorizedUser},
		{name: "TestGetAllBySysAdmin", fn: testGetAllBySysAdmin},
		{name: "TestDeleteOneByIdNotFoundNonOwner", fn: testDeleteOneByIdNotFoundNonOwner},
		{name: "TestUpdateOneByIdNotFound", fn: testUpdateOneByIdNotFound},
		{name: "TestDeleteBySharedToUser", fn: testDeleteBySharedToUser},
		{name: "TestUpdateBySharedToUser", fn: testUpdateBySharedToUser},
		{name: "TestGetOneBySharedToUser", fn: testGetOneBySharedToUser},
		{name: "TestUpdateOneBySharedToUser", fn: testUpdateOneBySharedToUser},
		{name: "TestGetOneIdDoesNotExist", fn: testGetOneIdDoesNotExist},
		{name: "TestGetAllWithMultipleOwners", fn: testGetAllWithMultipleOwners},
		{name: "TestDeleteBySysAdmin", fn: testDeleteBySysAdmin},
	})
}
