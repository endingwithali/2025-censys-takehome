package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

// GetSnapshotForHost handles GET /api/snapshot?host={host}&at={timestamp}
//
// Summary: Get snapshot at specific timestamp for a host.
// Path Params:
//   - host: string (IPv4/IPv6)
//   - at: string (timestamp of file 1)
//
// Example:
// GET /snapshots?host=125.199.235.74&at=2025-09-10T03:00:00Z
//
// Responses:
//   - 200: ListSnapshotsResponse
//   - 204: APIError (Snapshot not found in DB or on disk)
//   - 500: API Error (Unable to create difference)
//
// Response Body:
//
//	{
//		JSON string of contents of snapshot file
//	}
func (server *Server) GetSnapshotForHost(w http.ResponseWriter, r *http.Request) {
	log.Println("GetSnapshotForHost: CALLED")

	host_ip := r.URL.Query().Get("ip")
	timestamp := r.URL.Query().Get("at")
	if host_ip == "" || timestamp == "" {
		log.Println("GetSnapshotForHost: FAILED")
		http.Error(w, "Error: No host ip or timestamp defined", http.StatusNotAcceptable)
		return
	}
	ctx := r.Context()

	log.Println("Getting Snapshots for Host", host_ip, timestamp)

	snapshotPath, err := server.snapshotService.GetSnapshotByTimestamp(ctx, host_ip, timestamp)
	if err != nil {
		log.Println("GetSnapshotForHost: FAILED")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	file, err := os.Open(snapshotPath)
	if err != nil {
		log.Println("GetSnapshotForHost: FAILED")
		http.Error(w, "Unable to read file from disk: "+err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, file)
	log.Println("GetSnapshotForHost: SUCCESS")
}

// GetAllSnapshotsForHost handles GET /api/host?host={host}
//
// Summary: Get all timestamps of all snapshots available for a host.
// Path Params:
//   - host: string (IPv4/IPv6)
//
// Example:
// GET /api/host?host=125.199.235.74
//
// Responses:
//   - 200: ListSnapshotsResponse
//   - 204: APIError (Snapshot not found in DB or on disk)
//   - 500: API Error (Unable to create difference)
//
// Response Body:
//
//	{
//	 	[List of all timestamps for the snapshots available for a host as strings]
//	}
func (server *Server) GetAllSnapshotsForHost(w http.ResponseWriter, r *http.Request) {
	log.Println("GetAllSnapshotsForHost: CALLED")

	host_ip := r.URL.Query().Get("ip")
	log.Println("Getting All Snapshots for Host", host_ip)
	if host_ip == "" {
		log.Println("GetAllSnapshotsForHost: Failed")
		http.Error(w, "No Host_IP found", http.StatusNotAcceptable)
		return
	}

	ctx := r.Context()

	availableSnapshots, err := server.snapshotService.ListAllSnapshotsForHost(ctx, host_ip)
	if err != nil {
		log.Println("GetAllSnapshotsForHost: Failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(availableSnapshots)
	log.Println("GetAllSnapshotsForHost: Success")
}

// CreateSnapshot handles Post /api/snapshot
//
// Summary: Create a snapshot for a host.
// Body Params:
//   - file: string (JSON String of body of snapshot file)
//
// Example:
// POST /api/snapshot
// Content-Type: multipart/form-data
// Body:
//
//	{
//		 	file: (File)
//	}
//
// Responses:
//   - 200: Success
//   - 209: API Error (Snapshot failed to be created by DB)
//   - 400: API Error (Invalid file format)
//   - 500: Server Error (Unable to create snapshot)
func (server *Server) CreateSnapshot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.Println("CreateSnapshot: CALLED")

	//Read the file from the http request
	r.Body = http.MaxBytesReader(w, r.Body, int64(server.MaxFileSize))

	file, header, err := r.FormFile("file")
	if err != nil {
		log.Println("CreateSnapshot: FAILED")
		http.Error(w, "Missing File under form field 'file': "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := filepath.Base(header.Filename)
	if !server.validateFileNameFormat(filename) {
		log.Println("CreateSnapshot: FAILED")
		http.Error(w, "expected host_<ip>_<YYYY-MM-DD>T<HH-MM-SS>[.fraction](Z|±HH-MM).json", http.StatusBadRequest)
		return
	}
	err = server.snapshotService.CreateSnapshot(ctx, file, filename)
	if err != nil {
		log.Println("CreateSnapshot: FAILED")
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusOK)
	log.Println("CreateSnapshot: SUCCESS")
}

// validateFileNameFormat validates the filename format to the expected format
//
// Summary: Validate the filename format
// Expected Format:
//
//	host_<ip>_<YYYY-MM-DD>T<HH-MM-SS>[.fraction](Z|±HH-MM).json
//	- Timestamp will be ISO 8601 format
//
// Example:
// validateFileNameFormat("host_125.199.235.74_2025-09-10T03-00-00.json")
//
// Returns:
//   - bool: true if the filename format is valid, false otherwise
func (server *Server) validateFileNameFormat(filename string) bool {
	fileNameRegex := regexp.MustCompile(
		`^host_` +
			`((?:\d{1,3}\.){3}\d{1,3})_` + // IPv4
			`(\d{4}-\d{2}-\d{2})T` + // date
			`(\d{2})-(\d{2})-(\d{2})` + // time HH-MM-SS (dashes instead of colons)
			`(\.\d+)?` + // optional fractional seconds
			`(Z|[+\-]\d{2}-\d{2})` + // Z or ±HH-MM (dash instead of colon)
			`\.json$`,
	)
	matches := fileNameRegex.FindStringSubmatch(filename)
	if matches == nil {
		return false
	}
	return true
}
