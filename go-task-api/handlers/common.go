package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type Handler struct {
	DB *sql.DB
}

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

// WriteJSON sets the content type and writes JSON data to the response
func (h *Handler) WriteJSON(w http.ResponseWriter, status int, data any) {
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

func (h *Handler) SendError(w http.ResponseWriter, message string, statusCode int) {
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
