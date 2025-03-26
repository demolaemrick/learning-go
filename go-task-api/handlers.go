package main

import (
	"encoding/json"
	"net/http"
	"slices"
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

// In-memory task store
var tasks []Task
var nextID = 1

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
	writeJSON(w, http.StatusOK, tasks)
}

func getTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		sendError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	for _, task := range tasks {
		if task.ID == id {
			writeJSON(w, http.StatusOK, task)
			return
		}
	}
	sendError(w, "Task not found", http.StatusNotFound)
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

	newTask.ID = nextID
	nextID++

	tasks = append(tasks, newTask)

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

	for i, task := range tasks {
		if task.ID == id {
			updatedTask.ID = id
			tasks[i] = updatedTask
			writeJSON(w, http.StatusOK, updatedTask)
			return
		}
	}
	sendError(w, "Task not found", http.StatusNotFound)
}
func deleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		sendError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	for i, task := range tasks {
		if task.ID == id {
			// tasks = append(tasks[:i], tasks[i+1:]...)
			tasks = slices.Delete(tasks, i, i+1)
			writeJSON(w, http.StatusNoContent, nil)
			return
		}
	}
	sendError(w, "Task not found", http.StatusNotFound)
}

func toggleTaskCompletion(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		sendError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	for i, task := range tasks {
		if task.ID == id {
			tasks[i].Completed = !tasks[i].Completed
			writeJSON(w, http.StatusOK, tasks[i])
			return
		}
	}
	sendError(w, "Task not found", http.StatusNotFound)
}
