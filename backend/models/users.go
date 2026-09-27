package models

import (
	"errors"
	"strings"
	"time"
)

type User struct {
	ID            int       `json:"id"`
	Email         string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt     time.Time `json:"created_at"`
}

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func ValidateCredentials(c *Credentials) error {
	if c == nil {
		return errors.New("credentials are required")
	}
	if !strings.Contains(c.Email, "@") {
		return errors.New("email must contain @")
	}
	if len(c.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	if len([]byte(c.Password)) > 72 {
		return errors.New("password must be 72 bytes or less")
	}
	return nil
}