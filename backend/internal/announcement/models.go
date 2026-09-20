package announcement

import db "github.com/diuk/raiseup/db/generated"

// CreateRequest is the JSON body for POST /api/v1/announcements.
type CreateRequest struct {
	Title        string   `json:"title"`
	Excerpt      string   `json:"excerpt"`
	Body         string   `json:"body"`
	Category     string   `json:"category"`
	Visibility   string   `json:"visibility"`
	ThumbnailURL *string  `json:"thumbnail_url"`
	RecipientIDs []string `json:"recipient_ids"`
}

// UpdateRequest is the JSON body for PATCH /api/v1/announcements/:id.
// Nil pointer fields are left unchanged. RecipientIDs is replaced when non-nil.
type UpdateRequest struct {
	Title        *string   `json:"title"`
	Excerpt      *string   `json:"excerpt"`
	Body         *string   `json:"body"`
	Category     *string   `json:"category"`
	Visibility   *string   `json:"visibility"`
	ThumbnailURL *string   `json:"thumbnail_url"`
	RecipientIDs *[]string `json:"recipient_ids"`
}

// StatusRequest is the JSON body for PATCH /api/v1/announcements/:id/status.
type StatusRequest struct {
	Status string `json:"status"`
}

// ListMeta is pagination metadata.
type ListMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ListResult is the service result for listing announcements.
type ListResult struct {
	Items []AnnouncementSummary
	Meta  ListMeta
}

// Announcement is the API-safe announcement representation.
type Announcement struct {
	ID           string                    `json:"id"`
	Title        string                    `json:"title"`
	Excerpt      string                    `json:"excerpt"`
	Body         string                    `json:"body"`
	Category     string                    `json:"category"`
	Visibility   db.AnnouncementVisibility `json:"visibility"`
	Status       db.AnnouncementStatus     `json:"status"`
	ThumbnailURL *string                   `json:"thumbnail_url"`
	AuthorID     string                    `json:"author_id"`
	PublishedAt  *string                   `json:"published_at"`
	CreatedAt    string                    `json:"created_at"`
	UpdatedAt    string                    `json:"updated_at"`
	RecipientIDs []string                  `json:"recipient_ids"`
}

// AnnouncementSummary is used for list responses (no recipient IDs).
type AnnouncementSummary struct {
	ID           string                    `json:"id"`
	Title        string                    `json:"title"`
	Excerpt      string                    `json:"excerpt"`
	Body         string                    `json:"body"`
	Category     string                    `json:"category"`
	Visibility   db.AnnouncementVisibility `json:"visibility"`
	Status       db.AnnouncementStatus     `json:"status"`
	ThumbnailURL *string                   `json:"thumbnail_url"`
	AuthorID     string                    `json:"author_id"`
	PublishedAt  *string                   `json:"published_at"`
	CreatedAt    string                    `json:"created_at"`
	UpdatedAt    string                    `json:"updated_at"`
}
