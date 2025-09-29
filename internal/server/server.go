package server

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github/tahcohcat/same-same/internal/embedders"
	"github/tahcohcat/same-same/internal/embedders/quotes/gemini"
	"github/tahcohcat/same-same/internal/embedders/quotes/huggingface"
	"github/tahcohcat/same-same/internal/embedders/quotes/local/tfidf"
	"github/tahcohcat/same-same/internal/handlers"
	"github/tahcohcat/same-same/internal/storage/memory"
)

type Server struct {
	storage *memory.Storage
	handler *handlers.VectorHandler
	router  *mux.Router
}

func NewServer() *Server {
	storage := memory.NewStorage()

	embedderType := os.Getenv("EMBEDDER_TYPE")
	var embedder embedders.Embedder

	switch embedderType {
	case "gemini":
		googleAPIKey := os.Getenv("GEMINI_API_KEY")
		if googleAPIKey == "" {
			log.Fatal("GEMINI_API_KEY environment variable is required")
		}
		embedder = gemini.NewGeminiEmbedder(googleAPIKey)
	case "huggingface":
		hfAPIKey := os.Getenv("HUGGINGFACE_API_KEY")
		if hfAPIKey == "" {
			log.Fatal("HUGGINGFACE_API_KEY environment variable is required")
		}
		embedder = huggingface.NewHuggingFaceEmbedder(hfAPIKey)
	default:
		embedder = tfidf.NewTFIDFEmbedder()
	}

	handler := handlers.NewVectorHandler(storage, embedder)
	router := mux.NewRouter()

	server := &Server{
		storage: storage,
		handler: handler,
		router:  router,
	}

	server.setupRoutes()
	return server
}

func (s *Server) setupRoutes() {
	api := s.router.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/vectors/embed", s.handler.EmbedVector).Methods("POST")
	api.HandleFunc("/vectors/count", s.handler.CountVectors).Methods("GET")
	api.HandleFunc("/vectors", s.handler.CreateVector).Methods("POST")
	api.HandleFunc("/vectors", s.handler.ListVectors).Methods("GET")
	api.HandleFunc("/vectors/metadata", s.handler.ListVectorMetadata).Methods("GET")
	api.HandleFunc("/vectors/{id}", s.handler.GetVector).Methods("GET")
	api.HandleFunc("/vectors/{id}", s.handler.UpdateVector).Methods("PUT")
	api.HandleFunc("/vectors/{id}", s.handler.DeleteVector).Methods("DELETE")
	api.HandleFunc("/vectors/search", s.handler.SearchVectors).Methods("POST")
	api.HandleFunc("/search", s.handler.SearchByText).Methods("POST")

	s.router.HandleFunc("/health", s.healthCheck).Methods("GET")
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "healthy"}`))
}

func (s *Server) Start(addr string) error {
	log.Printf("starting server on :%s", addr)
	return http.ListenAndServe(addr, s.router)
}
