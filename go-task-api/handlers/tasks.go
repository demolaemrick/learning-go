package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// Validate env vars at startup
func init() {
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}
}

func GetJWTSecret() []byte {
	return jwtSecret
}

func (h *Handler) GetTasks(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int)
	log.Printf("User ID from context: %v", ok)
	if !ok {
		h.SendError(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}
	query := "SELECT id, title, description, completed, created_at, updated_at FROM tasks WHERE user_id = $1"
	var args []interface{}
	args = append(args, userID)
	var conditions []string
	paramIndex := 2 // Track $n placeholders

	completedParam := r.URL.Query().Get("completed")
	if completedParam != "" {
		completed, err := strconv.ParseBool(completedParam)
		if err != nil {
			h.SendError(w, "Invalid 'completed' parameter - must be true or false", http.StatusBadRequest)
			return
		}
		conditions = append(conditions, "completed = $1")
		args = append(args, completed)
	}

	searchParam := r.URL.Query().Get("search")
	if searchParam != "" && strings.TrimSpace(searchParam) != "" {
		searchPattern := "%" + searchParam + "%"
		conditions = append(conditions, "(title ILIKE $"+strconv.Itoa(paramIndex)+" OR description ILIKE $"+strconv.Itoa(paramIndex)+")")
		args = append(args, searchPattern)
		paramIndex++
	}

	sortField := strings.ToLower(r.URL.Query().Get("sort"))
	order := strings.ToLower(r.URL.Query().Get("order"))
	validSortFields := []string{"title", "created_at", "updated_at"}
	if sortField != "" {
		if !slices.Contains(validSortFields, sortField) {
			h.SendError(w, "Invalid 'sort' parameter - must be title, created_at, or updated_at", http.StatusBadRequest)
			return
		}

		sortOrder := "ASC"

		if order == "desc" {
			sortOrder = "DESC"
		} else if order != "" && order != "asc" {
			h.SendError(w, "Invalid 'order' parameter - must be asc or desc", http.StatusBadRequest)
			return
		}
		query += " ORDER BY " + sortField + " " + sortOrder
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		h.SendError(w, "Failed to fetch tasks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task

	for rows.Next() {
		var t Task
		err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &t.CreatedAt, &t.UpdatedAt)

		if err != nil {
			h.SendError(w, "Failed to scan tasks", http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, t)
	}
	h.WriteJSON(w, http.StatusOK, tasks)
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		h.SendError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		h.SendError(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	var t Task
	err = h.DB.QueryRow("SELECT id, title, description, completed, created_at, updated_at FROM tasks WHERE id = $1 AND user_id = $2", id, userID).
		Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &t.CreatedAt, &t.UpdatedAt)

	if err == sql.ErrNoRows {
		h.SendError(w, "Task not found", http.StatusNotFound)
		return
	} else if err != nil {
		h.SendError(w, "Failed to fetch task", http.StatusInternalServerError)
		return
	}
	h.WriteJSON(w, http.StatusOK, t)
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var newTask Task
	err := json.NewDecoder(r.Body).Decode(&newTask)

	if err != nil {
		h.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		h.SendError(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	// Validation
	if newTask.Title == "" {
		h.SendError(w, "Title is required", http.StatusBadRequest)
		return
	}

	if len(newTask.Title) > 100 {
		h.SendError(w, "Title must be 100 characters or less", http.StatusBadRequest)
		return
	}

	err = h.DB.QueryRow(
		"INSERT INTO tasks (title, description, completed, user_id) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at",
		newTask.Title, newTask.Description, newTask.Completed, userID,
	).Scan(&newTask.ID, &newTask.CreatedAt, &newTask.UpdatedAt)

	if err != nil {
		log.Println(err)
		h.SendError(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	h.WriteJSON(w, http.StatusCreated, newTask)
}

func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		h.SendError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var updatedTask Task
	err = json.NewDecoder(r.Body).Decode(&updatedTask)

	if err != nil {
		h.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		h.SendError(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	// Validation
	if updatedTask.Title == "" {
		h.SendError(w, "Title is required", http.StatusBadRequest)
		return
	}
	if len(updatedTask.Title) > 100 {
		h.SendError(w, "Title must be 100 characters or less", http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec("UPDATE tasks SET title = $1, description = $2, completed = $3 WHERE id = $4 AND user_id = $5",
		updatedTask.Title, updatedTask.Description, updatedTask.Completed, id, userID)

	if err != nil {
		h.SendError(w, "Failed to update task", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		h.SendError(w, "Task not found", http.StatusNotFound)
		return
	}

	// Fetch updated task to include timestamps
	err = h.DB.QueryRow(
		"SELECT id, title, description, completed, created_at, updated_at FROM tasks WHERE id = $1", id,
	).Scan(&updatedTask.ID, &updatedTask.Title, &updatedTask.Description, &updatedTask.Completed, &updatedTask.CreatedAt, &updatedTask.UpdatedAt)
	if err != nil {
		h.SendError(w, "Failed to fetch updated task", http.StatusInternalServerError)
		return
	}

	h.WriteJSON(w, http.StatusOK, updatedTask)
}

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		h.SendError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		h.SendError(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}
	result, err := h.DB.Exec("DELETE FROM tasks WHERE id = $1 AND user_id = $2", id, userID)

	if err != nil {
		h.SendError(w, "Failed to delete task", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		h.SendError(w, "Task not found", http.StatusNotFound)
		return
	}

	h.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *Handler) ToggleTaskCompletion(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		h.SendError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		h.SendError(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	var t Task

	err = h.DB.QueryRow("SELECT id, title, description, completed FROM tasks WHERE id = $1 AND user_id = $2", id, userID).
		Scan(&t.ID, &t.Title, &t.Description, &t.Completed)

	if err == sql.ErrNoRows {
		h.SendError(w, "Task not found", http.StatusNotFound)
		return
	} else if err != nil {
		h.SendError(w, "Failed to fetch task", http.StatusInternalServerError)
		return
	}

	t.Completed = !t.Completed

	_, err = h.DB.Exec("UPDATE tasks SET completed = $1 WHERE id = $2", t.Completed, id)
	if err != nil {
		h.SendError(w, "Failed to update task", http.StatusInternalServerError)
		return
	}

	// Fetch updated task
	err = h.DB.QueryRow(
		"SELECT id, title, description, completed, created_at, updated_at FROM tasks WHERE id = $1", id,
	).Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		h.SendError(w, "Failed to fetch updated task", http.StatusInternalServerError)
		return
	}

	h.WriteJSON(w, http.StatusOK, t)
}

func (h *Handler) DeleteAllTasks(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		h.SendError(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}
	_, err := h.DB.Exec("DELETE FROM tasks WHERE user_id = $1", userID)

	if err != nil {
		log.Printf("Delete error: %v", err)
		h.SendError(w, "Failed to delete all tasks", http.StatusInternalServerError)
		return
	}

	log.Printf("All tasks deleted successfully")
	h.WriteJSON(w, http.StatusNoContent, nil)
}
