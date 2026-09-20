package middleware

import (
	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/pkg/authctx"
	"github.com/diuk/raiseup/pkg/response"
	"github.com/gin-gonic/gin"
)

// RequireRole allows the request only when the authenticated user has one of the given roles.
// Must be used after RequireAuth.
func RequireRole(roles ...db.UserRole) gin.HandlerFunc {
	allowed := make(map[db.UserRole]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		principal, ok := authctx.GetPrincipal(c)
		if !ok {
			response.Unauthorized(c, "UNAUTHORIZED", "Authentication required")
			return
		}

		if _, ok := allowed[principal.Role]; !ok {
			response.Forbidden(c, "Insufficient permissions")
			return
		}

		c.Next()
	}
}
