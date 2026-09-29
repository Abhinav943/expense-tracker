package api

import (
	"backend/models"
	"backend/storage"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var dummyHash []byte

func init() {
	dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy"), bcrypt.DefaultCost)
}

func (api *API) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var req models.Credentials
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	err = models.ValidateCredentials(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	PasswordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user := models.User{
		Email:        req.Email,
		PasswordHash: string(PasswordHash),
	}

	err = api.storage.CreateUser(r.Context(), &user)
	if errors.Is(err, storage.ErrEmailTaken) {
		http.Error(w, "Email already taken", http.StatusConflict)
		return
	} else if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (api *API) loginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.Credentials
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	err = models.ValidateCredentials(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := api.storage.GetUserByEmail(r.Context(), req.Email)
	if errors.Is(err, storage.ErrNotFound) {
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(req.Password))
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Failed to retrieve user", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	secretKey := []byte(api.jwtSecret)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	})
}
