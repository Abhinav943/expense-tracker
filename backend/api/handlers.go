package api

import (
	"backend/models"
	"backend/storage"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type API struct {
	storage *storage.Storage
}

func NewAPI(storage *storage.Storage) *API {
	return &API{storage: storage}
}

func (api *API) createTransactionHandler(w http.ResponseWriter, r *http.Request) {
	var t models.Transaction

	err := json.NewDecoder(r.Body).Decode(&t)

	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	err = models.ValidateTransaction(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = api.storage.CreateTransaction(r.Context(), &t)

	if err != nil {
		http.Error(w, "Failed to save transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

func (api *API) getTransactionHandler(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "Invalid transaction id", http.StatusBadRequest)
		return
	}

	transaction, err := api.storage.GetTransaction(r.Context(), id)
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
	transactions, err := api.storage.GetAllTransactions(r.Context())
	if err != nil {
		http.Error(w, "Failed to retrieve transactions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

func (api *API) deleteTransactionHandler(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "Invalid transaction id", http.StatusBadRequest)
		return
	}

	err = api.storage.DeleteTransaction(r.Context(), id)
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

	err = api.storage.UpdateTransaction(r.Context(), &transaction, id)
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

func (api *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /transactions", api.createTransactionHandler)
	mux.HandleFunc("GET /transactions", api.getAllTransactionsHandler)
	mux.HandleFunc("GET /transactions/{id}", api.getTransactionHandler)
	mux.HandleFunc("DELETE /transactions/{id}", api.deleteTransactionHandler)
	mux.HandleFunc("PUT /transactions/{id}", api.updateTransactionHandler)
}