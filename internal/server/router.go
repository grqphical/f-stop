package server

import (
	"net/http"

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

func (s *Server) GenerateRouter() http.Handler {
	router := gin.Default()

	api := router.Group("/api")
	v1 := api.Group("/v1")

	v1.POST("/create-account", s.CreateAccountHandler)

	return router
}
