package db

import (
	"database/sql"
	"fmt"
	"strconv"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func AddTask(task *Task) (int64, error) {
	if DB() == nil {
		if err := Init(DBFile); err != nil {
			return 0, err
		}
	}
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := DB().Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func ListTasks(search string) ([]Task, error) {
	if DB() == nil {
		if err := Init(DBFile); err != nil {
			return nil, err
		}
	}
	var rows *sql.Rows
	var err error
	if search == "" {
		rows, err = DB().Query(`SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC`)
	} else {
		like := "%" + search + "%"
		rows, err = DB().Query(`SELECT id, date, title, comment, repeat FROM scheduler WHERE date LIKE ? OR title LIKE ? OR comment LIKE ? ORDER BY date ASC`, like, like, like)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Task
	for rows.Next() {
		var id int64
		var date, title, comment, repeat string
		if err := rows.Scan(&id, &date, &title, &comment, &repeat); err != nil {
			return nil, err
		}
		out = append(out, Task{ID: strconv.FormatInt(id, 10), Date: date, Title: title, Comment: comment, Repeat: repeat})
	}
	return out, nil
}
func GetTask(idStr string) (*Task, error) {
	if DB() == nil {
		if err := Init(DBFile); err != nil {
			return nil, err
		}
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid task id: %s", idStr)
	}
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`

	var (
		ID      int64
		date    string
		title   string
		comment string
		repeat  string
	)

	err = DB().QueryRow(query, id).Scan(&ID, &date, &title, &comment, &repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task with ID %s not found", idStr)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	task := Task{
		ID:      strconv.FormatInt(ID, 10),
		Date:    date,
		Title:   title,
		Comment: comment,
		Repeat:  repeat,
	}

	return &task, nil
}
func UpdateTask(idStr string, task *Task) error {
	if DB() == nil {
		if err := Init(DBFile); err != nil {
			return err
		}
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid task ID: '%s'", idStr)
	}

	query := `UPDATE scheduler SET date=?, title=?, comment=?, repeat=? WHERE id=?`
	res, err := DB().Exec(query, task.Date, task.Title, task.Comment, task.Repeat, id) // ← id — int64
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("task with ID %s not found", idStr)
	}
	return nil
}

func DeleteTask(idStr string) error {
	if DB() == nil {
		if err := Init(DBFile); err != nil {
			return err
		}
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid task ID: '%s'", idStr)
	}

	query := `DELETE FROM scheduler WHERE id=?`
	res, err := DB().Exec(query, id)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("task with ID %s not found", idStr)
	}
	return nil
}

func UpdateDate(idStr string, newDate string) error {
	if DB() == nil {
		if err := Init(DBFile); err != nil {
			return err
		}
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid task ID: '%s'", idStr)
	}
	query := `UPDATE scheduler SET date=? WHERE id=?`
	res, err := DB().Exec(query, newDate, id)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("task with ID %s not found", idStr)
	}
	return nil
}
