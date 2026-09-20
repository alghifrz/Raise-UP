package village

// UpdateProfileRequest is the JSON body for PATCH /api/v1/village-profile.
type UpdateProfileRequest struct {
	History *string `json:"history"`
	Vision  *string `json:"vision"`
	Mission *string `json:"mission"`
}

// CreateOfficialRequest is the JSON body for POST /api/v1/village-profile/officials.
type CreateOfficialRequest struct {
	Name      string `json:"name"`
	Position  string `json:"position"`
	PhotoURL  string `json:"photo_url"`
	SortOrder *int32 `json:"sort_order"`
}

// UpdateOfficialRequest is the JSON body for PATCH /api/v1/village-profile/officials/:id.
// profile_id is intentionally omitted (not client-writable).
type UpdateOfficialRequest struct {
	Name      *string `json:"name"`
	Position  *string `json:"position"`
	PhotoURL  *string `json:"photo_url"`
	SortOrder *int32  `json:"sort_order"`
}

// Profile is the API-safe village profile representation.
type Profile struct {
	ID        string `json:"id"`
	History   string `json:"history"`
	Vision    string `json:"vision"`
	Mission   string `json:"mission"`
	UpdatedAt string `json:"updated_at"`
}

// Official is the API-safe village official representation.
type Official struct {
	ID        string  `json:"id"`
	ProfileID string  `json:"profile_id"`
	Name      string  `json:"name"`
	Position  string  `json:"position"`
	PhotoURL  *string `json:"photo_url"`
	SortOrder int32   `json:"sort_order"`
}
