package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github/tahcohcat/same-same/internal/server"
)

func main() {
	var addr = flag.String("addr", ":8080", "HTTP service address")
	flag.Parse()

	srv := server.NewServer()

	go func() {
		log.Printf("Vector database microservice starting on %s", *addr)
		if err := srv.Start(*addr); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}
