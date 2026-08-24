package server

import (
	"log"
	"net/http"

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
