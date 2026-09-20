package finance

import (
	"bytes"
	"encoding/json"
	"fmt"

	db "github.com/diuk/raiseup/db/generated"
)

// CreateRequest is the JSON body for POST /api/v1/finance/transactions.
type CreateRequest struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Amount   Amount `json:"amount"`
	Category string `json:"category"`
	Note     string `json:"note"`
}

// UpdateRequest is the JSON body for PATCH /api/v1/finance/transactions/:id.
type UpdateRequest struct {
	Type     *string `json:"type"`
	Title    *string `json:"title"`
	Amount   *Amount `json:"amount"`
	Category *string `json:"category"`
	Note     *string `json:"note"`
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

// ListResult is the service result for listing transactions.
type ListResult struct {
	Items []Transaction
	Meta  ListMeta
}

// Transaction is the API-safe cash transaction representation.
type Transaction struct {
	ID        string                 `json:"id"`
	Type      db.CashTransactionType `json:"type"`
	Title     string                 `json:"title"`
	Amount    int64                  `json:"amount"`
	Category  string                 `json:"category"`
	Note      string                 `json:"note"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

// Summary is the aggregated finance summary.
type Summary struct {
	TotalIncome  int64 `json:"total_income"`
	TotalExpense int64 `json:"total_expense"`
	Balance      int64 `json:"balance"`
}
