package common

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

var ResourceKey = "resource"

type CrudWrapperFunctionType int

const (
	CrudWrapperFunctionGetOne CrudWrapperFunctionType = iota
	CrudWrapperFunctionGetMany
	CrudWrapperFunctionCreate
	CrudWrapperFunctionUpdate
	CrudWrapperFunctionDelete
)

// CrudWrapperFunctionAll is a shorthand method to get all CrudWrapperFunctionType values at once
var CrudWrapperFunctionAll = []CrudWrapperFunctionType{
	CrudWrapperFunctionGetOne,
	CrudWrapperFunctionGetMany,
	CrudWrapperFunctionCreate,
	CrudWrapperFunctionUpdate,
	CrudWrapperFunctionDelete,
}

// genericApiRoute is a generic implementation of the ApiRoute[T] interface
// and is used to create ApiRoute objects which can be passed into a RouteFamily.
// This is used in CrudCommon.HandleRouteTypes to auto-generate API routes from
// boilerplate CRUD methods defined on CrudCommon
type genericApiRoute[T CrudRecord] struct {
	httpMethod     HttpMethod
	path           string
	requestBody    T
	useRequestBody bool
	handle         func(route ApiRoute[T], request ApiRequest[T]) (any, int, error)
}

func (g genericApiRoute[T]) Path() (HttpMethod, string) {
	return g.httpMethod, g.path
}

func (g genericApiRoute[T]) RequestBody() (T, bool) {
	return g.requestBody, g.useRequestBody
}

func (g genericApiRoute[T]) Handler(request ApiRequest[T]) (any, int, error) {
	return g.handle(g, request)
}

// CrudCommon is a wrapper class implementing a common mechanism to
// expose CrudRecord families via authenticated REST API calls.
//
// This allows you to implement a CrudRecord with its own business
// logic and automatically generate the relevant CRUD APIs in a
// common format at the model's BaseRoute and assign them to the router
type CrudCommon[T CrudRecord] struct {
	CreateRecord func() T
	UseAuth      bool

	// Middleware is a list of gin.HandlerFunc functions that will be
	// run before the common methods (e.g. getCrudRecordById) are run
	// for any given request.
	Middleware []gin.HandlerFunc

	// DatabaseProvider is the DatabaseProvider that we will use for
	// all CRUD operations.
	DatabaseProvider DatabaseProvider

	// BaseRoute is the base route (e.g. `/user`) used for the various
	// endpoints on the CrudCommon. It is derived from the Type() function
	// on the CrudCommon created by CreateRecord, so the generated base
	// route will be the same as the database table name.
	baseRoute string
}

func NewCrudCommon[T CrudRecord](createFunc func() T, userAuth bool) *CrudCommon[T] {
	baseRoute := createFunc().Type()
	return &CrudCommon[T]{
		CreateRecord: createFunc,
		UseAuth:      userAuth,
		// baseRoute is automatically set up based on the table name of the provided createFunc
		baseRoute:        fmt.Sprintf("/%s", baseRoute),
		DatabaseProvider: GlobalDatabaseProvider,
	}
}

// createCrudRecord creates a CrudRecord based on the given ApiRequest
func (c *CrudCommon[T]) createCrudRecord(route ApiRoute[T], request ApiRequest[T]) (any, int, error) {

	// token must be passed into any create request so that we can assign an owner to the record
	// for later API requests such as updates, subsequent access-controlled reads and deletes
	if request.Token == nil {
		return nil, http.StatusBadRequest, errors.New("token is required for create endpoints")
	}

	// get the body from the request and overwrite the ownership with the user ID that was present in the token
	body := request.Body
	body.SetOwner(request.Token.UserId)

	// Create endpoint does not need ownership/access authentication as there is not a
	// record yet which has any AccessibleTo / EditableBy functions defined.
	//
	// On creation, the CrudRecord which is being created will set all the fields necessary
	// to validate these things on future get/update/delete requests (e.g. setting a team ID
	// to enforce accessible-only-to-team constraints or setting a user ID to enforce only
	// editable by creator constraints
	// a
	v, err := CreateOne(c.DatabaseProvider, body)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	return gin.H{ResourceKey: v}, http.StatusOK, nil
}

// getCrudRecordById gets a CrudRecord base on the type T that this CrudCommon is configured to use,
// verifying that the user who made the ApiRequest is allowed to access the record based on the
// type-specific AccessibleTo logic for the CrudRecord type
func (c *CrudCommon[T]) getCrudRecordById(route ApiRoute[T], req ApiRequest[T]) (t any, status int, err error) {
	// we need to instantiate a record to pass into GetOneById
	recordType, _ := route.RequestBody()

	// helper class to validate that the ApiRequest passed in here is able to access this record
	wac := WithAccessControl[T]{Database: c.DatabaseProvider, AccessControlUser: getTokenUserIdIfExists(req)}
	v, exists, err := wac.GetOneById(recordType, req.PathId)
	if err != nil {
		return t, http.StatusBadRequest, err
	}
	if !exists {
		return t, http.StatusNotFound, nil
	}
	return gin.H{ResourceKey: v}, http.StatusOK, nil
}

// getAllCrudRecordsById gets all CrudRecord of the type T that this CrudCommon is configured to use,
// which are accessible to the user who made the ApiRequest.
func (c *CrudCommon[T]) getAllCrudRecords(route ApiRoute[T], req ApiRequest[T]) (t any, status int, err error) {
	body, _ := route.RequestBody()

	wac := WithAccessControl[T]{Database: c.DatabaseProvider, AccessControlUser: getTokenUserIdIfExists(req)}
	v, err := wac.GetAll(body)
	if err != nil {
		return t, http.StatusInternalServerError, err
	}
	return gin.H{ResourceKey: v}, http.StatusOK, nil
}

// deleteCrudRecordById deletes a CrudRecord by RecordId as long as the user in the provided ApiRequest is able to do so
func (c *CrudCommon[T]) deleteCrudRecordById(route ApiRoute[T], req ApiRequest[T]) (t any, status int, err error) {
	if req.Token == nil {
		return t, http.StatusBadRequest, errors.New("token must be passed into delete route")
	}

	wac := WithAccessControl[T]{Database: c.DatabaseProvider, AccessControlUser: getTokenUserIdIfExists(req)}
	recordType, _ := route.RequestBody()
	v, exists, err := wac.DeleteOneById(recordType, req.PathId)
	if err != nil {
		return t, http.StatusBadRequest, err
	}

	if !exists {
		return t, http.StatusOK, nil
	}

	return gin.H{ResourceKey: v}, http.StatusOK, err
}

func (c *CrudCommon[T]) updateCrudRecord(route ApiRoute[T], request ApiRequest[T]) (t any, status int, err error) {
	if request.Token == nil {
		return t, http.StatusBadRequest, errors.New("token must be passed into delete route")
	}
	// set the path ID on the request body so that our GetOneById call at the top of
	// wax.UpdateOneById gets the correct record from the DatabaseProvider
	request.Body.SetId(request.PathId)

	wac := WithAccessControl[T]{Database: c.DatabaseProvider, AccessControlUser: getTokenUserIdIfExists(request)}
	err = wac.UpdateOneById(request.Body)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	return gin.H{ResourceKey: request.Body}, http.StatusOK, nil
}

// HandleRouteTypes configures which CrudWrapperFunctionType values this CrudCommon should listen for,
// e.g. CrudWrapperFunctionAll for all HTTP methods or a specific subset of methods such as create,
// get all, and delete
func (c *CrudCommon[T]) HandleRouteTypes(e *gin.RouterGroup, crudRouteTypes ...CrudWrapperFunctionType) {
	if c.CreateRecord == nil {
		panic("CreateRecord function was not set on CrudCommon instance")
	}

	// create a RouteFamily for this CrudCommon
	f := RouteFamily[T]{
		UseAuth: c.UseAuth,
	}

	// generate as many genericApiRoutes as we need
	routes := make([]ApiRoute[T], 0, len(crudRouteTypes))
	for _, f := range crudRouteTypes {
		r := genericApiRoute[T]{requestBody: c.CreateRecord()}
		if f == CrudWrapperFunctionGetOne {
			r.httpMethod = HttpMethodGet
			r.path = AppendPathId(c.baseRoute)
			r.handle = c.getCrudRecordById
		} else if f == CrudWrapperFunctionGetMany {
			r.httpMethod = HttpMethodGet
			r.path = c.baseRoute
			r.handle = c.getAllCrudRecords
		} else if f == CrudWrapperFunctionDelete {
			r.httpMethod = HttpMethodDelete
			r.path = AppendPathId(c.baseRoute)
			r.handle = c.deleteCrudRecordById
		} else if f == CrudWrapperFunctionCreate {
			r.httpMethod = HttpMethodPost
			r.path = c.baseRoute
			r.useRequestBody = true
			r.handle = c.createCrudRecord
		} else if f == CrudWrapperFunctionUpdate {
			r.httpMethod = HttpMethodPut
			r.path = AppendPathId(c.baseRoute)
			r.useRequestBody = true
			r.handle = c.updateCrudRecord
		}
		routes = append(routes, r)
	}
	f.Handle(e, routes...)
}

func getTokenUserIdIfExists[T CrudRecord](req ApiRequest[T]) RecordId {
	userId := InvalidRecordId
	if req.Token != nil {
		userId = req.Token.UserId
	}
	return userId
}
