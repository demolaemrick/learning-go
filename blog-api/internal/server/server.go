package server

import (
	"net/http"

	"github.com/demolaemrick/learning-go/blog-api/internal/handlers"
	"github.com/demolaemrick/learning-go/blog-api/internal/store"
	"github.com/gin-gonic/gin"
)

func Run() error {
	router := gin.Default()
	postStore := store.NewPostStore()
	postsHandler := handlers.NewPostsHandler(postStore)

	v1 := router.Group("/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  "ok",
				"service": "blog-api",
				"version": "v1",
			})
		})
		v1.POST("/posts", postsHandler.CreatePost)
		v1.GET("posts", postsHandler.GetPosts)
	}

	return router.Run(":8080")
}
