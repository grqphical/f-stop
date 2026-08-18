package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) GenerateRouter() http.Handler {
	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	return router
}
