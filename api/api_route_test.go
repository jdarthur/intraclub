package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"intraclub/database"

	"github.com/gin-gonic/gin"
)

type testRouteRecord struct {
	ID    database.RecordId
	Value string
}

func (t *testRouteRecord) GetOwner() database.UserId {
	return database.InvalidUserId
}

func (t *testRouteRecord) SetOwner(userId database.UserId) {}

func (t *testRouteRecord) Type() string {
	return "test_route"
}

func (t *testRouteRecord) GetId() database.RecordId {
	return t.ID
}

func (t *testRouteRecord) SetId(id database.RecordId) {
	t.ID = id
}

func (t *testRouteRecord) EditableBy(_ context.Context, db database.Provider) []database.UserId {
	return nil
}

func (t *testRouteRecord) AccessibleTo(_ context.Context, db database.Provider) []database.UserId {
	return nil
}

func (t *testRouteRecord) StaticallyValid() error {
	return nil
}

func (t *testRouteRecord) DynamicallyValid(_ context.Context, db database.Provider) error {
	return nil
}

func (t *testRouteRecord) NewRecord() database.CrudRecord {
	return new(testRouteRecord)
}

type simpleRoute struct{}

func (s simpleRoute) Path() (HttpMethod, string) {
	return HttpMethodGet, "/simple"
}

func (s simpleRoute) RequestBody() (*testRouteRecord, bool) {
	return new(testRouteRecord), false
}

func (s simpleRoute) Handler(request ApiRequest[*testRouteRecord]) (any, int, error) {
	return gin.H{"message": "ok"}, http.StatusOK, nil
}

type routeWithBody struct{}

func (r routeWithBody) Path() (HttpMethod, string) {
	return HttpMethodPost, "/withbody"
}

func (r routeWithBody) RequestBody() (*testRouteRecord, bool) {
	return new(testRouteRecord), true
}

func (r routeWithBody) Handler(request ApiRequest[*testRouteRecord]) (any, int, error) {
	return gin.H{"value": request.Body.Value}, http.StatusOK, nil
}

type routeWithId struct{}

func (r routeWithId) Path() (HttpMethod, string) {
	return HttpMethodGet, "/:id"
}

func (r routeWithId) RequestBody() (*testRouteRecord, bool) {
	return new(testRouteRecord), false
}

func (r routeWithId) Handler(request ApiRequest[*testRouteRecord]) (any, int, error) {
	return gin.H{"id": request.PathId.String()}, http.StatusOK, nil
}

func TestRouteFamilyHandle(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	family := RouteFamily[*testRouteRecord]{
		DatabaseProvider: db,
		UseAuth:          false,
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/api")
	family.Handle(group, simpleRoute{})

	// Verify the route was added
	if len(family.wrappers) != 1 {
		t.Fatalf("expected 1 wrapper, got %d", len(family.wrappers))
	}

	wrapper := family.wrappers[0]
	if wrapper.Database != db {
		t.Fatal("wrapper should have the same DatabaseProvider")
	}
	if wrapper.UseAuth != false {
		t.Fatal("wrapper.UseAuth should match RouteFamily")
	}
}

func TestRouteFamilyWithAuth(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	family := RouteFamily[*testRouteRecord]{
		DatabaseProvider: db,
		UseAuth:          true,
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/api")
	family.Handle(group, simpleRoute{})

	wrapper := family.wrappers[0]
	if wrapper.UseAuth != true {
		t.Fatal("wrapper.UseAuth should be true")
	}
}

func TestRouteFamilyMultipleRoutes(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	family := RouteFamily[*testRouteRecord]{
		DatabaseProvider: db,
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/api")
	family.Handle(group, simpleRoute{}, routeWithBody{}, routeWithId{})

	if len(family.wrappers) != 3 {
		t.Fatalf("expected 3 wrappers, got %d", len(family.wrappers))
	}
}

func TestRouteWrapperHandleNoAuth(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	family := RouteFamily[*testRouteRecord]{
		DatabaseProvider: db,
		UseAuth:          false,
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/api")
	family.Handle(group, simpleRoute{})

	// Make a request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/simple", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRouteWrapperHandleWithPathId(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	family := RouteFamily[*testRouteRecord]{
		DatabaseProvider: db,
		UseAuth:          false,
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/api")
	family.Handle(group, routeWithId{})

	testId := database.NewRecordId()

	// Make a request with path ID
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/%s", testId.String()), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var result map[string]any
	json.Unmarshal(w.Body.Bytes(), &result)
	if result["id"] != testId.String() {
		t.Fatalf("expected id %s, got %s", testId.String(), result["id"])
	}
}

func TestRouteWrapperHandleWithBody(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	family := RouteFamily[*testRouteRecord]{
		DatabaseProvider: db,
		UseAuth:          false,
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/api")
	family.Handle(group, routeWithBody{})

	body := `{"value":"test value"}`

	// Make a request with body
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/withbody", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRouteFamilyWithMiddleware(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	middlewareCalled := false

	family := RouteFamily[*testRouteRecord]{
		DatabaseProvider: db,
		UseAuth:          false,
		Middleware: []gin.HandlerFunc{
			func(c *gin.Context) {
				middlewareCalled = true
				c.Next()
			},
		},
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/api")
	family.Handle(group, simpleRoute{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/simple", nil)
	router.ServeHTTP(w, req)

	if !middlewareCalled {
		t.Fatal("middleware should have been called")
	}
}

func TestAppendPathId(t *testing.T) {
	result := AppendPathId("/users")
	if result != "/users/:id" {
		t.Fatalf("expected '/users/:id', got '%s'", result)
	}

	result2 := AppendPathId("/api/v1/items")
	if result2 != "/api/v1/items/:id" {
		t.Fatalf("expected '/api/v1/items/:id', got '%s'", result2)
	}
}

func TestHttpMethodString(t *testing.T) {
	tests := []struct {
		method   HttpMethod
		expected string
	}{
		{HttpMethodGet, "GET"},
		{HttpMethodPost, "POST"},
		{HttpMethodPut, "PUT"},
		{HttpMethodDelete, "DELETE"},
		{HttpMethodInvalid, "INVALID"},
	}

	for _, tt := range tests {
		result := tt.method.String()
		if result != tt.expected {
			t.Fatalf("expected '%s', got '%s'", tt.expected, result)
		}
	}
}

func TestHttpMethodValid(t *testing.T) {
	if !HttpMethodGet.Valid() {
		t.Fatal("HttpMethodGet should be valid")
	}
	if !HttpMethodPost.Valid() {
		t.Fatal("HttpMethodPost should be valid")
	}
	if !HttpMethodPut.Valid() {
		t.Fatal("HttpMethodPut should be valid")
	}
	if !HttpMethodDelete.Valid() {
		t.Fatal("HttpMethodDelete should be valid")
	}
	if HttpMethodInvalid.Valid() {
		t.Fatal("HttpMethodInvalid should not be valid")
	}
}

func TestApiRequestParseQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req, _ := http.NewRequest("GET", "/test?name=john&age=30", nil)

	var parsed struct {
		Name []string `json:"name"`
		Age  []string `json:"age"`
	}

	apiReq := ApiRequest[*testRouteRecord]{
		request: req,
	}
	err := apiReq.ParseQuery(&parsed)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed.Name) != 1 || parsed.Name[0] != "john" {
		t.Fatalf("expected name ['john'], got %v", parsed.Name)
	}
	if len(parsed.Age) != 1 || parsed.Age[0] != "30" {
		t.Fatalf("expected age ['30'], got %v", parsed.Age)
	}
}

func TestApiRequestParseQueryEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req, _ := http.NewRequest("GET", "/test", nil)

	var parsed map[string]any

	apiReq := ApiRequest[*testRouteRecord]{
		request: req,
	}
	err := apiReq.ParseQuery(&parsed)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed) != 0 {
		t.Fatalf("expected empty map, got %v", parsed)
	}
}

func TestRouteFamilyDatabaseProviderSetOnWrappers(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	family := RouteFamily[*testRouteRecord]{
		DatabaseProvider: db,
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/api")
	family.Handle(group, simpleRoute{})

	for _, wrapper := range family.wrappers {
		if wrapper.Database != db {
			t.Fatal("all wrappers should have the same DatabaseProvider as RouteFamily")
		}
	}
}

// StringReader wraps a string to implement io.Reader
type StringReader struct {
	s    string
	pos  int
	done bool
}

func (r *StringReader) Read(p []byte) (n int, err error) {
	if r.done {
		return 0, nil
	}
	if r.pos >= len(r.s) {
		r.done = true
		return 0, nil
	}
	n = copy(p, r.s[r.pos:])
	r.pos += n
	return n, nil
}
