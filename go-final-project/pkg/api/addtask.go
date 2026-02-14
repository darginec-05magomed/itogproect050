package api

import (
	"encoding/json"
	"go1f/pkg/dateutils"
	"go1f/pkg/db"
	"log"
	"net/http"
	"strings"
	"time"
)

func checkDate(task *db.Task) error {
	now := time.Now()

	if strings.TrimSpace(task.Date) == "" || strings.TrimSpace(task.Date) == "today" {
		task.Date = now.Format("20060102")
	}

	t, err := time.Parse("20060102", task.Date)
	if err != nil {
		return err
	}

	if task.Repeat != "" {
		next, err := dateutils.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return err
		}
		if afterNow(now, t) {
			task.Date = next
		}
	} else if afterNow(now, t) {
		task.Date = now.Format("20060102")
	}

	return nil
}

func writeJSON(w http.ResponseWriter, data any, code int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)
	if data == nil {
		data = map[string]interface{}{}
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to encode json: %v", err)
	}
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			writeJSON(w, map[string]string{"error": "internal server error"},
				http.StatusInternalServerError)
		}
	}()
	if r.Method != http.MethodPost {
		writeJSON(w, map[string]string{"error": "method not allowed"}, http.StatusMethodNotAllowed)
		return
	}

	var task db.Task

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJSON(w, map[string]string{"error": "invalid json"}, http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		writeJSON(w, map[string]string{"error": "title is required"}, http.StatusBadRequest)
		return
	}

	if err := checkDate(&task); err != nil {
		writeJSON(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return
	}

	id, err := db.AddTask(&task)
	if err != nil {
		writeJSON(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"id": id}, http.StatusOK)

	return
}
