package api

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"intraclub/database"

	"github.com/gin-gonic/gin"
)

type testCrudRecord struct {
	ID        database.RecordId
	Owner     database.UserId
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (t *testCrudRecord) GetOwner() database.UserId {
	return t.Owner
}

func (t *testCrudRecord) SetOwner(userId database.UserId) {
	t.Owner = userId
}

func newTestCrudRecord() *testCrudRecord {
	return &testCrudRecord{
		ID:    database.NewRecordId(),
		Value: "default",
	}
}

func (t *testCrudRecord) Type() string {
	return "test_crud"
}

func (t *testCrudRecord) GetId() database.RecordId {
	return t.ID
}

func (t *testCrudRecord) SetId(id database.RecordId) {
	t.ID = id
}

func (t *testCrudRecord) EditableBy(_ context.Context, db database.Provider) []database.UserId {
	return []database.UserId{t.Owner, database.SysAdminUserId}
}

func (t *testCrudRecord) AccessibleTo(_ context.Context, db database.Provider) []database.UserId {
	return database.AccessibleToEveryone
}

func (t *testCrudRecord) StaticallyValid() error {
	if t.Value == "" {
		return fmt.Errorf("value cannot be empty")
	}
	return nil
}

func (t *testCrudRecord) DynamicallyValid(_ context.Context, db database.Provider) error {
	return nil
}

func (t *testCrudRecord) GetTimeStamps() (created, updated time.Time) {
	return t.CreatedAt, t.UpdatedAt
}

func (t *testCrudRecord) SetCreateTimestamp(tm time.Time) time.Time {
	oldValue := t.CreatedAt
	t.CreatedAt = tm
	return oldValue
}

func (t *testCrudRecord) SetUpdateTimestamp(tm time.Time) time.Time {
	oldValue := t.UpdatedAt
	t.UpdatedAt = tm
	return oldValue
}

func (t *testCrudRecord) NewRecord() database.CrudRecord {
	return new(testCrudRecord)
}

func TestNewCrudCommon(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	cc := NewCrudCommon(newTestCrudRecord, true, db)

	if cc.CreateRecord == nil {
		t.Fatal("CreateRecord should be set")
	}
	if !cc.UseAuth {
		t.Fatal("UseAuth should be true")
	}
	if cc.DatabaseProvider != db {
		t.Fatal("DatabaseProvider should be set")
	}
	if cc.baseRoute != "/test_crud" {
		t.Fatalf("expected baseRoute '/test_crud', got '%s'", cc.baseRoute)
	}
}

func TestNewCrudCommonNoAuth(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	cc := NewCrudCommon(newTestCrudRecord, false, db)

	if cc.UseAuth {
		t.Fatal("UseAuth should be false")
	}
}

func TestCrudCommonCreateMissingToken(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	cc := NewCrudCommon(newTestCrudRecord, true, db)

	route := genericApiRoute[*testCrudRecord]{
		requestBody:    newTestCrudRecord,
		useRequestBody: true,
		handle:         cc.createCrudRecord,
	}

	req := ApiRequest[*testCrudRecord]{
		Context:          context.Background(),
		DatabaseProvider: db,
		Token:            nil,
		Body:             newTestCrudRecord(),
	}

	_, status, err := route.Handler(req)
	if err == nil {
		t.Fatal("expected error for missing token")
	}
	if status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, status)
	}
}

func TestCrudCommonCreateSuccess(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	cc := NewCrudCommon(newTestCrudRecord, true, db)

	route := genericApiRoute[*testCrudRecord]{
		requestBody:    newTestCrudRecord,
		useRequestBody: true,
		handle:         cc.createCrudRecord,
	}

	ownerId := database.NewRecordId()
	req := ApiRequest[*testCrudRecord]{
		Context:          context.Background(),
		DatabaseProvider: db,
		Token:            &AuthToken{UserId: database.UserId(ownerId)},
		Body:             newTestCrudRecord(),
	}

	resp, status, err := route.Handler(req)
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, status)
	}

	result := resp.(gin.H)
	resource := result[ResourceKey]
	if resource.(*testCrudRecord).Owner != database.UserId(ownerId) {
		t.Fatal("Owner should be set from token")
	}
	if resource.(*testCrudRecord).ID == database.InvalidRecordId {
		t.Fatal("ID should be set")
	}
}

func TestCrudCommonGetOneSuccess(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	cc := NewCrudCommon(newTestCrudRecord, true, db)

	// Create a record first
	ownerId := database.NewRecordId()
	v := newTestCrudRecord()
	v.Owner = database.UserId(ownerId)
	created, err := database.CreateOne(context.Background(), db, v)
	if err != nil {
		t.Fatal(err)
	}

	route := genericApiRoute[*testCrudRecord]{
		requestBody: newTestCrudRecord,
		handle:      cc.getCrudRecordById,
	}

	req := ApiRequest[*testCrudRecord]{
		Context:          context.Background(),
		DatabaseProvider: db,
		Token:            &AuthToken{UserId: database.UserId(ownerId)},
		PathId:           created.ID,
	}

	resp, status, err := route.Handler(req)
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, status)
	}

	result := resp.(gin.H)
	resource := result[ResourceKey].(*testCrudRecord)
	if resource.ID != created.ID {
		t.Fatalf("expected ID %s, got %s", created.ID, resource.ID)
	}
}

func TestCrudCommonGetOneNotFound(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	cc := NewCrudCommon(newTestCrudRecord, true, db)

	route := genericApiRoute[*testCrudRecord]{
		requestBody: newTestCrudRecord,
		handle:      cc.getCrudRecordById,
	}

	req := ApiRequest[*testCrudRecord]{
		Context:          context.Background(),
		DatabaseProvider: db,
		Token:            &AuthToken{UserId: database.InvalidUserId},
		PathId:           database.NewRecordId(),
	}

	_, status, err := route.Handler(req)
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, status)
	}
}

func TestCrudCommonGetAllSuccess(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	cc := NewCrudCommon(newTestCrudRecord, true, db)

	ownerId := database.NewRecordId()
	for i := 0; i < 3; i++ {
		v := newTestCrudRecord()
		v.Owner = database.UserId(ownerId)
		_, err := database.CreateOne(context.Background(), db, v)
		if err != nil {
			t.Fatal(err)
		}
	}

	route := genericApiRoute[*testCrudRecord]{
		requestBody: newTestCrudRecord,
		handle:      cc.getAllCrudRecords,
	}

	req := ApiRequest[*testCrudRecord]{
		Context:          context.Background(),
		DatabaseProvider: db,
		Token:            &AuthToken{UserId: database.UserId(ownerId)},
	}

	resp, status, err := route.Handler(req)
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, status)
	}

	result := resp.(gin.H)
	records := result[ResourceKey].([]*testCrudRecord)
	if len(records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(records))
	}
}

func TestCrudCommonDeleteSuccess(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	cc := NewCrudCommon(newTestCrudRecord, true, db)

	ownerId := database.NewRecordId()
	v := newTestCrudRecord()
	v.Owner = database.UserId(ownerId)
	created, err := database.CreateOne(context.Background(), db, v)
	if err != nil {
		t.Fatal(err)
	}

	route := genericApiRoute[*testCrudRecord]{
		requestBody: newTestCrudRecord,
		handle:      cc.deleteCrudRecordById,
	}

	req := ApiRequest[*testCrudRecord]{
		Context:          context.Background(),
		DatabaseProvider: db,
		Token:            &AuthToken{UserId: database.UserId(ownerId)},
		PathId:           created.ID,
	}

	resp, status, err := route.Handler(req)
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, status)
	}

	result := resp.(gin.H)
	deleted := result[ResourceKey].(*testCrudRecord)
	if deleted.ID != created.ID {
		t.Fatalf("expected ID %s, got %s", created.ID, deleted.ID)
	}
}

func TestCrudCommonDeleteMissingToken(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	cc := NewCrudCommon(newTestCrudRecord, true, db)

	route := genericApiRoute[*testCrudRecord]{
		requestBody: newTestCrudRecord,
		handle:      cc.deleteCrudRecordById,
	}

	req := ApiRequest[*testCrudRecord]{
		Context:          context.Background(),
		DatabaseProvider: db,
		Token:            nil,
		PathId:           database.NewRecordId(),
	}

	_, status, err := route.Handler(req)
	if err == nil {
		t.Fatal("expected error for missing token")
	}
	if status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, status)
	}
}

func TestCrudCommonDeleteNotFound(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	cc := NewCrudCommon(newTestCrudRecord, true, db)

	ownerId := database.NewRecordId()
	route := genericApiRoute[*testCrudRecord]{
		requestBody: newTestCrudRecord,
		handle:      cc.deleteCrudRecordById,
	}

	req := ApiRequest[*testCrudRecord]{
		Context:          context.Background(),
		DatabaseProvider: db,
		Token:            &AuthToken{UserId: database.UserId(ownerId)},
		PathId:           database.NewRecordId(),
	}

	_, status, err := route.Handler(req)
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, status)
	}
}

func TestCrudCommonUpdateSuccess(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	cc := NewCrudCommon(newTestCrudRecord, true, db)

	ownerId := database.NewRecordId()
	v := newTestCrudRecord()
	v.Owner = database.UserId(ownerId)
	created, err := database.CreateOne(context.Background(), db, v)
	if err != nil {
		t.Fatal(err)
	}

	route := genericApiRoute[*testCrudRecord]{
		requestBody:    newTestCrudRecord,
		useRequestBody: true,
		handle:         cc.updateCrudRecord,
	}

	req := ApiRequest[*testCrudRecord]{
		Context:          context.Background(),
		DatabaseProvider: db,
		Token:            &AuthToken{UserId: database.UserId(ownerId)},
		PathId:           created.ID,
		Body: &testCrudRecord{
			ID:    created.ID,
			Owner: database.UserId(ownerId),
			Value: "updated",
		},
	}

	resp, status, err := route.Handler(req)
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, status)
	}

	result := resp.(gin.H)
	updated := result[ResourceKey].(*testCrudRecord)
	if updated.Value != "updated" {
		t.Fatalf("expected value 'updated', got '%s'", updated.Value)
	}
}

func TestCrudCommonUpdateMissingToken(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	cc := NewCrudCommon(newTestCrudRecord, true, db)

	route := genericApiRoute[*testCrudRecord]{
		requestBody:    newTestCrudRecord,
		useRequestBody: true,
		handle:         cc.updateCrudRecord,
	}

	req := ApiRequest[*testCrudRecord]{
		Context:          context.Background(),
		DatabaseProvider: db,
		Token:            nil,
		PathId:           database.NewRecordId(),
		Body:             newTestCrudRecord(),
	}

	_, status, err := route.Handler(req)
	if err == nil {
		t.Fatal("expected error for missing token")
	}
	if status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, status)
	}
}

func TestCrudCommonCreateEmptyValue(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	cc := NewCrudCommon(newTestCrudRecord, true, db)

	route := genericApiRoute[*testCrudRecord]{
		requestBody:    func() *testCrudRecord { r := newTestCrudRecord(); r.Value = ""; return r },
		useRequestBody: true,
		handle:         cc.createCrudRecord,
	}

	ownerId := database.NewRecordId()
	req := ApiRequest[*testCrudRecord]{
		Context:          context.Background(),
		DatabaseProvider: db,
		Token:            &AuthToken{UserId: database.UserId(ownerId)},
		Body:             func() *testCrudRecord { r := newTestCrudRecord(); r.Value = ""; return r }(),
	}

	_, status, err := route.Handler(req)
	if err == nil {
		t.Fatal("expected error for empty value")
	}
	if status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, status)
	}
}

func TestCrudCommonGetOneAccessibleToEveryone(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	cc := NewCrudCommon(newTestCrudRecord, true, db)

	ownerId := database.NewRecordId()
	v := newTestCrudRecord()
	v.Owner = database.UserId(ownerId)
	created, err := database.CreateOne(context.Background(), db, v)
	if err != nil {
		t.Fatal(err)
	}

	route := genericApiRoute[*testCrudRecord]{
		requestBody: newTestCrudRecord,
		handle:      cc.getCrudRecordById,
	}

	// Even a "non-owner" user can access it because AccessibleTo returns AccessibleToEveryone
	nonOwnerId := database.NewRecordId()
	req := ApiRequest[*testCrudRecord]{
		Context:          context.Background(),
		DatabaseProvider: db,
		Token:            &AuthToken{UserId: database.UserId(nonOwnerId)},
		PathId:           created.ID,
	}

	_, status, err := route.Handler(req)
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusOK {
		t.Fatalf("expected status %d (accessible to everyone), got %d", http.StatusOK, status)
	}
}

func TestCrudCommonDeleteUnauthorized(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	cc := NewCrudCommon(newTestCrudRecord, true, db)

	ownerId := database.NewRecordId()
	v := newTestCrudRecord()
	v.Owner = database.UserId(ownerId)
	created, err := database.CreateOne(context.Background(), db, v)
	if err != nil {
		t.Fatal(err)
	}

	route := genericApiRoute[*testCrudRecord]{
		requestBody: newTestCrudRecord,
		handle:      cc.deleteCrudRecordById,
	}

	unauthorizedId := database.NewRecordId()
	req := ApiRequest[*testCrudRecord]{
		Context:          context.Background(),
		DatabaseProvider: db,
		Token:            &AuthToken{UserId: database.UserId(unauthorizedId)},
		PathId:           created.ID,
	}

	_, status, err := route.Handler(req)
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, status)
	}
}

func TestCrudCommonHandleRouteTypesPanicOnNilCreateRecord(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when CreateRecord is nil")
		}
	}()

	db := database.NewUnitTestDBProvider()
	cc := &CrudCommon[*testCrudRecord]{
		CreateRecord:     nil,
		UseAuth:          true,
		DatabaseProvider: db,
		baseRoute:        "/test",
	}

	gin.SetMode(gin.TestMode)
	ginRouter := gin.New().Group("")
	cc.HandleRouteTypes(ginRouter, CrudWrapperFunctionGetOne)
}
