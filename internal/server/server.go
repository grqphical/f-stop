package server

import (
	"fmt"
	"log"
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

func New() *http.Server {
	s := &Server{}

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatalf("failed to parse port: %v\n", err)
	}

	s.port = port
	s.db = database.New(workerCount)
	s.si = storage.New()
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
	return httpServer
}

func NewMockServer() *gin.Engine {
	s := &Server{}
	s.db = database.NewMockDatabase()
	s.si = storage.NewMockInterface()
	s.wm = workers.NewWorkerManager(workerCount, workers.ImageProcessingWorker, s.db, s.si)

	return s.GenerateRouter()

}
