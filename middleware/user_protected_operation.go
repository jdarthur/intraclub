package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"intraclub/common"
	"net/http"
)

type UserBasedRecord interface {
	GetUserId() string
	common.CrudRecord
}

func ValidateRecordIsOwnedByUser(r UserBasedRecord, userIdInToken string) error {
	if r.GetUserId() != userIdInToken {
		return fmt.Errorf("user %s cannot modify record owned by user %s", userIdInToken, r.GetUserId())
	}

	return nil
}

type OwnedByUserWrapper struct {
	Record UserBasedRecord
}

func (w *OwnedByUserWrapper) Bind(c *gin.Context) (common.CrudRecord, error) {

	record := w.Record.OneRecord()
	err := c.Bind(record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (w *OwnedByUserWrapper) OwnedByUser(c *gin.Context) {
	token, err := GetTokenFromAuthMiddleware(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	record, err := w.Bind(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record.SetId(objectId)

	record, exists, err := common.GetOne(common.GlobalDbProvider, record)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Println(record)

	if !exists {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": common.RecordDoesNotExist(record)})
		return
	}

	err = ValidateRecordIsOwnedByUser(record.(UserBasedRecord), token.UserId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.Next()
}
