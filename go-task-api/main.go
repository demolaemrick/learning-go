package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3" // Blank import for SQLite driver
)

var db *sql.DB // Global DB handle so it is accessible to all handlers

func main() {
	PORT := ":9000"
	var err error

	// Initialize SQLite database
	db, err = sql.Open("sqlite3", "./tasks.db")

	if err != nil {
		log.Fatal("Failed to open database connection:", err)
	}
	defer db.Close() // close it when main exit

	// Create tasks table if it doesn’t exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			completed BOOLEAN DEFAULT FALSE
		)
	`)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}

	router := mux.NewRouter()

	//Define routes
	router.HandleFunc("/tasks", getTasks).Methods("GET")
	router.HandleFunc("/tasks/{id}", getTask).Methods("GET")
	router.HandleFunc("/tasks", createTask).Methods("POST")
	router.HandleFunc("/tasks/{id}", updateTask).Methods("PUT")
	router.HandleFunc("/tasks/{id}", deleteTask).Methods("DELETE")
	router.HandleFunc("/tasks/{id}/toggle", toggleTaskCompletion).Methods("PATCH")

	log.Printf("Starting server on port %v...", PORT)

	err = http.ListenAndServe(PORT, router)

	if err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
