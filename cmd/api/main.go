package main

import (
	"log"

	"github.com/grqphical/f-stop/internal/server"
)

func main() {
	s := server.New()

	log.Printf("starting server on %s\n", s.Addr)
	s.ListenAndServe()
}
