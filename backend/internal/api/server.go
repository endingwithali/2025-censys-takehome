package api

import (
	"log"
	"net/http"

	"github.com/endingwithali/2025censys/internal/service"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
)

type Server struct {
	snapshotService   *service.SnapshotService
	differenceService *service.DifferencesService
	MaxFileSize       int
}

func New(snapshotService *service.SnapshotService, differenceService *service.DifferencesService, maxFileSize int) http.Handler {
	server := &Server{
		snapshotService:   snapshotService,
		differenceService: differenceService,
		MaxFileSize:       maxFileSize,
	}
	router := chi.NewRouter()

	// CORS middleware
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Global fallbacks
	router.NotFound(server.NotFound)
	router.MethodNotAllowed(server.MethodNotAllowed)

	// API Routes
	router.Route("/api", func(r chi.Router) {
		r.Get("/health", server.Get)
		r.Get("/host/all", server.ListAllHosts)
		r.Get("/host", server.GetAllSnapshotsForHost)
		r.Get("/snapshot", server.GetSnapshotForHost)
		r.Post("/snapshot", server.CreateSnapshot)
		r.Get("/snapshot/diff", server.GetSnapshotDiffs)
	})
	return router
}

// Health Check
// GET /api/health
//
// Summary: Check if the server is running.
//
// Responses:
//   - 200: OK
//   - 500: Internal Server Error
func (server *Server) Get(w http.ResponseWriter, r *http.Request) {
	log.Print("In Health Check")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("All Connected!"))
}

func (server *Server) MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	log.Print("In MethodNotAllowed")
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte("Method Not Allowed"))
}

func (server *Server) NotFound(w http.ResponseWriter, r *http.Request) {
	log.Print("In Not Found")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not Found"))
}
