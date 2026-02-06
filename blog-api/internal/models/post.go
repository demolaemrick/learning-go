package models

import "time"

type Post struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type CreatePostRequest struct {
	Title string `json:"title" binding:"required,min=1,max=200"`
	Body string `json:"body" binding:"required,min=1,max=5000"`
}
