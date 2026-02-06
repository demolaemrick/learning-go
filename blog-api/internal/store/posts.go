package store

import (
	"sync"
	"time"

	"github.com/demolaemrick/learning-go/blog-api/internal/models"
)

// PostStore holds posts in memory (for development; replace with DB later).
type PostStore struct {
	mu     sync.RWMutex
	posts  []models.Post
	nextID int
}

// NewPostStore returns a new in-memory post store.
func NewPostStore() *PostStore {
	return &PostStore{nextID: 1}
}

// Create adds a new post and returns it with ID and CreatedAt set.
func (s *PostStore) Create(title, body string) models.Post {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	p := models.Post{
		ID:        s.nextID,
		Title:     title,
		Body:      body,
		CreatedAt: now,
	}
	s.nextID++
	s.posts = append(s.posts, p)
	return p
}

func (s *PostStore) List() []models.Post {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]models.Post{}, s.posts...)
}
