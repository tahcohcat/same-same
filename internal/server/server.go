package server

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github/tahcohcat/same-same/internal/handlers"
	"github/tahcohcat/same-same/internal/storage"
)

type Server struct {
	storage *storage.MemoryStorage
	handler *handlers.VectorHandler
	router  *mux.Router
}

func NewServer() *Server {
	storage := storage.NewMemoryStorage()
	handler := handlers.NewVectorHandler(storage)
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

	api.HandleFunc("/vectors", s.handler.CreateVector).Methods("POST")
	api.HandleFunc("/vectors", s.handler.ListVectors).Methods("GET")
	api.HandleFunc("/vectors/{id}", s.handler.GetVector).Methods("GET")
	api.HandleFunc("/vectors/{id}", s.handler.UpdateVector).Methods("PUT")
	api.HandleFunc("/vectors/{id}", s.handler.DeleteVector).Methods("DELETE")
	api.HandleFunc("/vectors/search", s.handler.SearchVectors).Methods("POST")

	s.router.HandleFunc("/health", s.healthCheck).Methods("GET")
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy"}`))
}

func (s *Server) Start(addr string) error {
	log.Printf("Starting server on %s", addr)
	return http.ListenAndServe(addr, s.router)
}