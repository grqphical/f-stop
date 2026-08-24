package server

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/grqphical/f-stop/internal/models"
)

func (s *Server) CreateTagHandler(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	user := userVal.(models.User)

	tagName := c.PostForm("name")

	tagId, err := s.db.CreateTag(tagName, user.ID)
	if err != nil {
		httpError(c, http.StatusInternalServerError, "InternalServerError", "An internal server error occured")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"tagId": tagId,
	})
}

func (s *Server) GetTagByNameHandler(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	user := userVal.(models.User)

	tagName := c.Param("name")

	tag, err := s.db.GetTagByName(tagName, user.ID)

	if err != nil {
		httpError(c, http.StatusInternalServerError, "InternalServerError", "An internal server error occured")
		log.Printf("error: %v\n", err)
		return
	}

	c.JSON(http.StatusOK, tag)
}

func (s *Server) GetTagByIDHandler(c *gin.Context) {
	_, exists := c.Get("user")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	tagID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		httpError(c, http.StatusBadRequest, "InvalidID", "ID must be a positive integer")
		return
	}

	tag, err := s.db.GetTagByID(tagID)

	if err != nil {
		httpError(c, http.StatusInternalServerError, "InternalServerError", "An internal server error occured")
		log.Printf("error: %v\n", err)
		return
	}

	c.JSON(http.StatusOK, tag)
}
