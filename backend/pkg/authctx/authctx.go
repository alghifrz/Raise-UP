package authctx

import (
	db "github.com/diuk/raiseup/db/generated"
	"github.com/gin-gonic/gin"
)

const principalKey = "auth_principal"

// Principal is the authenticated identity stored in the Gin context.
type Principal struct {
	UserID string
	Email  string
	Role   db.UserRole
}

// SetPrincipal stores the authenticated principal on the request context.
func SetPrincipal(c *gin.Context, principal Principal) {
	c.Set(principalKey, principal)
}

// GetPrincipal returns the authenticated principal, if present.
func GetPrincipal(c *gin.Context) (Principal, bool) {
	value, ok := c.Get(principalKey)
	if !ok {
		return Principal{}, false
	}
	principal, ok := value.(Principal)
	return principal, ok
}
