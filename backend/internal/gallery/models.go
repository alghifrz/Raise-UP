package gallery

// UploadRequest contains a validated image and its gallery metadata.
type UploadRequest struct {
	FileName    string
	ContentType string
	Data        []byte
	Caption     string
	SortOrder   *int32
}

// ReplaceImageRequest contains a replacement image and optional metadata updates.
type ReplaceImageRequest struct {
	UploadRequest
	UpdateCaption   bool
	UpdateSortOrder bool
}

// CreateRequest is the JSON body for POST /api/v1/gallery.
type CreateRequest struct {
	ImageURL    string `json:"image_url"`
	StoragePath string `json:"storage_path"`
	Caption     string `json:"caption"`
	SortOrder   *int32 `json:"sort_order"`
}

// UpdateRequest is the JSON body for PATCH /api/v1/gallery/:id.
type UpdateRequest struct {
	ImageURL    *string `json:"image_url"`
	StoragePath *string `json:"storage_path"`
	Caption     *string `json:"caption"`
	SortOrder   *int32  `json:"sort_order"`
}

// ListMeta is pagination metadata.
type ListMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ListResult is the service result for listing gallery items.
type ListResult struct {
	Items []Item
	Meta  ListMeta
}

// Item is the API-safe gallery item representation.
type Item struct {
	ID          string `json:"id"`
	ImageURL    string `json:"image_url"`
	StoragePath string `json:"storage_path"`
	Caption     string `json:"caption"`
	SortOrder   int32  `json:"sort_order"`
	CreatedAt   string `json:"created_at"`
}
