package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/grqphical/f-stop/internal/auth"
	"github.com/grqphical/f-stop/internal/database"
	"github.com/grqphical/f-stop/internal/models"
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
			httpError(c, http.StatusInternalServerError, "InternalServerError", "an internal server error occured")
			return
		}
	}

	c.JSON(http.StatusCreated, user)
}

func (s *Server) LoginHandler(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	user, err := s.db.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			httpError(c, http.StatusNotFound, "UserNotFound", "user with given email could not be found")
			return
		}
		httpError(c, http.StatusInternalServerError, "InternalServerError", "an internal server error occured")
		fmt.Printf("error: %v\n", err)
		return
	}

	match, err := auth.ComparePasswordAndHash(password, user.PasswordHash)
	if err != nil {
		httpError(c, http.StatusInternalServerError, "InternalServerError", "an internal server error occured")
		fmt.Printf("error: %v\n", err)
		return
	}

	if !match {
		httpError(c, http.StatusUnauthorized, "InvalidCredentials", "invalid credentials")
		return
	}

	jwtToken, err := auth.GenerateJWT(user.ID)
	if err != nil {
		httpError(c, http.StatusInternalServerError, "InternalServerError", "an internal server error occured")
		fmt.Printf("error: %v\n", err)
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", jwtToken, int(auth.JWTExpiryDuration.Seconds()), "", "", false, true)

	c.JSON(http.StatusOK, gin.H{})

}

func (s *Server) GetUserHandler(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
	user := userVal.(models.User)
	c.JSON(http.StatusOK, user)
}
