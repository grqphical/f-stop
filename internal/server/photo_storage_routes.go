package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/grqphical/f-stop/internal/models"
	"github.com/jackc/pgx/v5"
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

// API call that returns the metadata of a photo based on it's ID
func (s *Server) GetPhotoHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	metadata, err := s.db.GetPhotoMetadataFromID(id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpError(c, http.StatusNotFound, "PhotoNotFound", "A photo with that ID could not be found")
			return
		}
		httpError(c, http.StatusInternalServerError, "InternalServerError", "An internal server error occured")
		log.Printf("error: %v\n", err)
		return

	}

	c.JSON(http.StatusOK, metadata)
}

// Handler in charge of serving the actual photo files
func (s *Server) StaticPhotoHandler(c *gin.Context) {
	filename := c.Param("filename")
	id := strings.TrimSuffix(filename, filepath.Ext(filename))

	photo, err := s.db.GetPhotoMetadataFromID(id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		fmt.Printf("error: %v\n", err)
		return
	}

	// make sure photo actually belongs to the user
	userVal, exists := c.Get("user")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	user := userVal.(models.User)

	if user.ID != photo.OwnerID {
		c.AbortWithStatus(http.StatusUnauthorized)
		return

	}

	photoFile, err := os.Open(photo.Filepath)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		fmt.Printf("error: %v\n", err)
		return
	}
	defer photoFile.Close()

	c.Header("Content-Type", photo.MimeType)
	io.Copy(c.Writer, photoFile)
}

func (s *Server) DeletePhotoHandler(c *gin.Context) {
	id := c.Param("id")

	photo, err := s.db.GetPhotoMetadataFromID(id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		fmt.Printf("error: %v\n", err)
		return
	}

	// make sure photo actually belongs to the user
	userVal, exists := c.Get("user")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	user := userVal.(models.User)

	if user.ID != photo.OwnerID {
		c.AbortWithStatus(http.StatusUnauthorized)
		return

	}

	err = s.db.DeletePhoto(id)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		fmt.Printf("error: %v\n", err)
		return
	}

	err = os.Remove(photo.Filepath)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		fmt.Printf("error: %v\n", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
