package server

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/tahcohcat/same-same/internal/embedders"
	"github.com/tahcohcat/same-same/internal/embedders/quotes/gemini"
	"github.com/tahcohcat/same-same/internal/embedders/quotes/huggingface"
	"github.com/tahcohcat/same-same/internal/embedders/quotes/local/tfidf"
	"github.com/tahcohcat/same-same/internal/handlers"
	"github.com/tahcohcat/same-same/internal/storage"
)

type Server struct {
	storage storage.Storage
	handler *handlers.VectorHandler
	router  *mux.Router
}

func NewServer() *Server {
	store, err := storage.NewStorageFromEnv()
	if err != nil {
		log.Fatalf("failed to initialize storage adapter: %v", err)
	}

	handler := handlers.NewVectorHandler(store, CreateEmbedder(os.Getenv("EMBEDDER_TYPE")))
	router := mux.NewRouter()

	server := &Server{
		storage: store,
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
	api.HandleFunc("/search", s.handler.AdvancedSearch).Methods("POST")
	// api.HandleFunc("/search/temporal", s.handler.TemporalSearch).Methods("POST") // Temporal-aware search (TODO: implement)

	api.HandleFunc("/embedder/stats", s.handler.GetEmbedderStats).Methods("GET")
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

func CreateEmbedder(eType string) embedders.Embedder {

	switch eType {
	case "gemini":
		googleAPIKey := os.Getenv("GEMINI_API_KEY")
		if googleAPIKey == "" {
			log.Fatal("GEMINI_API_KEY environment variable is required")
		}
		return gemini.NewGeminiEmbedder(googleAPIKey)
	case "huggingface":
		hfAPIKey := os.Getenv("HUGGINGFACE_API_KEY")
		if hfAPIKey == "" {
			log.Fatal("HUGGINGFACE_API_KEY environment variable is required")
		}
		return huggingface.NewHuggingFaceEmbedder(hfAPIKey)
	default:
		return tfidf.NewTFIDFEmbedder()
	}
}
