package api

import (
	"backend/models"
	"backend/storage"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)


func getUserIDFromContext(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value(userIDKey).(int)
	return userID, ok
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
	transactions, err := api.storage.GetAllTransactions(r.Context(), userID)
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
	} else if err != nil {
		http.Error(w, "Failed to update transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transaction)
}
