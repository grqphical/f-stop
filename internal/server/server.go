package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/grqphical/f-stop/internal/database"
	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port int
	db   *database.Database
}

func New() *http.Server {
	s := &Server{}

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatalf("failed to parse port: %v\n", err)
	}

	s.port = port
	s.db = database.New()

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: s.GenerateRouter(),
	}

	httpServer.RegisterOnShutdown(s.db.Close)
	return httpServer
}
