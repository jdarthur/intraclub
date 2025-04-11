package common

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

// PathIdField is the field in a request to an ApiRoute
// that specifies a RecordId, such as a user ID in `GET /users/:id`
var PathIdField = "id"

func AppendPathId(route string) string {
	return fmt.Sprintf("%s/:%s", route, PathIdField)
}

// HttpMethod is a type used to enforce method correctness in ApiRoute
type HttpMethod int

const (
	HttpMethodGet HttpMethod = iota
	HttpMethodPost
	HttpMethodPut
	HttpMethodDelete
	HttpMethodInvalid
)

func (m HttpMethod) String() string {
	switch m {
	case HttpMethodGet:
		return "GET"
	case HttpMethodPost:
		return "POST"
	case HttpMethodPut:
		return "PUT"
	case HttpMethodDelete:
		return "DELETE"
	default:
		return "INVALID"
	}
}

func (m HttpMethod) Valid() bool {
	return m < HttpMethodInvalid
}

// ApiRoute is a generic API route interface which can be accessed
// at a particular path, with an optional request body and with a
// handler function which returns a response value, an HTTP response
// code, and an error if the request was invalid or something went wrong
type ApiRoute[T Validatable] interface {

	// Path returns the HttpMethod and route for this ApiRout
	Path() (method HttpMethod, route string)

	// RequestBody instantiates a new instance of the Validatable
	// type passed a type parameter, or false if no request body is used.
	//
	// This is used to create a new object to pass into deserialization
	// functions such as gin.Bind which pull the body out of the gin context
	RequestBody() (t T, usesRequestBody bool)

	// Handler is a function which handles an ApiRequest of the given type
	// and returns a response and successful status code (if the request was
	// valid) or an error and unsuccessful status code (if the request failed
	// for whatever reason, e.g. invalid body, failed access control checks,
	// internal server such as database issues, etc.)
	Handler(request ApiRequest[T]) (response any, statusCode int, error error)
}

// ApiRequest is a typed struct used to pass in a defined set of parameters to an ApiRoute
type ApiRequest[T Validatable] struct {
	// PathId is the RecordId parsed out of the path e.g. the `id` from `/user/:id`
	PathId RecordId

	// Body is the request body from the gin.Context, if applicable to the ApiRoute.
	// This can be validated using the Validatable functions defined for the type
	Body T

	// Token is the AuthToken passed into the request, if one was passed in.
	// This can be used to enforce access control on get requests and auto-assign
	// values such as record owners on create requests
	Token *AuthToken

	// DatabaseProvider is a DatabaseProvider interface passed to the
	// ApiRoute if it needs to do something in the database. This will not be
	// used unless the route is manually using the RouteFamily struct
	DatabaseProvider DatabaseProvider
}

// RouteFamily is a typed helper struct to implement a common interface for
// API endpoints. Most API routes should use common helpers such as CrudCommon,
// but this helper can be used for e.g. special one-off routes or routes
// which don't conform to the normal constraints around CRUD access/edit rights
type RouteFamily[T Validatable] struct {
	// UseAuth requires all requests to this RouteFamily to be sent with a valid JWT (if true)
	UseAuth bool

	// DatabaseProvider is the DatabaseProvider that the Handler functions will use
	DatabaseProvider DatabaseProvider

	// Middleware is a list of gin.HandlerFunc functions that will be
	// run before the wrapper's handler functions are run for each request
	Middleware []gin.HandlerFunc

	// wrappers are added via Handle
	wrappers []*routeWrapper[T]
}

// Handle adds one or more ApiRoutes to this RouteFamily and applies the
// routes to the provided gin.Engine.
func (r *RouteFamily[T]) Handle(e *gin.RouterGroup, routes ...ApiRoute[T]) {

	// if DatabaseProvider was not set on the RouteFamily, we will use the global one
	if r.DatabaseProvider == nil {
		r.DatabaseProvider = GlobalDatabaseProvider
	}

	// create a RouteWrapper for each ApiRoute provided
	for _, route := range routes {
		wrapper := &routeWrapper[T]{
			Route:    route,
			Database: r.DatabaseProvider,
			UseAuth:  r.UseAuth,
		}
		r.wrappers = append(r.wrappers, wrapper)
	}

	// add the routes to the engine
	r.addToEngine(e)
}

// addToEngine adds all the routes in the wrapper list to the given gin.Engine.
// This is a shim layer to convert the RouteFamily syntax into the format
// needed by a gin.Engine (i.e. getting the method and route from each ApiRoute
// and applying the middleware and RouteWrapper.Handle function to the engine)
func (r *RouteFamily[T]) addToEngine(e *gin.RouterGroup) {
	// handle each ApiRoute in the family
	for _, wrapper := range r.wrappers {

		// get method + path from ApiRoute
		method, route := wrapper.Route.Path()

		// add Handle() function to the end of the family's middleware chain
		handlers := r.Middleware
		handlers = append(handlers, wrapper.Handle)

		// handle the route
		e.Handle(method.String(), route, handlers...)
	}
}

// routeWrapper is an internal struct which wraps an ApiRoute with a DatabaseProvider
// and stores whether the given ApiRoute will need authentication. This is used inside
// the RouteFamily helper functions and applies common DB and auth logic to an ApiRoute
type routeWrapper[T Validatable] struct {
	Route    ApiRoute[T]
	Database DatabaseProvider
	UseAuth  bool
}

// Handle parses a raw request from a gin.Context, gets the token from the request (if
// necessary) and converts it into a typed ApiRequest which is passed to the ApiRoute's
// handler function for processing.
func (r *routeWrapper[T]) Handle(c *gin.Context) {
	var token *AuthToken
	var err error

	// get token from request if configured for this route
	if r.UseAuth {
		token, err = GetToken(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
	}

	// parse ID from path
	pathId, err := r.getPathId(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// get the request into the format that ApiRoute implementers will expect
	apiRequest := ApiRequest[T]{
		PathId:           pathId,
		Token:            token,
		DatabaseProvider: r.Database,
	}

	// parse out the body of the request if this route uses a request body
	b, useBody := r.Route.RequestBody()
	if useBody {
		err := c.Bind(b)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		apiRequest.Body = b
	}

	// handle the response via the ApiRoute's route handler
	resp, response, err := r.Route.Handler(apiRequest)
	if err != nil {
		c.AbortWithStatusJSON(response, gin.H{"error": err.Error()})
		return
	}
	c.JSON(response, resp)
}

// getPathId parses the PathIdField from the raw request (if present) and converts
// it into a RecordId which will be passed into the ApiRequest provided to the
// ApiRoute Handler function which processes the request
func (r *routeWrapper[T]) getPathId(c *gin.Context) (RecordId, error) {
	// get the :id field from the request path
	v, ok := c.Params.Get(PathIdField)
	if ok {

		// parse the record ID into a base-10 uint64
		id, err := RecordIdFromString(v)
		if err != nil {
			return InvalidRecordId, fmt.Errorf("invalid field for :%s path parameter: %s", PathIdField, v)
		}

		// convert the uint64 into a RecordId (which is a uint64 wrapper type)
		return RecordId(id), nil
	}

	// return default value if not present not all ApiRoutes will use a path ID
	return InvalidRecordId, nil
}
