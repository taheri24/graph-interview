package middleware

import (
	"net/http"

	"taheri24.ir/graph1/internal/dto"

	"github.com/gin-gonic/gin"
)

// Recovery returns a middleware that recovers from any panics and writes a 500 response.
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(error); ok {
			FullErrorCapture(err)
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request"})
		}
		c.Abort()
	})
}

// TODO: implement FullErrorCapture with log and persis error details on the file
var FullErrorCapture func(err error) = func(err error) {
	// Default implementation - will be overridden in tests
}
