package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jonathanface/storm-reporter/API/middleware"
)

func GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dao := middleware.GetDAO(r.Context())

	// Get the date from query parameters or default to today
	dateParam := r.URL.Query().Get("date")
	var unixTimestamp int64
	var err error

	if dateParam == "" {
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		unixTimestamp = today.Unix()
	} else {
		unixTimestamp, err = strconv.ParseInt(dateParam, 10, 64)
		if err != nil {
			http.Error(w, "Invalid 'date' query parameter", http.StatusBadRequest)
			return
		}
	}
	fmt.Println("time", unixTimestamp)

	startOfDay := unixTimestamp
	endOfDay := unixTimestamp + 86400 - 1

	reports, err := dao.GetStormReports(startOfDay, endOfDay)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve storm reports: %v", err), http.StatusInternalServerError)
		return
	}

	if len(reports) == 0 {
		http.Error(w, "No storm reports found for the given date", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(reports)
}
