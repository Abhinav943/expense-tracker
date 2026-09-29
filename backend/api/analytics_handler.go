package api

import (
	"encoding/json"
	"net/http"
	"time"
)

func (api *API) getSummaryHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	start_date, err := time.Parse("2006-01-02", r.URL.Query().Get("start_date"))
	if err != nil {
		http.Error(w, "Invalid start date format", http.StatusBadRequest)
		return
	}
	end_date, err := time.Parse("2006-01-02", r.URL.Query().Get("end_date"))
	if err != nil {
		http.Error(w, "Invalid end date format", http.StatusBadRequest)
		return
	}

	if end_date.Before(start_date) {
		http.Error(w, "End date cannot be before start date", http.StatusBadRequest)
		return
	}

	summary, err := api.storage.GetSummary(r.Context(), userID, start_date, end_date)
	if err != nil {
		http.Error(w, "Failed to retrieve summary", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}
