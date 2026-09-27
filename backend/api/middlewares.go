package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDKey contextKey = "user_id"

var unexpectedError = errors.New("unexpected signing method")

func (api *API) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		tokenString := strings.TrimPrefix(token, "Bearer ")
		if tokenString == "" {
			http.Error(w, "Missing or invalid token", http.StatusUnauthorized)
			return
		}

		parseToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, unexpectedError
			}
			return api.jwtSecret, nil
		})

		if err != nil || !parseToken.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if claims, ok := parseToken.Claims.(jwt.MapClaims); ok && parseToken.Valid {
			userIDFloat, ok := claims[string(userIDKey)].(float64)
			if !ok {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}
			userID := int(userIDFloat)

			r = r.WithContext(context.WithValue(r.Context(), userIDKey, userID))
		} else {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
