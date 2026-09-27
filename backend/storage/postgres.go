package storage

import (
	"backend/models"
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

var ErrNotFound = errors.New("not found")
var ErrEmailTaken = errors.New("email already taken")

type Storage struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) CreateTransaction(ctx context.Context, t *models.Transaction) error {
	const insertQuery = "INSERT INTO transactions (amount, kind, note, user_id) VALUES ($1, $2, $3, $4) RETURNING id, created_at"
	err := s.db.QueryRowContext(ctx,
		insertQuery,
		t.Amount,
		t.Kind,
		t.Note,
		t.UserID,
	).Scan(&t.ID, &t.CreatedAt)

	return err
}

func (s *Storage) GetTransaction(ctx context.Context, userID int, id int) (models.Transaction, error) {
	var transaction models.Transaction
	const getQuery = "SELECT id, amount, kind, note, created_at FROM transactions WHERE user_id = $1 AND id = $2"
	err := s.db.QueryRowContext(ctx, getQuery, userID, id).Scan(
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

func (s *Storage) GetAllTransactions(ctx context.Context, userID int) ([]models.Transaction, error) {
	const getAllQuery = "SELECT id, amount, kind, note, created_at FROM transactions WHERE user_id = $1 ORDER BY created_at DESC, id DESC"
	rows, err := s.db.QueryContext(ctx, getAllQuery, userID)
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

func (s *Storage) CreateUser(ctx context.Context, u *models.User) error {
	const insertQuery = "INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, created_at"
	err := s.db.QueryRowContext(ctx, insertQuery, u.Email, u.PasswordHash).Scan(&u.ID, &u.CreatedAt)

	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return ErrEmailTaken
	}
	return err
}

func (s *Storage) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	const getQuery = "SELECT id, email, password_hash, created_at FROM users WHERE email = $1"
	err := s.db.QueryRowContext(ctx, getQuery, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return user, ErrNotFound
	}
	return user, err
}
