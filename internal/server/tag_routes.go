package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/grqphical/f-stop/internal/database"
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
		if errors.Is(err, database.ErrNotFound) {
			httpError(c, http.StatusNotFound, "TagNotFound", "Tag with given name could not be found")
			return
		}
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
		if errors.Is(err, database.ErrNotFound) {
			httpError(c, http.StatusNotFound, "TagNotFound", "Tag with given ID could not be found")
			return
		}

		httpError(c, http.StatusInternalServerError, "InternalServerError", "An internal server error occured")
		log.Printf("error: %v\n", err)
		return
	}

	c.JSON(http.StatusOK, tag)
}

func (s *Server) GetUserTagsHandler(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	user := userVal.(models.User)

	tags, err := s.db.GetUserTags(user.ID)
	if err != nil {
		httpError(c, http.StatusInternalServerError, "InternalServerError", "An internal server error occured")
		log.Printf("error: %v\n", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tags": tags,
	})
}

func (s *Server) DeleteTagHandler(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	user := userVal.(models.User)

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

	if tag.OwnerID != user.ID {
		httpError(c, http.StatusUnauthorized, "Unauthorized", "You are not authorized to do this")
		return
	}

	err = s.db.DeleteTag(tagID)

	if err != nil {
		httpError(c, http.StatusInternalServerError, "InternalServerError", "An internal server error occured")
		log.Printf("error: %v\n", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func (s *Server) RenameTagHandler(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	user := userVal.(models.User)

	tagID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		httpError(c, http.StatusBadRequest, "InvalidID", "ID must be a positive integer")
		return
	}

	newName := c.PostForm("name")

	tag, err := s.db.GetTagByID(tagID)
	if err != nil {
		httpError(c, http.StatusInternalServerError, "InternalServerError", "An internal server error occured")
		log.Printf("error: %v\n", err)
		return
	}

	if tag.OwnerID != user.ID {
		httpError(c, http.StatusUnauthorized, "Unauthorized", "You are not authorized to do this")
		return
	}

	err = s.db.RenameTag(tagID, newName)
	if err != nil {
		httpError(c, http.StatusInternalServerError, "InternalServerError", "An internal server error occured")
		log.Printf("error: %v\n", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func (s *Server) AssignTagHandler(c *gin.Context) {
	_, exists := c.Get("user")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Header("Accept", "application/json")

	photoID := c.Param("id")

	var payload models.TagAssigmentPayload
	err := json.NewDecoder(c.Request.Body).Decode(&payload)
	if err != nil {
		httpError(c, http.StatusBadRequest, "InvalidRequest", "Your request had an invalid structure")
		log.Printf("error: %v\n", err)
		return
	}

	err = s.db.AssignPhotoTags(payload.Tags, photoID)
	if err != nil {
		httpError(c, http.StatusInternalServerError, "InternalServerError", "An internal server error occured")
		log.Printf("error: %v\n", err)
		return
	}
}

func (s *Server) GetPhotoTagsHandler(c *gin.Context) {
	_, exists := c.Get("user")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Header("Content-Type", "application/json")

	photoID := c.Param("id")

	tags, err := s.db.GetPhotoTags(photoID)
	if err != nil {
		httpError(c, http.StatusInternalServerError, "InternalServerError", "An internal server error occured")
		log.Printf("error: %v\n", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tags": tags,
	})
}
