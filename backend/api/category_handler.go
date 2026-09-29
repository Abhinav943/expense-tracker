package api

import (
	"backend/models"
	"backend/storage"
	"encoding/json"
	"errors"
	"net/http"
)


func (api *API) createCategoryHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	var category models.Category

	err := json.NewDecoder(r.Body).Decode(&category)

	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	category.UserID = userID

	err = api.storage.CreateCategory(r.Context(), &category)

	if errors.Is(err, storage.ErrCategoryAlreadyExists) {
		http.Error(w, "Category already exists", http.StatusConflict)
		return
	} else if err != nil {
		http.Error(w, "Failed to save category", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}
