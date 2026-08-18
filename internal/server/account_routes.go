package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/grqphical/f-stop/internal/database"
)

func (s *Server) CreateAccountHandler(c *gin.Context) {
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")

	user, err := s.db.CreateUser(username, email, password)
	if err != nil {
		if errors.Is(err, database.ErrUniqueConstraint) {
			httpError(c, http.StatusBadRequest, "UserAlreadyExists", "username and/or email already exists")
			return
		} else {
			httpError(c, http.StatusInternalServerError, "InternalServerError", "internal server error")
		}
	}

	c.JSON(http.StatusCreated, user)
}
