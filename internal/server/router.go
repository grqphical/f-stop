package server

import (
	"github.com/gin-gonic/gin"
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

	api := router.Group("/api")
	v1 := api.Group("/v1")

	v1.POST("/create-account", s.CreateAccountHandler)
	v1.POST("/login", s.LoginHandler)
	v1.GET("/user", s.Authorization(), s.GetUserHandler)

	v1.PUT("/photo", s.Authorization(), s.UploadPhotoHandler)
	v1.GET("/photo/:id", s.Authorization(), s.GetPhotoHandler)
	v1.DELETE("/photo/:id", s.Authorization(), s.DeletePhotoHandler)
	v1.GET("/photo/all", s.Authorization(), s.GetUserPhotosHandler)

	v1.GET("/jobs/:id", s.Authorization(), s.GetJobHandler)

	return router
}
