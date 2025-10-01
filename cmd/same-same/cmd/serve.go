package cmd

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tahcohcat/same-same/internal/server"
)

var (
	// Serve-specific flags
	addr  string
	debug bool
)

func init() {
	rootCmd.AddCommand(serveCmd)

	// Serve flags
	serveCmd.Flags().StringVarP(&addr, "addr", "a", ":8080", "HTTP service address")
	serveCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the vector database server",
	Long: `Start the Same-Same vector database HTTP server.

The server provides a RESTful API for:
  - Storing and retrieving vectors
  - Similarity search using cosine similarity
  - Automatic embedding generation
  - Metadata filtering and advanced search

By default, the server uses in-memory storage and a local TF-IDF embedder.
You can configure different embedders using environment variables.`,
	Example: `  # Start server on default port 8080
  same-same serve

  # Start on custom port
  same-same serve -a :9000

  # Enable debug logging
  same-same serve -d

  # Use Gemini embedder
  export GEMINI_API_KEY=your_key
  export EMBEDDER_TYPE=gemini
  same-same serve`,
	Run: runServe,
}

func runServe(cmd *cobra.Command, args []string) {
	// Setup logging
	if debug || verbose {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("debug logging enabled")
	}

	// Create and start server
	srv := server.NewServer()

	go func() {
		log.Printf("vector database microservice starting on %s", addr)
		if err := srv.Start(addr); err != nil {
			log.Fatalf("server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
}
