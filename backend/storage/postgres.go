package storage

import (
	"backend/models"
	"context"
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("not found")

type Storage struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) CreateTransaction(ctx context.Context, t *models.Transaction) error {
	const insertQuery = "INSERT INTO transactions (amount, kind, note) VALUES ($1, $2, $3) RETURNING id, created_at"
	err := s.db.QueryRowContext(ctx,
		insertQuery,
		t.Amount,
		t.Kind,
		t.Note,
	).Scan(&t.ID, &t.CreatedAt)

	return err
}

func (s *Storage) GetTransaction(ctx context.Context, id int) (models.Transaction, error) {
	var transaction models.Transaction
	const getQuery = "SELECT id, amount, kind, note, created_at FROM transactions WHERE id = $1"
	err := s.db.QueryRowContext(ctx, getQuery, id).Scan(
		&transaction.ID,
		&transaction.Amount,
		&transaction.Kind,
		&transaction.Note,
		&transaction.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return transaction, ErrNotFound
	}
	return transaction, err
}

func (s *Storage) GetAllTransactions(ctx context.Context) ([]models.Transaction, error) {
	const getAllQuery = "SELECT id, amount, kind, note, created_at FROM transactions ORDER BY created_at DESC, id DESC"
	rows, err := s.db.QueryContext(ctx, getAllQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := []models.Transaction{}
	for rows.Next() {
		var t models.Transaction
		err := rows.Scan(
			&t.ID,
			&t.Amount,
			&t.Kind,
			&t.Note,
			&t.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (s *Storage) DeleteTransaction(ctx context.Context, id int) error {
	const deleteQuery = "DELETE FROM transactions WHERE id = $1"
	result, err := s.db.ExecContext(ctx, deleteQuery, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Storage) UpdateTransaction(ctx context.Context, t *models.Transaction, id int) error {
	const updateQuery = "UPDATE transactions SET amount = $1, kind = $2, note = $3 WHERE id = $4 RETURNING id, created_at"
	err := s.db.QueryRowContext(ctx, updateQuery, t.Amount, t.Kind, t.Note, id).Scan(&t.ID, &t.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
