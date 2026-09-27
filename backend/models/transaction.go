package models

import (
	"errors"
	"time"
)

type Transaction struct {
	ID        int       `json:"id"`
	Amount    int       `json:"amount"`
	Kind      string    `json:"kind"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
	UserID  int       `json:"-"`
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
