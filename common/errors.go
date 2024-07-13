package common

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"runtime"
)

// ApiError is an error wrapper that provides a common way to structure errors
// encountered in the functions provided by the API.
type ApiError struct {
	// References can be one or more values that will be passed into the ErrorCode.String
	// function. These will most likely end up being fmt.Sprintf arguments
	References []interface{} `json:"references"`

	// Code is the ErrorCode that this error represents. It allows us to provide
	// an HTTP status code and a unique error ID
	Code ErrorCode `json:"code"`
}

func (a ApiError) Error() string {
	return a.Code.String(a.References)
}

type ErrorCode int

const (
	UserWithEmailDoesNotExist          ErrorCode = iota // attempted to create a one-time password for a nonexistent user
	MultipleUsersExistForEmail                          // shouldn't happen due to API validation
	UserHasNoActiveOneTimePasswords                     // user tried to authenticate via OTP when none existed
	InvalidUuidForUsername                              // uuid provided in "create token" request did not match what was expected for the user
	TokenWasNotPresentOnGinContext                      // programming error when you forget to put the WithToken middleware before another middleware that relies on the token being set on the context
	InvalidPrimitiveObjectId                            // passing a bad primitive.ObjectId to a function that takes a string version
	CrudRecordWithObjectIdDoesNotExist                  // CrudRecord with given objectId does not exist
	FieldNotUpdatable                                   // Attempted to update a field in violation of the HasNonUpdatable rules
	FieldMustBeGloballyUnique                           // Attempted to create or update a record in violation of a ValueMustBeGloballyUnique constraint
	UserIsNotAdmin                                      // User attempted to call a route protected by middleware.AsAdminUser but was not an admin user
	FieldIsRequired                                     // User did not provide a value for a required field
	FacilityMustHaveAtLeastOneCourt                     // create or update a model.Facility where `courts` = 0
	InvalidNestedObjectId                               // InvalidPrimitiveObjectId, but with a CrudRecord object type attached
)

func (e ErrorCode) String(references []any) string {
	return [...]string{
		fmt.Sprintf("User with email '%s' was not found", references...),         // UserWithEmailDoesNotExist
		fmt.Sprintf("Multiple users exist with email '%s'", references...),       // MultipleUsersExistForEmail
		fmt.Sprintf("User '%s' has no active one-time passwords", references...), // UserHasNoActiveOneTimePasswords
		fmt.Sprintf("Invalid one-time password for user '%s'", references...),    // InvalidUuidForUsername
		"Token was not present on gin context",                                   // TokenWasNotPresentOnGinContext
		fmt.Sprintf("Invalid object ID: '%s'", references...),                    // InvalidPrimitiveObjectId
		fmt.Sprintf("%s with ID '%s' does not exist", references...),             // CrudRecordWithObjectIdDoesNotExist
		fmt.Sprintf("Field %s is not updatable", references...),                  // FieldNotUpdatable
		fmt.Sprintf("%s with %s %v already exists", references...),               //FieldMustBeGloballyUnique
		fmt.Sprintf("User %+v is not an admin user", references...),              // UserIsNotAdmin
		fmt.Sprintf("Field '%s' is required", references...),                     // FieldIsRequired
		"Facility must have at least 1 court",                                    // FacilityMustHaveAtLeastOneCourt
		fmt.Sprintf("Invalid object ID for nested %s: '%s'", references...),      // InvalidNestedObjectId
	}[e]
}

func (e ErrorCode) HttpStatus() int {
	return [...]int{
		http.StatusNotFound,            // UserWithEmailDoesNotExist
		http.StatusInternalServerError, // MultipleUsersExistForEmail
		http.StatusUnauthorized,        // UserHasNoActiveOneTimePasswords
		http.StatusBadRequest,          // InvalidUuidForUsername
		http.StatusInternalServerError, // TokenWasNotPresentOnGinContext
		http.StatusBadRequest,          // InvalidObjectId
		http.StatusNotFound,            // CrudRecordWithObjectIdDoesNotExist
		http.StatusBadRequest,          // FieldNotUpdatable
		http.StatusBadRequest,          // FieldMustBeGloballyUnique
		http.StatusUnauthorized,        // UserIsNotAdmin
		http.StatusBadRequest,          // FieldIsRequired
		http.StatusBadRequest,          // FacilityMustHaveAtLeastOneCourt
		http.StatusBadRequest,          //InvalidNestedObjectId
	}[e]
}

func RespondWithError(c *gin.Context, err error) {
	var apiError ApiError
	if errors.As(err, &apiError) {
		RespondWithApiError(c, apiError)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func RespondWithApiError(c *gin.Context, apiError ApiError) {

	_, x, y, ok := runtime.Caller(2)
	if ok {
		fmt.Printf("ApiError at: %s:%d\n", x, y)
	}

	c.JSON(apiError.Code.HttpStatus(), gin.H{"error": apiError.Error(), "code": int(apiError.Code)})
}

func RespondWithBadRequest(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

func TryParsingObjectId(objectId string) (primitive.ObjectID, error) {
	id, err := primitive.ObjectIDFromHex(objectId)
	if err != nil {
		return primitive.ObjectID{}, InvalidObjectIdError(objectId)
	}

	return id, nil
}

func InvalidObjectIdError(objectId string) ApiError {
	return ApiError{
		References: []any{objectId},
		Code:       InvalidPrimitiveObjectId,
	}
}
