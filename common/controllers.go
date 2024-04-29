package common

import (
	"errors"
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	v, err := GetAllWhere(cc.Database, cc.Controller.Model(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{ResourceKey: v})
}

func (cc *CrudController) Update(c *gin.Context) {

	schema := cc.Controller.Model()
	err := c.Bind(schema)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record, err := cc.Controller.ValidateRequest(schema, true, cc.Database)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check if the :id field was a valid record for cc.ControllerType
	existingRecord := cc.idValidation(c)
	if existingRecord == nil {
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

	err = Update(cc.Database, record)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{ResourceKey: record})
}

func (cc *CrudController) Create(c *gin.Context) {

	schema := cc.Controller.Model()
	err := c.Bind(schema)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// validate the payload first for illegal values
	record, err := cc.Controller.ValidateRequest(schema, false, cc.Database)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !record.GetId().IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id field must not be set in create request"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	id, err := getId(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil
	}

	schema := cc.Controller.Model()
	schema.SetId(id)

	record, exists, err := GetOne(cc.Database, schema)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": RecordDoesNotExist(schema).Error()})
		return nil
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return nil
	}

	return record

}
