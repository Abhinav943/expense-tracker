package api

import (
	"backend/models"
	"backend/storage"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
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

	category.Name = strings.TrimSpace(category.Name)
	if category.Name == "" {
		http.Error(w, "Category name is required", http.StatusBadRequest)
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
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(category)
}

func (api *API) getAllCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	categories, err := api.storage.GetAllCategories(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to retrieve categories", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

func (api *API) updateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	categoryID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || categoryID <= 0 {
		http.Error(w, "Invalid category id", http.StatusBadRequest)
		return
	}

	var request struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	request.Name = strings.TrimSpace(request.Name)
	if request.Name == "" {
		http.Error(w, "Category name is required", http.StatusBadRequest)
		return
	}

	err = api.storage.UpdateCategory(r.Context(), categoryID, userID, request.Name)
	if errors.Is(err, storage.ErrCategoryAlreadyExists) {
		http.Error(w, "Category already exists", http.StatusConflict)
		return
	}
	if errors.Is(err, storage.ErrNotFound) {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Failed to update category", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (api *API) deleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	categoryID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || categoryID <= 0 {
		http.Error(w, "Invalid category id", http.StatusBadRequest)
		return
	}

	err = api.storage.DeleteCategory(r.Context(), userID, categoryID)
	if errors.Is(err, storage.ErrNotFound) {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Failed to delete category", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
