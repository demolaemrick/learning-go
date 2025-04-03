package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type APIResponse struct {
	Status string       `json:"status"`
	Data   any          `json:"data,omitempty"`  // For success
	Error  *ErrorDetail `json:"error,omitempty"` // For errors, nil if success
}

type ErrorDetail struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"` // Optional, e.g., "VALIDATION_ERROR"
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// writeJSON sets the content type and writes JSON data to the response
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if status == http.StatusNoContent {
		return // No body for 204
	}
	resp := APIResponse{
		Status: "success",
		Data:   data,
	}
	json.NewEncoder(w).Encode(resp)
}

func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	resp := APIResponse{
		Status: "error",
		Error: &ErrorDetail{
			Message: message,
		},
	}
	json.NewEncoder(w).Encode(resp)
}

func getTasks(w http.ResponseWriter, r *http.Request) {
	query := "SELECT id, title, description, completed, created_at, updated_at FROM tasks"
	var args []interface{}
	var conditions []string

	completedParam := r.URL.Query().Get("completed")
	if completedParam != "" {
		completed, err := strconv.ParseBool(completedParam)
		if err != nil {
			sendError(w, "Invalid 'completed' parameter - must be true or false", http.StatusBadRequest)
			return
		}
		conditions = append(conditions, "completed = $1")
		args = append(args, completed)
	}

	sortField := strings.ToLower(r.URL.Query().Get("sort"))
	order := strings.ToLower(r.URL.Query().Get("order"))
	
	validSortFields := []string{"title", "created_at", "updated_at"}
	if sortField != "" {
		if !slices.Contains(validSortFields, sortField) {
			sendError(w, "Invalid 'sort' parameter - must be title, created_at, or updated_at", http.StatusBadRequest)
			return
		}

		sortOrder := "ASC"

		if order == "desc" {
			sortOrder = "DESC"
		} else if order != "" && order != "asc" {
			sendError(w, "Invalid 'order' parameter - must be asc or desc", http.StatusBadRequest)
			return
		}
		query += " ORDER BY " + sortField + " " + sortOrder
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		sendError(w, "Failed to fetch tasks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task

	for rows.Next() {
		var t Task
		err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &t.CreatedAt, &t.UpdatedAt)

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

	err = db.QueryRow("SELECT id, title, description, completed, created_at, updated_at FROM tasks WHERE id = $1", id).
		Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &t.CreatedAt, &t.UpdatedAt)

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
		"INSERT INTO tasks (title, description, completed) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at",
		newTask.Title, newTask.Description, newTask.Completed,
	).Scan(&newTask.ID, &newTask.CreatedAt, &newTask.UpdatedAt)

	if err != nil {
		log.Println(err)
		sendError(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

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

	// Fetch updated task to include timestamps
	err = db.QueryRow(
		"SELECT id, title, description, completed, created_at, updated_at FROM tasks WHERE id = $1", id,
	).Scan(&updatedTask.ID, &updatedTask.Title, &updatedTask.Description, &updatedTask.Completed, &updatedTask.CreatedAt, &updatedTask.UpdatedAt)
	if err != nil {
		sendError(w, "Failed to fetch updated task", http.StatusInternalServerError)
		return
	}

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

	// Fetch updated task
	err = db.QueryRow(
		"SELECT id, title, description, completed, created_at, updated_at FROM tasks WHERE id = $1", id,
	).Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		sendError(w, "Failed to fetch updated task", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, t)
}

func deleteAllTasks(w http.ResponseWriter, r *http.Request) {
	_, err := db.Exec("TRUNCATE TABLE tasks RESTART IDENTITY")

	if err != nil {
		log.Printf("Error truncating tasks: %v", err)
		sendError(w, "Failed to delete all tasks", http.StatusInternalServerError)
		return
	}

	log.Printf("All tasks deleted successfully")
	writeJSON(w, http.StatusNoContent, nil)
}
