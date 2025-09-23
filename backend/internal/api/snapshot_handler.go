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

// Expecting URL enconded timestamp
// GET /snapshots?host=125.199.235.74&at=2025-09-10T03:00:00Z
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

// GET /snapshots?host=125.199.235.74
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
