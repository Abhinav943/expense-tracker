package api

import (
	"backend/models"
	"backend/storage"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const maxNoteFilterLength = 100

func getUserIDFromContext(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value(userIDKey).(int)
	return userID, ok
}

func parseTransactionFilter(r *http.Request) (models.TransactionFilter, error) {
	query := r.URL.Query()

	validFilters := map[string]bool{
		"kind":        true,
		"category_id": true,
		"min_amount":  true,
		"max_amount":  true,
		"from":        true,
		"to":          true,
		"note":        true,
	}

	for key := range query {
		if !validFilters[key] {
			return models.TransactionFilter{}, fmt.Errorf("Unknown filter '%s'. Valid filters are: kind, category_id, min_amount, max_amount, from, to, note", key)
		}
	}

	var filter models.TransactionFilter

	if value := query.Get("kind"); value != "" {
		filter.Kind = value
	}

	if value := query.Get("category_id"); value != "" {
		id, err := strconv.Atoi(value)
		if err != nil || id < 1 {
			return models.TransactionFilter{}, errors.New("Invalid category_id: must be a positive integer")
		}
		filter.CategoryID = id
	}

	if value := query.Get("min_amount"); value != "" {
		amount, err := strconv.Atoi(value)
		if err != nil || amount < 0 {
			return models.TransactionFilter{}, errors.New("Invalid min_amount: must be a non-negative integer")
		}
		filter.MinAmount = &amount
	}

	if value := query.Get("max_amount"); value != "" {
		amount, err := strconv.Atoi(value)
		if err != nil || amount < 0 {
			return models.TransactionFilter{}, errors.New("Invalid max_amount: must be a non-negative integer")
		}
		filter.MaxAmount = &amount
	}

	if v := query.Get("from"); v != "" {
		t, err := time.ParseInLocation("2006-01-02", v, indiaTimeZone)
		if err != nil {
			return models.TransactionFilter{}, errors.New("Invalid from: must be a date in YYYY-MM-DD format")
		}
		filter.From = &t
	}

	if v := query.Get("to"); v != "" {
		t, err := time.ParseInLocation("2006-01-02", v, indiaTimeZone)
		if err != nil {
			return models.TransactionFilter{}, errors.New("Invalid to: must be a date in YYYY-MM-DD format")
		}
		filter.To = &t
	}

	if v := query.Get("note"); v != "" {
		if len(v) > maxNoteFilterLength {
			return models.TransactionFilter{}, errors.New("Invalid note: filter text is too long")
		}
		filter.Note = v
	}

	if err := filter.Validate(); err != nil {
		return models.TransactionFilter{}, err
	}
	return filter, nil
}

func (api *API) createTransactionHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}
	var t models.Transaction

	err := json.NewDecoder(r.Body).Decode(&t)

	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	t.UserID = userID

	err = models.ValidateTransaction(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = api.storage.CreateTransaction(r.Context(), &t)

	if errors.Is(err, storage.ErrInvalidCategory) {
		http.Error(w, "Invalid category", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "Failed to save transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

func (api *API) getTransactionHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	idString := r.PathValue("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "Invalid transaction id", http.StatusBadRequest)
		return
	}

	transaction, err := api.storage.GetTransaction(r.Context(), userID, id)
	if errors.Is(err, storage.ErrNotFound) {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to retrieve transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transaction)
}

func (api *API) getAllTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	filter, err := parseTransactionFilter(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	transactions, err := api.storage.GetAllTransactions(r.Context(), userID, filter)
	if err != nil {
		http.Error(w, "Failed to retrieve transactions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

func (api *API) deleteTransactionHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}
	idString := r.PathValue("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "Invalid transaction id", http.StatusBadRequest)
		return
	}

	err = api.storage.DeleteTransaction(r.Context(), userID, id)
	if errors.Is(err, storage.ErrNotFound) {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to delete transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (api *API) updateTransactionHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	idString := r.PathValue("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "Invalid transaction id", http.StatusBadRequest)
		return
	}

	var transaction models.Transaction
	err = json.NewDecoder(r.Body).Decode(&transaction)
	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	err = models.ValidateTransaction(&transaction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = api.storage.UpdateTransaction(r.Context(), &transaction, userID, id)
	if errors.Is(err, storage.ErrNotFound) {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	} else if errors.Is(err, storage.ErrInvalidCategory) {
		http.Error(w, "Invalid category", http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, "Failed to update transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transaction)
}
