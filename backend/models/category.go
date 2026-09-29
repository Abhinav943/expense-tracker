package models

import "time"

type Category struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	UserID    int        `json:"-"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}
