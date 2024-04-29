package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"intraclub/common"
	"intraclub/model"
	"net/http"
)

var TokenHeaderKey = "x-session-token"

func WithToken(c *gin.Context) {

	t := c.Request.Header.Get(TokenHeaderKey)
	if t == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No token found"})
		return
	}

	token, err := model.ParseToken(t)
	if err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token was not valid"})
		return
	}

	c.Set(common.TokenContextKey, token)
	c.Next()
}
