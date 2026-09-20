package complaint

import (
	"bytes"
	"encoding/json"

	db "github.com/diuk/raiseup/db/generated"
)

// CreateRequest is the JSON body for POST /api/v1/complaints.
//
// Snapshot rule: when resident_id is provided it is authoritative. The service
// loads that resident and stores its current name/phone as the complaint
// snapshot, ignoring any resident_name/phone supplied in the same request.
type CreateRequest struct {
	ResidentID   *string `json:"resident_id"`
	ResidentName string  `json:"resident_name"`
	Phone        string  `json:"phone"`
	Block        string  `json:"block"`
	Category     string  `json:"category"`
	Urgency      string  `json:"urgency"`
	Message      string  `json:"message"`
}

// UpdateRequest is the JSON body for PATCH /api/v1/complaints/:id.
// Nil pointer fields are left unchanged. ResidentID uses OptionalUUID so JSON
// null can clear the link while preserving the stored name/phone snapshot.
type UpdateRequest struct {
	ResidentID OptionalUUID `json:"resident_id"`
	Block      *string      `json:"block"`
	Category   *string      `json:"category"`
	Urgency    *string      `json:"urgency"`
	Message    *string      `json:"message"`
}

// StatusRequest is the JSON body for PATCH /api/v1/complaints/:id/status.
type StatusRequest struct {
	Status string `json:"status"`
}

// OptionalUUID distinguishes omitted, null, and string values in PATCH bodies.
type OptionalUUID struct {
	Present bool
	Valid   bool
	Value   string
}

// UnmarshalJSON implements json.Unmarshaler.
func (o *OptionalUUID) UnmarshalJSON(data []byte) error {
	o.Present = true
	if bytes.Equal(data, []byte("null")) {
		o.Valid = false
		o.Value = ""
		return nil
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	o.Valid = true
	o.Value = value
	return nil
}

// ListMeta is pagination metadata.
type ListMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ListResult is the service result for listing complaints.
type ListResult struct {
	Items []Complaint
	Meta  ListMeta
}

// Complaint is the API-safe complaint representation.
type Complaint struct {
	ID           string              `json:"id"`
	Ref          string              `json:"ref"`
	ResidentID   *string             `json:"resident_id"`
	ResidentName string              `json:"resident_name"`
	Phone        string              `json:"phone"`
	Block        string              `json:"block"`
	Category     string              `json:"category"`
	Urgency      db.ComplaintUrgency `json:"urgency"`
	Status       db.ComplaintStatus  `json:"status"`
	Message      string              `json:"message"`
	ReceivedAt   string              `json:"received_at"`
	UpdatedAt    string              `json:"updated_at"`
}
