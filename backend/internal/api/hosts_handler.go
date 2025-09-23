package api

import (
	"encoding/json"
	"log"
	"net/http"
)

func (server *Server) ListAllHosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	hosts, err := server.snapshotService.GetAllHosts(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Printf("ListAllHosts: Found a total of %d hosts", len(hosts))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(hosts)
}
