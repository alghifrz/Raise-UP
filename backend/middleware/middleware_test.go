package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/middleware"
	"github.com/diuk/raiseup/pkg/authctx"
	jwtutil "github.com/diuk/raiseup/pkg/jwt"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRequireAuthValidToken(t *testing.T) {
	tokens := jwtutil.NewManager("test-secret-at-least-16", time.Hour)
	token, _, err := tokens.Generate("user-1", "admin@example.com", db.UserRoleADMINRW)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	router := gin.New()
	router.GET("/secure", middleware.RequireAuth(tokens), func(c *gin.Context) {
		principal, ok := authctx.GetPrincipal(c)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "missing principal"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"role": principal.Role, "email": principal.Email})
	})

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
}

func TestRequireAuthMissingToken(t *testing.T) {
	tokens := jwtutil.NewManager("test-secret-at-least-16", time.Hour)
	router := gin.New()
	router.GET("/secure", middleware.RequireAuth(tokens), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}

	var body map[string]map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if body["error"]["code"] != "UNAUTHORIZED" {
		t.Fatalf("code = %q, want UNAUTHORIZED", body["error"]["code"])
	}
}

func TestRequireAuthInvalidToken(t *testing.T) {
	tokens := jwtutil.NewManager("test-secret-at-least-16", time.Hour)
	router := gin.New()
	router.GET("/secure", middleware.RequireAuth(tokens), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.value")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestRequireRoleForbidden(t *testing.T) {
	tokens := jwtutil.NewManager("test-secret-at-least-16", time.Hour)
	token, _, err := tokens.Generate("user-1", "admin@example.com", db.UserRoleADMINRW)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	router := gin.New()
	router.GET(
		"/admin",
		middleware.RequireAuth(tokens),
		middleware.RequireRole(db.UserRoleSUPERADMIN),
		func(c *gin.Context) { c.Status(http.StatusOK) },
	)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}

	var body map[string]map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if body["error"]["code"] != "FORBIDDEN" {
		t.Fatalf("code = %q, want FORBIDDEN", body["error"]["code"])
	}
}

func TestRequireRoleAllowed(t *testing.T) {
	tokens := jwtutil.NewManager("test-secret-at-least-16", time.Hour)
	token, _, err := tokens.Generate("user-1", "admin@example.com", db.UserRoleADMINRW)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	router := gin.New()
	router.GET(
		"/rw",
		middleware.RequireAuth(tokens),
		middleware.RequireRole(db.UserRoleADMINRW, db.UserRoleSUPERADMIN),
		func(c *gin.Context) { c.Status(http.StatusOK) },
	)

	req := httptest.NewRequest(http.MethodGet, "/rw", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
}
