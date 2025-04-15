package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}
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

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

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

// Validate env vars at startup
func init() {
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}
}

func signup(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if input.Username == "" || input.Password == "" {
		sendError(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Hashing error: %v", err)
		sendError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	var userID int
	err = db.QueryRow(
		"INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id",
		input.Username, string(hashedPassword),
	).Scan(&userID)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			sendError(w, "Username already taken", http.StatusConflict)
			return
		}
		log.Printf("Insert user error: %v", err)
		sendError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	user := User{ID: userID, Username: input.Username}
	writeJSON(w, http.StatusCreated, user)
}

func login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.Username == "" || input.Password == "" {
		sendError(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	var user User
	var hashedPassword string

	err := db.QueryRow(
		"SELECT id, username, password FROM users WHERE username = $1",
		input.Username,
	).Scan(&user.ID, &user.Username, &hashedPassword)

	if err == sql.ErrNoRows {
		sendError(w, "Invalid username or password", http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Printf("Query user error: %v", err)
		sendError(w, "Failed to login", http.StatusInternalServerError)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(input.Password)); err != nil {
		sendError(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // 1 day expiry
	})

	tokenString, err := token.SignedString(jwtSecret)

	if err != nil {
		log.Printf("JWT signing error: %v", err)
		sendError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": tokenString})
}

func getTasks(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(int)
	log.Printf("User ID from context: %v", ok)
	if !ok {
		sendError(w, "Invalid user ID", http.StatusInternalServerError)
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
			sendError(w, "Invalid 'completed' parameter - must be true or false", http.StatusBadRequest)
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

	userID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		sendError(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	var t Task
	err = db.QueryRow("SELECT id, title, description, completed, created_at, updated_at FROM tasks WHERE id = $1 AND user_id = $2", id, userID).
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

	userID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		sendError(w, "Invalid user ID", http.StatusInternalServerError)
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
		"INSERT INTO tasks (title, description, completed, user_id) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at",
		newTask.Title, newTask.Description, newTask.Completed, userID,
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

	userID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		sendError(w, "Invalid user ID", http.StatusInternalServerError)
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

	result, err := db.Exec("UPDATE tasks SET title = $1, description = $2, completed = $3 WHERE id = $4 AND user_id = $5",
		updatedTask.Title, updatedTask.Description, updatedTask.Completed, id, userID)

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

	userID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		sendError(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}
	result, err := db.Exec("DELETE FROM tasks WHERE id = $1 AND user_id = $2", id, userID)

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

	userID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		sendError(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	var t Task

	err = db.QueryRow("SELECT id, title, description, completed FROM tasks WHERE id = $1 AND user_id = $2", id, userID).
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
	userID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		sendError(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}
	_, err := db.Exec("DELETE FROM tasks WHERE user_id = $1", userID)

	if err != nil {
		log.Printf("Delete error: %v", err)
		sendError(w, "Failed to delete all tasks", http.StatusInternalServerError)
		return
	}

	log.Printf("All tasks deleted successfully")
	writeJSON(w, http.StatusNoContent, nil)
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, username FROM users")
	if err != nil {
		log.Printf("Query error: %v", err)
		sendError(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username); err != nil {
			log.Printf("Scan error: %v", err)
			sendError(w, "Failed to scan users", http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		sendError(w, "Error processing users", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, users)
}
