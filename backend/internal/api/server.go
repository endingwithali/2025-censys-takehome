package api

import (
	"log"
	"net/http"

	"github.com/endingwithali/2025censys/internal/service"
	"github.com/go-chi/chi"
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

	// Global fallbacks
	router.NotFound(server.NotFound)
	router.MethodNotAllowed(server.MethodNotAllowed)

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
