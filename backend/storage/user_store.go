package storage

import (
	"backend/models"
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

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
