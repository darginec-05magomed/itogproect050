package api

import (
	"encoding/json"
	"go1f/pkg/dateutils"
	"go1f/pkg/db"
	"net/http"
	"os"
	"strings"
	"time"
)

func Init() {
	http.HandleFunc("/api/nextdate", NextDayHandler)
	http.HandleFunc("/api/task", auth(taskHandler))
	http.HandleFunc("/api/tasks", auth(tasksHandler))
	http.HandleFunc("/api/task/done", auth(taskDone))
	http.HandleFunc("/api/signin", signInHandler)
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodGet:
		taskGet(w, r)
	case http.MethodPut:
		taskPut(w, r)
	case http.MethodDelete:
		taskDelete(w, r)
	}
}

func auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pass := os.Getenv("TODO_PASSWORD")
		if pass == "" {
			next(w, r)
			return
		}

		cookie, _ := r.Cookie("token")
		if cookie == nil || cookie.Value != pass {
			writeJSON(w, map[string]string{"error": "Auth required"}, http.StatusUnauthorized)
			return
		}

		next(w, r)
	})
}

func signInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, map[string]string{"error": "method not allowed"}, http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&input)

	envPass := os.Getenv("TODO_PASSWORD")
	if input.Password != envPass {
		writeJSON(w, map[string]string{"error": "Неверный пароль"}, http.StatusUnauthorized)
		return
	}

	writeJSON(w, map[string]string{"token": envPass}, http.StatusOK)
}

func taskDone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, map[string]string{"error": "method not allowed"}, http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, map[string]string{"error": "id is required"}, http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJSON(w, map[string]string{"error": "Задача не найдена"}, http.StatusNotFound)
		return
	}

	if task.Repeat == "" {
		err := db.DeleteTask(id)
		if err != nil {
			writeJSON(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
			return
		}
	} else {

		now := time.Now()
		nextDate, err := dateutils.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			writeJSON(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
			return
		}

		err = db.UpdateDate(id, nextDate)
		if err != nil {
			writeJSON(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, map[string]interface{}{}, http.StatusOK)
}

func taskDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeJSON(w, map[string]string{"error": "method not allowed"}, http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, map[string]string{"error": "id is required"}, http.StatusBadRequest)
		return
	}

	_, err := db.GetTask(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeJSON(w, map[string]string{"error": "Задача не найдена"}, http.StatusNotFound)
		} else {
			writeJSON(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		}
		return
	}

	err = db.DeleteTask(id)
	if err != nil {
		writeJSON(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{}, http.StatusOK)
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, map[string]string{"error": "method not allowed"}, http.StatusMethodNotAllowed)
		return
	}
	search := r.URL.Query().Get("search")
	tasks, err := db.ListTasks(search)
	if err != nil {
		writeJSON(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}
	out := make([]map[string]string, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, map[string]string{
			"id":      t.ID,
			"date":    t.Date,
			"title":   t.Title,
			"comment": t.Comment,
			"repeat":  t.Repeat,
		})
	}
	writeJSON(w, map[string]any{"tasks": out}, http.StatusOK)
}
func taskGet(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, map[string]string{"error": "id is required"}, http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "task with ID "+id+" not found" || strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		writeJSON(w, map[string]string{"error": err.Error()}, status)
		return
	}

	writeJSON(w, task, http.StatusOK)
}
func taskPut(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ID      string `json:"id"`
		Date    string `json:"date"`
		Title   string `json:"title"`
		Comment string `json:"comment"`
		Repeat  string `json:"repeat"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, map[string]string{"error": "invalid json"}, http.StatusBadRequest)
		return
	}

	if input.ID == "" {
		writeJSON(w, map[string]string{"error": "id is required"}, http.StatusBadRequest)
		return
	}

	if input.Title == "" {
		writeJSON(w, map[string]string{"error": "title is required"}, http.StatusBadRequest)
		return
	}

	if input.Date == "" {
		writeJSON(w, map[string]string{"error": "date is required"}, http.StatusBadRequest)
		return
	}

	if len(input.Date) != 8 {
		writeJSON(w, map[string]string{"error": "invalid date format"}, http.StatusBadRequest)
		return
	}

	_, err := time.Parse("20060102", input.Date)
	if err != nil {
		writeJSON(w, map[string]string{"error": "invalid date format"}, http.StatusBadRequest)
		return
	}

	if input.Repeat != "" {
		if input.Repeat != "y" &&
			!strings.HasPrefix(input.Repeat, "d ") &&
			!strings.HasPrefix(input.Repeat, "w ") &&
			!strings.HasPrefix(input.Repeat, "m ") {
			writeJSON(w, map[string]string{"error": "invalid repeat format"}, http.StatusBadRequest)
			return
		}
	}

	task := db.Task{
		Date:    input.Date,
		Title:   input.Title,
		Comment: input.Comment,
		Repeat:  input.Repeat,
	}

	if err := db.UpdateTask(input.ID, &task); err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
			writeJSON(w, map[string]string{"error": "Задача не найдена"}, status)
		} else {
			writeJSON(w, map[string]string{"error": err.Error()}, status)
		}
		return
	}

	writeJSON(w, map[string]interface{}{}, http.StatusOK)
}
