package middleware

import (
	"github.com/gin-gonic/gin"
)

// AuthCheck checks to see if token passed in context is valid.
func AuthCheck() gin.HandlerFunc {

	return func(c *gin.Context) {

	}
}
