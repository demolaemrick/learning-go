package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if input.Username == "" || input.Password == "" {
		h.SendError(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Hashing error: %v", err)
		h.SendError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	var userID int
	err = h.DB.QueryRow(
		"INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id",
		input.Username, string(hashedPassword),
	).Scan(&userID)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			h.SendError(w, "Username already taken", http.StatusConflict)
			return
		}
		log.Printf("Insert user error: %v", err)
		h.SendError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	user := User{ID: userID, Username: input.Username}
	h.WriteJSON(w, http.StatusCreated, user)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.Username == "" || input.Password == "" {
		h.SendError(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	var user User
	var hashedPassword string

	err := h.DB.QueryRow(
		"SELECT id, username, password FROM users WHERE username = $1",
		input.Username,
	).Scan(&user.ID, &user.Username, &hashedPassword)

	if err == sql.ErrNoRows {
		h.SendError(w, "Invalid username or password", http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Printf("Query user error: %v", err)
		h.SendError(w, "Failed to login", http.StatusInternalServerError)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(input.Password)); err != nil {
		h.SendError(w, "Invalid username or password", http.StatusUnauthorized)
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
		h.SendError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}
	h.WriteJSON(w, http.StatusOK, map[string]string{"token": tokenString})
}

func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query("SELECT id, username FROM users")
	if err != nil {
		log.Printf("Query error: %v", err)
		h.SendError(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username); err != nil {
			log.Printf("Scan error: %v", err)
			h.SendError(w, "Failed to scan users", http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		h.SendError(w, "Error processing users", http.StatusInternalServerError)
		return
	}

	h.WriteJSON(w, http.StatusOK, users)
}
