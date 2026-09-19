package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/grqphical/f-stop/internal/database"
	"github.com/grqphical/f-stop/internal/storage"
	"github.com/grqphical/f-stop/internal/workers"
	_ "github.com/joho/godotenv/autoload"
)

var workerCount int = 8

type Server struct {
	port int
	db   database.DBInterface
	si   storage.StorageInterface
	wm   *workers.WorkerManager
}

func New() (*http.Server, error) {
	s := &Server{}

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse port: %w", err)
	}

	s.port = port
	s.db, err = database.New(workerCount)
	if err != nil {
		return nil, err
	}
	s.si, err = storage.New()
	if err != nil {
		s.db.Close()
		return nil, err
	}
	s.wm = workers.NewWorkerManager(
		workerCount,
		workers.ImageProcessingWorker,
		s.db,
		s.si,
	)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: s.GenerateRouter(),
	}

	httpServer.RegisterOnShutdown(s.db.Close)
	httpServer.RegisterOnShutdown(s.wm.Close)
	return httpServer, nil
}

func NewMockServer() *gin.Engine {
	s := &Server{}
	s.db = database.NewMockDatabase()
	s.si = storage.NewMockInterface()
	s.wm = workers.NewWorkerManager(workerCount, workers.ImageProcessingWorker, s.db, s.si)

	return s.GenerateRouter()

}
