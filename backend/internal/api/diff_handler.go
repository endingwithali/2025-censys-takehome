package api

import (
	"encoding/json"
	"net/http"
)

type diffResponse struct {
	DiffStatus  string
	Differences string
}

// GetSnapshotDiffs handles GET /api/snapshot/diff
//
// Summary: Get snapshots differences for a host.
// Path Params:
//   - host: string (IPv4/IPv6)
//   - t1: string (timestamp of file 1)
//   - t2: string (timestamp of file 2)
//
// Responses:
//   - 200: ListSnapshotsResponse
//   - 204: APIError (Snapshot not found in DB or on disk)
//   - 500: API Error (Unable to create difference)
func (server *Server) GetSnapshotDiffs(w http.ResponseWriter, r *http.Request) {
	host_ip := r.URL.Query().Get("ip")
	t1 := r.URL.Query().Get("t1")
	t2 := r.URL.Query().Get("t2")
	if host_ip == "" || t1 == "" || t2 == "" {
		http.Error(w, "Error: No host ip or timestamps defined", http.StatusNotAcceptable)
		return
	}
	ctx := r.Context()
	// CHOICE: Don't optimize for case where t1 == t2. can skip all the matching and then just load file values from memory via snapshot

	file1Location, err := server.snapshotService.GetSnapshotByTimestamp(ctx, host_ip, t1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}
	file2Location, err := server.snapshotService.GetSnapshotByTimestamp(ctx, host_ip, t2)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}

	status, difference, err := server.differenceService.GetDifferences(file1Location, file2Location)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(diffResponse{
		DiffStatus:  status,
		Differences: difference,
	})
}
