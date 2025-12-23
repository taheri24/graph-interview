package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents an error response (moved here to avoid import cycle)
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// Recovery returns a middleware that recovers from any panics and writes a 500 response.
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(error); ok {
			FullErrorCapture(err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request"})
		}
		c.Abort()
	})
}

// TODO: implement FullErrorCapture with log and persis error details on the file
var FullErrorCapture func(err error) = func(err error) {
	// Default implementation - will be overridden in tests
}
