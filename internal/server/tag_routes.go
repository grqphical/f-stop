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
	if tagName == "" {
		httpError(c, http.StatusBadRequest, "MissingName", "A tag name must be provided")
		return
	}

	tagId, err := s.db.CreateTag(tagName, user.ID)
	if err != nil {
		if errors.Is(err, database.ErrUniqueConstraint) {
			httpError(c, http.StatusConflict, "TagAlreadyExists", "You already have a tag with that name")
			return
		}
		httpError(c, http.StatusInternalServerError, "InternalServerError", "An internal server error occured")
		log.Printf("error: %v\n", err)
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
		if errors.Is(err, database.ErrNotFound) {
			httpError(c, http.StatusNotFound, "TagNotFound", "Tag with given ID could not be found")
			return
		}

		httpError(c, http.StatusInternalServerError, "InternalServerError", "An internal server error occured")
		log.Printf("error: %v\n", err)
		return
	}

	if tag.OwnerID != user.ID {
		httpError(c, http.StatusUnauthorized, "Unauthorized", "You are not authorized to do this")
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
		if errors.Is(err, database.ErrNotFound) {
			httpError(c, http.StatusNotFound, "TagNotFound", "Tag with given ID could not be found")
			return
		}
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
	if newName == "" {
		httpError(c, http.StatusBadRequest, "MissingName", "A tag name must be provided")
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

	if tag.OwnerID != user.ID {
		httpError(c, http.StatusUnauthorized, "Unauthorized", "You are not authorized to do this")
		return
	}

	err = s.db.RenameTag(tagID, newName)
	if err != nil {
		if errors.Is(err, database.ErrUniqueConstraint) {
			httpError(c, http.StatusConflict, "TagAlreadyExists", "You already have a tag with that name")
			return
		}
		httpError(c, http.StatusInternalServerError, "InternalServerError", "An internal server error occured")
		log.Printf("error: %v\n", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func (s *Server) AssignTagHandler(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	user := userVal.(models.User)

	c.Header("Accept", "application/json")

	photoID := c.Param("id")

	var payload models.TagAssigmentPayload
	err := json.NewDecoder(c.Request.Body).Decode(&payload)
	if err != nil {
		httpError(c, http.StatusBadRequest, "InvalidRequest", "Your request had an invalid structure")
		log.Printf("error: %v\n", err)
		return
	}

	photo, err := s.db.GetPhotoMetadataFromID(photoID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			httpError(c, http.StatusNotFound, "PhotoNotFound", "A photo with that ID could not be found")
			return
		}
		httpError(c, http.StatusInternalServerError, "InternalServerError", "An internal server error occured")
		log.Printf("error: %v\n", err)
		return
	}

	if photo.OwnerID != user.ID {
		httpError(c, http.StatusUnauthorized, "Unauthorized", "You are not authorized to do this")
		return
	}

	for _, tagID := range payload.Tags {
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
		if tag.OwnerID != user.ID {
			httpError(c, http.StatusUnauthorized, "Unauthorized", "You are not authorized to do this")
			return
		}
	}

	err = s.db.AssignPhotoTags(payload.Tags, photoID)
	if err != nil {
		httpError(c, http.StatusInternalServerError, "InternalServerError", "An internal server error occured")
		log.Printf("error: %v\n", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func (s *Server) GetPhotoTagsHandler(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	user := userVal.(models.User)

	c.Header("Content-Type", "application/json")

	photoID := c.Param("id")

	photo, err := s.db.GetPhotoMetadataFromID(photoID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			httpError(c, http.StatusNotFound, "PhotoNotFound", "A photo with that ID could not be found")
			return
		}
		httpError(c, http.StatusInternalServerError, "InternalServerError", "An internal server error occured")
		log.Printf("error: %v\n", err)
		return
	}

	if photo.OwnerID != user.ID {
		httpError(c, http.StatusUnauthorized, "Unauthorized", "You are not authorized to do this")
		return
	}

	tags, err := s.db.GetPhotoTags(photoID)
	if err != nil {
		httpError(c, http.StatusInternalServerError, "InternalServerError", "An internal server error occured")
		log.Printf("error: %v\n", err)
		return
	}

	if tags == nil {
		tags = []models.Tag{}
	}

	c.JSON(http.StatusOK, gin.H{
		"tags": tags,
	})
}

func (s *Server) GetTagPhotosHandler(c *gin.Context) {
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
		if errors.Is(err, database.ErrNotFound) {
			httpError(c, http.StatusNotFound, "TagNotFound", "Tag with given ID could not be found")
			return
		}
		httpError(c, http.StatusInternalServerError, "InternalServerError", "An internal server error occured")
		log.Printf("error: %v\n", err)
		return
	}

	if tag.OwnerID != user.ID {
		httpError(c, http.StatusUnauthorized, "Unauthorized", "You are not authorized to do this")
		return
	}

	photos, err := s.db.GetTagPhotos(tagID)
	if err != nil {
		httpError(c, http.StatusInternalServerError, "InternalServerError", "An internal server error occured")
		log.Printf("error: %v\n", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"photos": photos,
	})
}

func (s *Server) RemovePhotoTagsHandler(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	user := userVal.(models.User)

	photoID := c.Param("id")

	photo, err := s.db.GetPhotoMetadataFromID(photoID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			httpError(c, http.StatusNotFound, "PhotoNotFound", "A photo with that ID could not be found")
			return
		}
		httpError(c, http.StatusInternalServerError, "InternalServerError", "An internal server error occured")
		log.Printf("error: %v\n", err)
		return
	}

	if photo.OwnerID != user.ID {
		httpError(c, http.StatusUnauthorized, "Unauthorized", "You are not authorized to do this")
		return
	}

	var payload models.TagAssigmentPayload
	if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil {
		httpError(c, http.StatusBadRequest, "InvalidRequest", "Your request had an invalid structure")
		log.Printf("error: %v\n", err)
		return
	}

	for _, tagID := range payload.Tags {
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
		if tag.OwnerID != user.ID {
			httpError(c, http.StatusUnauthorized, "Unauthorized", "You are not authorized to do this")
			return
		}
	}

	for _, tagID := range payload.Tags {
		if err := s.db.RemovePhotoTag(tagID, photoID); err != nil {
			httpError(c, http.StatusInternalServerError, "InternalServerError", "An internal server error occured")
			log.Printf("error: %v\n", err)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{})
}
