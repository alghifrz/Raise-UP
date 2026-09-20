package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorBody is the standard API error envelope.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error writes a consistent JSON error response.
func Error(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gin.H{
		"error": ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

// JSON writes a successful JSON response with a data envelope.
func JSON(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data})
}

// JSONWithMeta writes a successful JSON response with data and meta envelopes.
func JSONWithMeta(c *gin.Context, status int, data any, meta any) {
	c.JSON(status, gin.H{
		"data": data,
		"meta": meta,
	})
}

// BadRequest is a convenience helper for INVALID_REQUEST.
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, "INVALID_REQUEST", message)
}

// Unauthorized is a convenience helper for UNAUTHORIZED.
func Unauthorized(c *gin.Context, code, message string) {
	Error(c, http.StatusUnauthorized, code, message)
}

// Forbidden is a convenience helper for FORBIDDEN.
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound is a convenience helper for resource-not-found errors.
func NotFound(c *gin.Context, code, message string) {
	Error(c, http.StatusNotFound, code, message)
}

// Conflict is a convenience helper for conflict errors.
func Conflict(c *gin.Context, code, message string) {
	Error(c, http.StatusConflict, code, message)
}

// NoContent writes an empty 204 response.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Internal is a convenience helper for INTERNAL_ERROR.
func Internal(c *gin.Context) {
	Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred")
}
