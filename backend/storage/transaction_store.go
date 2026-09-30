package storage

import (
	"backend/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

var ErrNotFound = errors.New("not found")
var ErrEmailTaken = errors.New("email already taken")
var ErrInvalidCategory = errors.New("invalid category")

func (s *Storage) CreateTransaction(ctx context.Context, t *models.Transaction) error {
	const insertQuery = `
		INSERT INTO transactions (amount, kind, note, user_id, category_id)
		SELECT $1, $2, $3, $4, $5
		WHERE $5::integer IS NULL OR EXISTS (
			SELECT 1 FROM categories WHERE id = $5 AND user_id = $4
		)
		RETURNING id, created_at`
	err := s.db.QueryRowContext(ctx,
		insertQuery,
		t.Amount,
		t.Kind,
		t.Note,
		t.UserID,
		t.CategoryID,
	).Scan(&t.ID, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalidCategory
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23503" {
		return ErrInvalidCategory
	}
	return err
}

func (s *Storage) GetTransaction(ctx context.Context, userID int, id int) (models.Transaction, error) {
	var transaction models.Transaction
	const getQuery = "SELECT t.id, t.amount, t.kind, t.note, t.created_at, t.category_id, c.name FROM transactions t LEFT JOIN categories c ON t.category_id = c.id  WHERE t.user_id = $1 AND t.id = $2"
	err := s.db.QueryRowContext(ctx, getQuery, userID, id).Scan(
		&transaction.ID,
		&transaction.Amount,
		&transaction.Kind,
		&transaction.Note,
		&transaction.CreatedAt,
		&transaction.CategoryID,
		&transaction.CategoryName,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return transaction, ErrNotFound
	}
	return transaction, err
}

func (s *Storage) GetAllTransactions(ctx context.Context, userID int, filter models.TransactionFilter) ([]models.Transaction, error) {
	conditions := []string{"user_id = $1"}
	args := []any{userID}

	if filter.Kind != "" {
		conditions = append(conditions, fmt.Sprintf("kind = $%d", len(args)+1))
		args = append(args, filter.Kind)
	}
	if filter.CategoryID != 0 {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", len(args)+1))
		args = append(args, filter.CategoryID)
	}
	if filter.MinAmount != nil {
		conditions = append(conditions, fmt.Sprintf("amount >= $%d", len(args)+1))
		args = append(args, *filter.MinAmount)
	}
	if filter.MaxAmount != nil {
		conditions = append(conditions, fmt.Sprintf("amount <= $%d", len(args)+1))
		args = append(args, *filter.MaxAmount)
	}
	if filter.From != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", len(args)+1))
		args = append(args, *filter.From)
	}
	if filter.To != nil {
		toExclusive := filter.To.AddDate(0, 0, 1)
		conditions = append(conditions, fmt.Sprintf("created_at < $%d", len(args)+1))
		args = append(args, toExclusive)
	}
	if filter.Note != "" {
		conditions = append(conditions, fmt.Sprintf("note ILIKE '%%' || $%d || '%%'", len(args)+1))
		args = append(args, filter.Note)
	}

	query := "SELECT id, amount, kind, note, created_at FROM transactions WHERE " +
		strings.Join(conditions, " AND ") +
		" ORDER BY created_at DESC, id DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := []models.Transaction{}
	for rows.Next() {
		var t models.Transaction
		err := rows.Scan(&t.ID, &t.Amount, &t.Kind, &t.Note, &t.CreatedAt)
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

func (s *Storage) DeleteTransaction(ctx context.Context, userID int, id int) error {
	const deleteQuery = "DELETE FROM transactions WHERE user_id = $1 AND id = $2"
	result, err := s.db.ExecContext(ctx, deleteQuery, userID, id)
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

func (s *Storage) UpdateTransaction(ctx context.Context, t *models.Transaction, userID int, id int) error {
	const updateQuery = "UPDATE transactions SET amount = $1, kind = $2, note = $3 WHERE user_id = $4 AND id = $5 RETURNING id, created_at"
	err := s.db.QueryRowContext(ctx, updateQuery, t.Amount, t.Kind, t.Note, userID, id).Scan(&t.ID, &t.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
