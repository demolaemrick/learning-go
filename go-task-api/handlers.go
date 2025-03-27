package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// writeJSON sets the content type and writes JSON data to the response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, message string, statusCode int) {
	writeJSON(w, statusCode, ErrorResponse{Error: message})
}

func getTasks(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, title, description, completed FROM tasks")
	if err != nil {
		sendError(w, "Failed to fetch tasks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task

	for rows.Next() {
		var t Task
		err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Completed)

		if err != nil {
			sendError(w, "Failed to scan tasks", http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, t)
	}
	writeJSON(w, http.StatusOK, tasks)
}

func getTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		sendError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var t Task

	err = db.QueryRow("SELECT id, title, description, completed FROM tasks WHERE id = $1", id).
		Scan(&t.ID, &t.Title, &t.Description, &t.Completed)

	if err == sql.ErrNoRows {
		sendError(w, "Task not found", http.StatusNotFound)
		return
	} else if err != nil {
		sendError(w, "Failed to fetch task", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func createTask(w http.ResponseWriter, r *http.Request) {
	var newTask Task
	err := json.NewDecoder(r.Body).Decode(&newTask)

	if err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validation
	if newTask.Title == "" {
		sendError(w, "Title is required", http.StatusBadRequest)
		return
	}

	if len(newTask.Title) > 100 {
		sendError(w, "Title must be 100 characters or less", http.StatusBadRequest)
		return
	}

	err = db.QueryRow(
		"INSERT INTO tasks (title, description, completed) VALUES ($1, $2, $3) RETURNING id",
		newTask.Title, newTask.Description, newTask.Completed,
	).Scan(&newTask.ID)

	if err != nil {
		sendError(w, "Failed to create task", http.StatusInternalServerError)
		return
	}
	// id, _ := result.LastInsertId()
	// newTask.ID = int(id)
	writeJSON(w, http.StatusCreated, newTask)
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		sendError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var updatedTask Task
	err = json.NewDecoder(r.Body).Decode(&updatedTask)

	if err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validation
	if updatedTask.Title == "" {
		sendError(w, "Title is required", http.StatusBadRequest)
		return
	}
	if len(updatedTask.Title) > 100 {
		sendError(w, "Title must be 100 characters or less", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("UPDATE tasks SET title = $1, description = $2, completed = $3 WHERE id = $4",
		updatedTask.Title, updatedTask.Description, updatedTask.Completed, id)

	if err != nil {
		sendError(w, "Failed to update task", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		sendError(w, "Task not found", http.StatusNotFound)
		return
	}

	updatedTask.ID = id
	writeJSON(w, http.StatusOK, updatedTask)
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		sendError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("DELETE FROM tasks WHERE id = $1", id)

	if err != nil {
		sendError(w, "Failed to delete task", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		sendError(w, "Task not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}

func toggleTaskCompletion(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		sendError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var t Task

	err = db.QueryRow("SELECT id, title, description, completed FROM tasks WHERE id = $1", id).
		Scan(&t.ID, &t.Title, &t.Description, &t.Completed)

	if err == sql.ErrNoRows {
		sendError(w, "Task not found", http.StatusNotFound)
		return
	} else if err != nil {
		sendError(w, "Failed to fetch task", http.StatusInternalServerError)
		return
	}

	t.Completed = !t.Completed

	_, err = db.Exec("UPDATE tasks SET completed = $1 WHERE id = $2", t.Completed, id)
	if err != nil {
		sendError(w, "Failed to update task", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, t)
}
