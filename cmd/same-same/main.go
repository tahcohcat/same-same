package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tahcohcat/same-same/internal/server"

	"github.com/sirupsen/logrus"
)

func main() {
	var addr = flag.String("addr", ":8080", "HTTP service address")
	var debug = flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	if *debug {
		// Enable debug logging
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("debug logging enabled")
	}

	srv := server.NewServer()

	go func() {
		log.Printf("vector database microservice starting on %s", *addr)
		if err := srv.Start(*addr); err != nil {
			log.Fatalf("server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
}
