package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) CreateAccountHandler(c *gin.Context) {
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")

	user, err := s.db.CreateUser(username, email, password)
	if err != nil {
		httpError(c, http.StatusBadRequest, "InvalidUser", err.Error())
		return
	}

	c.JSON(http.StatusCreated, user)
}
