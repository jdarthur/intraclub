package model

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"intraclub/common"
)

type CrudController struct {
	Database common.DbProvider
	Record   common.CrudRecord
}

func NotFoundError(c common.CrudRecord, id string) error {
	return fmt.Errorf("%s with ID %s was not found", c.RecordType(), id)
}

func IdNotProvidedError() error {
	return fmt.Errorf("ID was not provided")
}

func AbortWithError(c *gin.Context, err error) {
	c.JSON(500, gin.H{"error": err.Error()})
}

func ReturnData(c *gin.Context, data interface{}) {
	c.JSON(200, gin.H{"data": data})
}

func (cc *CrudController) GetAll(c *gin.Context) {
	records, err := cc.Database.GetAll(cc.Record)
	if err != nil {
		AbortWithError(c, err)
		return
	}

	ReturnData(c, records)
}

func (cc *CrudController) GetOne(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		AbortWithError(c, IdNotProvidedError())
		return
	}

	record, exists, err := cc.Database.GetOne(cc.Record, id)
	if err != nil {
		AbortWithError(c, err)
		return
	}

	if !exists {
		AbortWithError(c, NotFoundError(cc.Record, id))
		return
	}

	ReturnData(c, record)
}

func (cc *CrudController) Create(c *gin.Context) {
	// bind to record type
	body := cc.Record.OneRecord()
	err := c.Bind(body)
	if err != nil {
		AbortWithError(c, err)
	}

	// validate that the ID field is not set in the request
	id := body.(common.CrudRecord).GetId()
	if id != "" {
		AbortWithError(c, fmt.Errorf("ID field may not be set in create call"))
	}

	// create the record
	record, err := cc.Database.Create(cc.Record, body)
	if err != nil {
		AbortWithError(c, err)
		return
	}

	ReturnData(c, record)
}

func (cc *CrudController) Update(c *gin.Context) {
	// bind to record type
	body := cc.Record.OneRecord()
	err := c.Bind(body)
	if err != nil {
		AbortWithError(c, err)
	}

	// validate that the record exists
	id := body.GetId()
	_, exists, err := cc.Database.GetOne(cc.Record, id)

	// shouldn't get a DB error here, but if we do we will return it
	if err != nil {
		AbortWithError(c, err)
		return
	}

	// if record does not exist, return "record not found" error
	if !exists {
		AbortWithError(c, NotFoundError(cc.Record, id))
		return
	}

	// do actual update
	err = cc.Database.Update(cc.Record, body)
	if err != nil {
		AbortWithError(c, err)
	}

	ReturnData(c, body)
}

func (cc *CrudController) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		AbortWithError(c, IdNotProvidedError())
		return
	}

	record, exists, err := cc.Database.GetOne(cc.Record, id)
	if err != nil {
		AbortWithError(c, err)
		return
	}

	if !exists {
		AbortWithError(c, NotFoundError(cc.Record, id))
		return
	}

	err = cc.Database.Delete(cc.Record, id)
	if err != nil {
		AbortWithError(c, err)
	}

	ReturnData(c, record)
}
