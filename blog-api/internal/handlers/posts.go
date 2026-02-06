package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/demolaemrick/learning-go/blog-api/internal/models"
	"github.com/demolaemrick/learning-go/blog-api/internal/response"
	"github.com/demolaemrick/learning-go/blog-api/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// PostsHandler handles post-related HTTP requests.
type PostsHandler struct {
	store *store.PostStore
}

// NewPostsHandler returns a new PostsHandler.
func NewPostsHandler(store *store.PostStore) *PostsHandler {
	return &PostsHandler{store: store}
}

// CreatePost handles POST /v1/posts.
func (h *PostsHandler) CreatePost(c *gin.Context) {
	var req models.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationError(c, err)
		return
	}

	// Extra safety: trim whitespace and ensure not empty after trimming.
	req.Title = strings.TrimSpace(req.Title)
	req.Body = strings.TrimSpace(req.Body)

	fieldErrors := map[string]string{}
	if req.Title == "" {
		fieldErrors["title"] = "title is required"
	}
	if req.Body == "" {
		fieldErrors["body"] = "body is required"
	}
	if len(fieldErrors) > 0 {
		response.JSONError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request payload", fieldErrors)
		return
	}

	post := h.store.Create(req.Title, req.Body)
	c.JSON(http.StatusCreated, post)
}

func (h *PostsHandler) GetPosts(c *gin.Context) {
	posts := h.store.List()
	c.JSON(http.StatusOK, posts)
}

// handleValidationError converts validator errors into a structured error response.
func handleValidationError(c *gin.Context, err error) {
	var verr validator.ValidationErrors
	if errors.As(err, &verr) {
		fields := make(map[string]string)
		for _, fe := range verr {
			field := strings.ToLower(fe.Field())
			switch fe.Tag() {
			case "required":
				fields[field] = "is required"
			case "min":
				fields[field] = "is too short"
			case "max":
				fields[field] = "is too long"
			default:
				fields[field] = fe.Error()
			}
		}
		response.JSONError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request payload", fields)
		return
	}

	response.JSONError(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON body", nil)
}

