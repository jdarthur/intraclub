package common

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

type UserBasedRecord interface {
	GetUserId() string
	SetUserId(userId string)
	CrudRecord
}

var ResourceKey = "resource"

type ControllerType interface {
	Model() CrudRecord
	ValidateRequest(c CrudRecord, isUpdate bool, provider DbProvider) (CrudRecord, error)
	GetAllFilter(c *gin.Context) (map[string]interface{}, error)
}

type CrudController struct {
	Controller ControllerType
	Database   DbProvider
}

func NewCrudController(controller ControllerType) *CrudController {
	return &CrudController{
		Controller: controller,
		Database:   GlobalDbProvider,
	}
}

func (cc *CrudController) SetGlobalDbIfNoDbProvided() {
	if cc.Database == nil {
		cc.Database = GlobalDbProvider
	}
}

func (cc *CrudController) GetOne(c *gin.Context) {
	// check if the :id field was a valid record for cc.ControllerType
	existingRecord := cc.idValidation(c)
	if existingRecord == nil {
		return
	}

	c.JSON(http.StatusOK, gin.H{ResourceKey: existingRecord})
}

func (cc *CrudController) GetAll(c *gin.Context) {
	filter, err := cc.Controller.GetAllFilter(c)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	v, err := GetAllWhere(cc.Database, cc.Controller.Model(), filter)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{ResourceKey: v})
}

func (cc *CrudController) Update(c *gin.Context) {

	request, ok := IsPostOwnedByUser(c)
	if !ok {
		request = cc.Controller.Model()
		err := c.Bind(request)
		if err != nil {
			RespondWithBadRequest(c, err)
			return
		}
	}

	fmt.Println(request)

	record, err := cc.Controller.ValidateRequest(request, true, cc.Database)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	// check if the :id field was a valid record for cc.ControllerType
	existingRecord := cc.idValidation(c)
	if existingRecord == nil {
		return
	}

	// make sure that the record has its ID set based off of what
	// we just pulled from the DB. This will prevent us from taking in
	// an object ID from the request that doesn't match the :id param
	record.SetId(existingRecord.GetId())

	// we will automatically set the user ID on this record based off of the user
	// in the token if this record is a UserBasedRecord
	_, ok = record.(UserBasedRecord)
	if ok {
		token, err := GetTokenFromAuthMiddleware(c)
		if err != nil {
			RespondWithError(c, err)
			return
		}

		record.(UserBasedRecord).SetUserId(token.UserId)
	}

	fmt.Println(record)

	err = Update(cc.Database, record)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{ResourceKey: record})
}

func (cc *CrudController) Create(c *gin.Context) {

	schema := cc.Controller.Model()
	err := c.Bind(schema)
	if err != nil {
		RespondWithBadRequest(c, err)
		return
	}

	// validate the payload first for illegal values
	record, err := cc.Controller.ValidateRequest(schema, false, cc.Database)
	if err != nil {
		RespondWithBadRequest(c, err)
		return
	}

	if !record.GetId().IsZero() {
		RespondWithBadRequest(c, err)
		return
	}

	// we will automatically set the user ID on this record based off of the user
	// in the token if this record is a UserBasedRecord
	_, ok := record.(UserBasedRecord)
	if ok {
		token, err := GetTokenFromAuthMiddleware(c)
		if err != nil {
			RespondWithError(c, err)
			return
		}

		record.(UserBasedRecord).SetUserId(token.UserId)
	}

	created, err := Create(cc.Database, record)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{ResourceKey: created})
}

func (cc *CrudController) Delete(c *gin.Context) {

	// check if the :id field was a valid record for cc.ControllerType
	existingRecord := cc.idValidation(c)
	if existingRecord == nil {
		return
	}

	err := Delete(cc.Database, existingRecord)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{ResourceKey: existingRecord})
}

func getId(c *gin.Context) (primitive.ObjectID, error) {
	id := c.Param("id")
	if id == "" {
		return primitive.ObjectID{}, errors.New(":id field was not provided")
	}

	return primitive.ObjectIDFromHex(id)
}

// idValidation checks that the provided :id field in the request was a valid
// object ID for the CrudController's ControllerType.Model. If not valid, this
// function will respond with an error code on the *gin.Context and return nil
func (cc *CrudController) idValidation(c *gin.Context) (recordIfExists CrudRecord) {
	return IdValidation(c, cc.Controller.Model(), cc.Database)
}

func IdValidation(c *gin.Context, schema CrudRecord, db DbProvider) (recordIfExists CrudRecord) {
	id, err := getId(c)
	if err != nil {
		RespondWithBadRequest(c, err)
		return nil
	}

	schema.SetId(id)

	record, exists, err := GetOne(db, schema)
	if !exists {
		RespondWithError(c, RecordDoesNotExist(schema))
		return nil
	}

	if err != nil {
		RespondWithError(c, err)
		return nil
	}

	return record
}

var UsingOwnedByUser = "using_owned_by_user"      // mark on the gin.Context that we are using this middleware
var OwnedByUserRecordKey = "owned_by_user_record" // store the common.CrudRecord that we parsed here

// IsPostOwnedByUser is a check to see if we have already called the OwnedByUserWrapper.OwnedByUser
// middleware on this particular gin.Context. In this situation, we don't want to call c.Bind(model)
// a second time, so we will just pull the common.CrudRecord off of the gin.Context instead.
func IsPostOwnedByUser(c *gin.Context) (CrudRecord, bool) {

	v, ok := c.Get(UsingOwnedByUser)
	if ok && v.(bool) == true {

		v2, ok := c.Get(OwnedByUserRecordKey)
		if !ok {
			panic("UsingOwnedByUser was set but OwnedByUserRecordKey was not")
		}

		return v2.(CrudRecord), true
	}

	return nil, false
}
