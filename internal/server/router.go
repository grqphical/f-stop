package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/grqphical/f-stop/internal/database"
)

type HTTPError struct {
	ErrorType string `json:"errorType"`
	Message   string `json:"message"`
}

func httpError(c *gin.Context, status int, errorType string, message string) {
	c.JSON(status, HTTPError{
		errorType,
		message,
	})
}

func (s *Server) GenerateRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/storage/:filename", s.Authorization(), s.StaticPhotoHandler)
	router.GET("/storage/thumbnails/:filename", s.Authorization(), s.StaticThumbnailHandler)

	router.GET("/health", s.HealthHandler)

	router.POST("/create-account", s.CreateAccountHandler)
	router.POST("/login", s.LoginHandler)
	router.GET("/user", s.Authorization(), s.GetUserHandler)

	router.GET("/tags/names/:name", s.Authorization(), s.GetTagByNameHandler)
	router.POST("/tags", s.Authorization(), s.CreateTagHandler)
	router.GET("/tags", s.Authorization(), s.GetUserTagsHandler)
	router.GET("/tags/:id", s.Authorization(), s.GetTagByIDHandler)
	router.DELETE("/tags/:id", s.Authorization(), s.DeleteTagHandler)
	router.PUT("/tags/:id", s.Authorization(), s.RenameTagHandler)
	router.GET("/tags/:id/photos", s.Authorization(), s.GetTagPhotosHandler)

	router.PUT("/photo", s.Authorization(), s.UploadPhotoHandler)
	router.GET("/photo/:id", s.Authorization(), s.GetPhotoHandler)
	router.DELETE("/photo/:id", s.Authorization(), s.DeletePhotoHandler)
	router.GET("/photo/all", s.Authorization(), s.GetUserPhotosHandler)
	router.PATCH("/photo/:id/tags", s.Authorization(), s.AssignTagHandler)
	router.GET("/photo/:id/tags", s.Authorization(), s.GetPhotoTagsHandler)
	router.DELETE("/photo/:id/tags", s.Authorization(), s.RemovePhotoTagsHandler)

	router.GET("/jobs/:id", s.Authorization(), s.GetJobHandler)

	return router
}

func (s *Server) HealthHandler(c *gin.Context) {
	err := s.db.Health()
	if errors.Is(err, database.ErrDBDown) {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Database down",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "Up",
	})
}
