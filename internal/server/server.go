package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port int
}

func New() *http.Server {
	s := &Server{}

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatalf("failed to parse port: %v\n", err)
	}

	s.port = port

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: s.GenerateRouter(),
	}
}
