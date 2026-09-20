package middleware

import (
	"strings"

	"github.com/diuk/raiseup/pkg/authctx"
	jwtutil "github.com/diuk/raiseup/pkg/jwt"
	"github.com/diuk/raiseup/pkg/response"
	"github.com/gin-gonic/gin"
)

// RequireAuth validates a Bearer JWT and stores the principal in context.
func RequireAuth(tokens *jwtutil.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Unauthorized(c, "UNAUTHORIZED", "Authentication required")
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			response.Unauthorized(c, "UNAUTHORIZED", "Authentication required")
			return
		}

		claims, err := tokens.Parse(strings.TrimSpace(parts[1]))
		if err != nil {
			response.Unauthorized(c, "UNAUTHORIZED", "Authentication required")
			return
		}

		authctx.SetPrincipal(c, authctx.Principal{
			UserID: claims.UserID,
			Email:  claims.Email,
			Role:   claims.Role,
		})
		c.Next()
	}
}
