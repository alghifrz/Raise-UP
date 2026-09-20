package dues

import (
	"bytes"
	"encoding/json"
	"fmt"
)

const (
	PaymentStatusPaid   = "PAID"
	PaymentStatusUnpaid = "UNPAID"
)

// CreatePeriodRequest is the JSON body for POST /api/v1/dues/periods.
type CreatePeriodRequest struct {
	Year   int32  `json:"year"`
	Month  int32  `json:"month"`
	Half   int32  `json:"half"`
	Amount Amount `json:"amount"`
}

// UpdatePeriodRequest is the JSON body for PATCH /api/v1/dues/periods/:id.
type UpdatePeriodRequest struct {
	Year   *int32  `json:"year"`
	Month  *int32  `json:"month"`
	Half   *int32  `json:"half"`
	Amount *Amount `json:"amount"`
}

// CreatePaymentRequest is the JSON body for POST /api/v1/dues/payments.
type CreatePaymentRequest struct {
	PeriodID   string `json:"period_id"`
	ResidentID string `json:"resident_id"`
	Amount     Amount `json:"amount"`
	PaidAt     string `json:"paid_at"`
}

// UpdatePaymentRequest is the JSON body for PATCH /api/v1/dues/payments/:id.
type UpdatePaymentRequest struct {
	PeriodID   *string `json:"period_id"`
	ResidentID *string `json:"resident_id"`
	Amount     *Amount `json:"amount"`
	PaidAt     *string `json:"paid_at"`
}

// Amount is an IDR integer amount. Decimal JSON values are rejected.
type Amount int64

// UnmarshalJSON rejects non-integer JSON numbers.
func (a *Amount) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if bytes.ContainsAny(data, ".eE") {
		return fmt.Errorf("amount must be an integer")
	}
	var value int64
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("amount must be an integer")
	}
	*a = Amount(value)
	return nil
}

// ListMeta is pagination metadata.
type ListMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// PeriodListResult is the service result for listing periods.
type PeriodListResult struct {
	Items []Period
	Meta  ListMeta
}

// PaymentListResult is the service result for listing payments.
type PaymentListResult struct {
	Items []Payment
	Meta  ListMeta
}

// StatusListResult is the service result for resident payment status.
type StatusListResult struct {
	Items []ResidentPaymentStatus
	Meta  ListMeta
}

// Period is the API-safe dues period representation.
type Period struct {
	ID        string `json:"id"`
	Year      int32  `json:"year"`
	Month     int32  `json:"month"`
	Half      int32  `json:"half"`
	Amount    int64  `json:"amount"`
	CreatedAt string `json:"created_at"`
}

// Payment is the API-safe dues payment representation.
type Payment struct {
	ID           string `json:"id"`
	PeriodID     string `json:"period_id"`
	ResidentID   string `json:"resident_id"`
	ResidentName string `json:"resident_name"`
	Amount       int64  `json:"amount"`
	PaidAt       string `json:"paid_at"`
	CreatedAt    string `json:"created_at"`
}

// PeriodSummary is aggregated collection status for a period.
type PeriodSummary struct {
	PeriodID         string `json:"period_id"`
	Year             int32  `json:"year"`
	Month            int32  `json:"month"`
	Half             int32  `json:"half"`
	ExpectedAmount   int64  `json:"expected_amount"`
	ResidentCount    int64  `json:"resident_count"`
	PaidCount        int64  `json:"paid_count"`
	UnpaidCount      int64  `json:"unpaid_count"`
	ExpectedTotal    int64  `json:"expected_total"`
	CollectedTotal   int64  `json:"collected_total"`
	OutstandingTotal int64  `json:"outstanding_total"`
}

// ResidentPaymentStatus is a derived PAID/UNPAID row for a resident in a period.
type ResidentPaymentStatus struct {
	ResidentID   string  `json:"resident_id"`
	ResidentName string  `json:"resident_name"`
	Phone        string  `json:"phone"`
	Status       string  `json:"status"`
	PaymentID    *string `json:"payment_id"`
	Amount       *int64  `json:"amount"`
	PaidAt       *string `json:"paid_at"`
}
