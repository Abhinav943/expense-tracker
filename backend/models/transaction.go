package models

import (
	"errors"
	"time"
)

var ErrInvalidAmount = errors.New("Invalid Amount: must be above or equal to zero")
var ErrInvalidKind = errors.New("Invalid kind: must be 'income' or 'expense'")

type Transaction struct {
	ID        int       `json:"id"`
	Amount    int       `json:"amount"`
	Kind      string    `json:"kind"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

func ValidateTransaction(transaction *Transaction) error {
	if transaction.Amount < 0 {
		return ErrInvalidAmount
	}

	if transaction.Kind != "income" && transaction.Kind != "expense" {
		return ErrInvalidKind
	}
	return nil
}
