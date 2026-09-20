package resident

import db "github.com/diuk/raiseup/db/generated"

// CreateRequest is the JSON body for POST /api/v1/residents.
type CreateRequest struct {
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	Gender string `json:"gender"`
}

// UpdateRequest is the JSON body for PATCH /api/v1/residents/:id.
// Nil fields are left unchanged.
type UpdateRequest struct {
	Name   *string `json:"name"`
	Phone  *string `json:"phone"`
	Gender *string `json:"gender"`
}

// ListParams holds validated list filters and pagination.
type ListParams struct {
	Page     int
	PageSize int
	Search   string
	Gender   *db.Gender
}

// ListMeta is pagination metadata for list responses.
type ListMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ListResult is the service result for listing residents.
type ListResult struct {
	Items []Resident `json:"-"`
	Meta  ListMeta   `json:"-"`
}

// Resident is the API-safe resident representation.
type Resident struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Gender    db.Gender `json:"gender"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}
