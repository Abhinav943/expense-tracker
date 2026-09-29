package storage

import (
	"backend/models"
	"context"
	"errors"

	"github.com/lib/pq"
)

var ErrCategoryAlreadyExists = errors.New("Category already exists")

func (s *Storage) CreateCategory(ctx context.Context, c *models.Category) error {
	const createCategoryQuery = "INSERT INTO categories (name,user_id) VALUES ($1, $2) RETURNING id,created_at"

	err := s.db.QueryRowContext(ctx, createCategoryQuery, c.Name, c.UserID).Scan(&c.ID, &c.CreatedAt)

	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return ErrCategoryAlreadyExists
	}
	return err
}

func (s *Storage) DeleteCategory(ctx context.Context, userID int, categoryID int) error {
	const deleteCategoryQuery = "DELETE FROM categories WHERE id = $1 AND user_id = $2"
	result, err := s.db.ExecContext(ctx, deleteCategoryQuery, categoryID, userID)
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

func (s *Storage) GetAllCategories(ctx context.Context, userID int) ([]models.Category, error) {
	const getAllCategoriesQuery = "SELECT id, name, created_at FROM categories WHERE user_id = $1 ORDER BY name ASC, created_at DESC, id DESC"
	rows, err := s.db.QueryContext(ctx, getAllCategoriesQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := []models.Category{}
	for rows.Next() {
		var c models.Category
		err := rows.Scan(&c.ID, &c.Name, &c.CreatedAt)
		if err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (s *Storage) UpdateCategory(ctx context.Context, categoryID int, userID int, newName string) error {
	const updateCategoryQuery = "UPDATE categories SET name = $1, updated_at = NOW() WHERE id = $2 AND user_id = $3"
	result, err := s.db.ExecContext(ctx, updateCategoryQuery, newName, categoryID, userID)
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return ErrCategoryAlreadyExists
	}

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
