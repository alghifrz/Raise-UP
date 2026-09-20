package activity

// CreateRequest is the JSON body for POST /api/v1/activities.
// Reminder operational fields are intentionally omitted (server-managed).
type CreateRequest struct {
	Name               string `json:"name"`
	Description        string `json:"description"`
	Date               string `json:"date"`
	ReminderDaysBefore *int32 `json:"reminder_days_before"`
	ReminderTime       string `json:"reminder_time"`
	ReminderMessage    string `json:"reminder_message"`
}

// UpdateRequest is the JSON body for PATCH /api/v1/activities/:id.
// Reminder operational fields are intentionally omitted (server-managed).
type UpdateRequest struct {
	Name               *string `json:"name"`
	Description        *string `json:"description"`
	Date               *string `json:"date"`
	ReminderDaysBefore *int32  `json:"reminder_days_before"`
	ReminderTime       *string `json:"reminder_time"`
	ReminderMessage    *string `json:"reminder_message"`
}

// ListMeta is pagination metadata.
type ListMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ListResult is the service result for listing activities.
type ListResult struct {
	Items []Activity
	Meta  ListMeta
}

// Activity is the API-safe activity representation including reminder config/state.
type Activity struct {
	ID                      string  `json:"id"`
	Name                    string  `json:"name"`
	Description             string  `json:"description"`
	Date                    string  `json:"date"`
	ReminderDaysBefore      int32   `json:"reminder_days_before"`
	ReminderTime            *string `json:"reminder_time"`
	ReminderMessage         string  `json:"reminder_message"`
	ReminderScheduledAt     *string `json:"reminder_scheduled_at"`
	ReminderScheduleVersion int32   `json:"reminder_schedule_version"`
	ReminderN8nExecutionID  *string `json:"reminder_n8n_execution_id"`
	ReminderSentAt          *string `json:"reminder_sent_at"`
	CreatedAt               string  `json:"created_at"`
	UpdatedAt               string  `json:"updated_at"`
}
