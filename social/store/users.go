package store

import (
	"context"
	"database/sql"
)

type User struct {
	ID        int64 `json:"id"`
	Username  int64 `json:"username"`
	Email     int64 `json:"email"`
	Password  int64 `json:"-"`
	CreatedAt int64 `json:"created_at"`
}
type UsersStore struct {
	db *sql.DB
}

func (s *UsersStore) Create(ctx context.Context, user *User) error {
	query := `
			INSERT INTO users (username, email, password, created_at)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`
	err := s.db.QueryRowContext(ctx, query, user.Username, user.Email, user.Password, user.CreatedAt).Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}
