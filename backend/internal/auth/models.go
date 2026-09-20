package auth

import db "github.com/diuk/raiseup/db/generated"

// LoginRequest is the JSON body for POST /api/v1/auth/login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// PublicUser is the safe user representation returned by the API.
type PublicUser struct {
	ID    string      `json:"id"`
	Email string      `json:"email"`
	Name  string      `json:"name"`
	Role  db.UserRole `json:"role"`
}

// LoginResult is returned by a successful login.
type LoginResult struct {
	AccessToken string     `json:"access_token"`
	TokenType   string     `json:"token_type"`
	ExpiresIn   int64      `json:"expires_in"`
	User        PublicUser `json:"user"`
}
