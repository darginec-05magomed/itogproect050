package api

import (
	"fmt"
	"go1f/pkg/dateutils"
	"net/http"
	"strings"
	"time"
)

func afterNow(now, t time.Time) bool {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return today.After(t)
}

func NextDayHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Debug: now=%q, date=%q, repeat=%q\n",
		r.URL.Query().Get("now"),
		r.URL.Query().Get("date"),
		r.URL.Query().Get("repeat"))

	nowVol := strings.TrimSpace(r.URL.Query().Get("now"))
	dateVol := strings.TrimSpace(r.URL.Query().Get("date"))
	repeat := strings.TrimSpace(r.URL.Query().Get("repeat"))

	now, err := time.Parse("20060102", nowVol)
	if err != nil {
		writeJSON(w, map[string]string{"error": "ошибка формата now"}, http.StatusBadRequest)
		return
	}

	result, err := dateutils.NextDate(now, dateVol, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprint(w, result)
}
