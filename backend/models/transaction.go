package models

import (
	"errors"
	"time"
)

type Transaction struct {
	ID           int       `json:"id"`
	Amount       int       `json:"amount"`
	Kind         string    `json:"kind"`
	Note         string    `json:"note"`
	CreatedAt    time.Time `json:"created_at"`
	UserID       int       `json:"-"`
	CategoryID   *int      `json:"category_id,omitempty"`
	CategoryName *string   `json:"category_name,omitempty"`
}

type TransactionFilter struct {
	Kind       string
	CategoryID int
	MinAmount  *int
	MaxAmount  *int
	From       *time.Time
	To         *time.Time
	Note       string
}

func ValidateTransaction(transaction *Transaction) error {
	if transaction.Amount < 0 {
		return errors.New("Invalid Amount: must be above or equal to zero")
	}

	if transaction.Kind != "income" && transaction.Kind != "expense" {
		return errors.New("Invalid kind: must be 'income' or 'expense'")
	}
	return nil
}

func (filter *TransactionFilter) Validate() error {
	if filter.Kind != "" && filter.Kind != "income" && filter.Kind != "expense" {
		return errors.New("Invalid kind: must be 'income' or 'expense'")
	}

	if filter.MinAmount != nil && *filter.MinAmount < 0 {
		return errors.New("Invalid MinAmount: must be above or equal to zero")
	}

	if filter.MaxAmount != nil && *filter.MaxAmount < 0 {
		return errors.New("Invalid MaxAmount: must be above or equal to zero")
	}

	if filter.MinAmount != nil && filter.MaxAmount != nil && *filter.MinAmount > *filter.MaxAmount {
		return errors.New("Invalid amount range: min_amount must be less than or equal to max_amount")
	}

	if filter.From != nil && filter.To != nil && filter.To.Before(*filter.From) {
		return errors.New("Invalid date range: FromDate must be before ToDate")
	}
	return nil
}
