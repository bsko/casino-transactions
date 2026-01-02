package http

import "time"

type TransactionSearchRequest struct {
	UserID          *string    `json:"user_id,omitempty"`
	TransactionType *string    `json:"transaction_type,omitempty"`
	AmountFrom      *float64   `json:"amount_from,omitempty"`
	AmountTo        *float64   `json:"amount_to,omitempty"`
	CreatedFrom     *time.Time `json:"created_from,omitempty"`
	CreatedTo       *time.Time `json:"created_to,omitempty"`
	Limit           int        `json:"limit,omitempty"`
	Offset          int        `json:"offset,omitempty"`
}

type TransactionDTO struct {
	UserID          string    `json:"user_id"`
	TransactionType string    `json:"transaction_type"`
	Amount          float64   `json:"amount"`
	Timestamp       time.Time `json:"timestamp"`
}

type TransactionListResponse struct {
	Total        int              `json:"total"`
	Limit        int              `json:"limit"`
	Offset       int              `json:"offset"`
	Transactions []TransactionDTO `json:"transactions"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
