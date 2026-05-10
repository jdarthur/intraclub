package route

import (
	"github.com/gin-gonic/gin"
	"intraclub/common"
	"intraclub/model"
	"net/http"
)

var UserImportBaseRoute = "/user_import"

type CsvImportResult struct {
	CreatedCount    int
	Created         []*model.User
	AlreadyExisting []*model.User
}

type CsvImportHandler struct {
	DatabaseProvider common.DatabaseProvider
}

func (h *CsvImportHandler) HandleCsvImport(c *gin.Context) {
	v, ok := c.GetPostForm("file")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "form file not provided"})
		return
	}

	userList, err := model.ParseUserCsvFromString(v)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	created, existing, err := model.ParseAndCreateCsvUsers(c.Request.Context(), h.DatabaseProvider, userList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, CsvImportResult{
		CreatedCount:    len(created),
		Created:         created,
		AlreadyExisting: existing,
	})
}
