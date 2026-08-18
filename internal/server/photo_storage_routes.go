package server

import (
	"log"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/grqphical/f-stop/internal/models"
)

func (s *Server) UploadPhotoHandler(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	user := userVal.(models.User)

	c.Header("Accept", "multipart/form-data")
	file, err := c.FormFile("file")
	if err != nil {
		httpError(c, http.StatusBadRequest, "InvalidFormData", "Server failed to parse form data. Data must be of type multipart/form-data")
		return
	}
	filePath := filepath.Base(file.Filename)
	extension := filepath.Ext(filePath)
	mimeType := mime.TypeByExtension(extension)
	size := file.Size

	uuid, err := s.db.CreatePhotoMetadata(size, mimeType, user.ID)
	if err != nil {
		httpError(c, http.StatusInternalServerError, "InternalServerError", "An internal server error occured")
		log.Printf("error: %v\n", err)
		return
	}

	outputPath := s.si.StoreItem(uuid + extension)
	c.SaveUploadedFile(file, outputPath)

	err = s.db.UpdatePhotoMetadataFilePath(uuid, outputPath)
	if err != nil {
		httpError(c, http.StatusInternalServerError, "InternalServerError", "An internal server error occured")
		log.Printf("error: %v\n", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"photoId": uuid,
	})

}

// API call that returns a permalink to a photo based on it's ID
func (s *Server) GetPhotoHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	metadata, err := s.db.GetPhotoMetadataFromID(id)
	if err != nil {
		httpError(c, http.StatusInternalServerError, "InternalServerError", "An internal server error occured")
		log.Printf("error: %v\n", err)
		return

	}

	c.JSON(http.StatusOK, metadata)
}
