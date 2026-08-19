package server

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/grqphical/f-stop/internal/database"
)

func (s *Server) GetJobHandler(c *gin.Context) {
	idString := c.Param("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		httpError(c, http.StatusBadRequest, "InvalidID", "ID must be an integer")
		return
	}

	job, err := s.db.GetJob(id)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			httpError(c, http.StatusNotFound, "NotFound", "Job with given ID not found")
			return
		}
		httpError(c, http.StatusInternalServerError, "InternalServerError", "An internal server error occured")
		log.Printf("error: %v\n", err)
		return
	}

	c.JSON(http.StatusOK, job)
}
