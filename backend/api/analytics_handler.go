package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

const maxAnalyticsRangeDays = 366

var indiaTimeZone = time.FixedZone("Asia/Kolkata", 5*60*60+30*60)

func parseAnalyticsDateRange(r *http.Request) (time.Time, time.Time, error) {
	startDateValue := r.URL.Query().Get("start_date")
	if startDateValue == "" {
		return time.Time{}, time.Time{}, errors.New("start_date is required.")
	}
	startDate, err := time.ParseInLocation("2006-01-02", startDateValue, indiaTimeZone)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("Invalid start date format")
	}
	endDateValue := r.URL.Query().Get("end_date")
	if endDateValue == "" {
		return time.Time{}, time.Time{}, errors.New("end_date is required.")
	}
	endDate, err := time.ParseInLocation("2006-01-02", endDateValue, indiaTimeZone)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("Invalid end date format")
	}

	if endDate.Before(startDate) {
		return time.Time{}, time.Time{}, errors.New("End date cannot be before start date")
	}
	if endDate.After(startDate.AddDate(0, 0, maxAnalyticsRangeDays-1)) {
		return time.Time{}, time.Time{}, errors.New("Date range cannot exceed 366 days")
	}
	return startDate, endDate, nil
}

func (api *API) getSummaryHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}
	startDate, endDate, err := parseAnalyticsDateRange(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	summary, err := api.storage.GetSummary(r.Context(), userID, startDate, endDate)
	if err != nil {
		http.Error(w, "Failed to retrieve summary", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func (api *API) getDailyUpdatesHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}
	startDate, endDate, err := parseAnalyticsDateRange(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dailyUpdates, err := api.storage.GetDailyUpdates(r.Context(), userID, startDate, endDate)
	if err != nil {
		http.Error(w, "Failed to retrieve daily updates", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dailyUpdates)
}
