package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/grqphical/f-stop/internal/server"
)

func main() {
	if err := vips.Startup(nil); err != nil {
		log.Fatalf("failed to start libvips: %v\n", err)
	}
	defer vips.Shutdown()

	s, err := server.New()
	if err != nil {
		log.Fatalf("server startup failed: %v\n", err)
	}

	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		log.Fatalf("graceful shutdown failed: %v\n", err)
	}

	log.Println("server stopped")
}
