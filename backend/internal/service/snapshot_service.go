package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/endingwithali/2025censys/internal/repo"
)

type SnapshotService struct {
	snapshotRepo repo.SnapshotRepo
	FileLocation string
}

func NewSnapshotService(snapshotRepo repo.SnapshotRepo, fileLocation string) *SnapshotService {
	return &SnapshotService{
		snapshotRepo: snapshotRepo,
		FileLocation: fileLocation,
	}
}

func (service *SnapshotService) CreateSnapshot(ctx context.Context, file multipart.File, filename string) error {
	hostIP, timestamp, err := service.parseFileName(filename)
	if err != nil {
		return fmt.Errorf("Failed to parse file name: %s", err.Error())
	}

	filepath := filepath.Join(service.FileLocation, filename)
	if _, err := os.Stat(filepath); err == nil {
		return fmt.Errorf("Attempting to add duplicate file for host: %s", filename)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("Failed to create file: %v", err.Error())
	}

	log.Println("File Path", filepath)

	/*
		os.OpenFile: low-level open with flags + permissions.
		dstPath: where you want to create the file.
		Flags:
			os.O_CREATE — create the file if it doesn’t exist.
			os.O_WRONLY — open write-only (you’ll use dst.Write / io.Copy).
			os.O_EXCL — exclusive create: if the path already exists, the open fails with EEXIST (no overwrite, no truncate). This is the safety switch that prevents clobbering an existing file.
			0o644 (octal) — file mode on Unix: rw-r--r--\
				Owner: read+write
				Group: read
				Others: read
			(Subject to the process umask, so the actual bits may be more restrictive.)

		ChatGPT ADDITION: these writes are not atomic - GPT says to use temp files to write to, then rename to final path using fsync
	*/
	dst, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("Failed to write file to disk: %v", err.Error())
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		_ = os.RemoveAll(filepath)
		return fmt.Errorf("Failed to write contents of file to file on OS: %v", err.Error())
	}

	err = service.snapshotRepo.Insert(ctx, hostIP, timestamp, filepath, filename)
	if err != nil {
		_ = os.RemoveAll(filepath)
		return fmt.Errorf("Failed to write file to DB: %v", err.Error())
	}
	return nil
}

func (service *SnapshotService) GetSnapshotByTimestamp(ctx context.Context, host_ip string, timestampString string) (string, error) {
	timestamp, err := time.Parse(time.RFC3339, timestampString)
	if err != nil {
		return "", fmt.Errorf("Incorrectly formatted timestamp string")
	}
	snapshot, err := service.snapshotRepo.GetSnapshotByTimeStamp(ctx, host_ip, timestamp)
	if err != nil {
		return "", err
	}
	return snapshot.File_PWD, nil
}

func (service *SnapshotService) GetAllHosts(ctx context.Context) ([]string, error) {
	return service.snapshotRepo.GetAllHosts(ctx)
}

func (service *SnapshotService) ListAllSnapshotsForHost(ctx context.Context, host_ip string) ([]string, error) {
	return service.snapshotRepo.ListAllHostSnapshots(ctx, host_ip)
}

func (service *SnapshotService) parseFileName(filename string) (string, time.Time, error) {
	// host_<ip>_<timestamp>.json
	// timestamp is file-safe ISO: 2006-01-02T15-04-05Z (colons replaced by dashes)
	var re = regexp.MustCompile(`^host_([0-9]{1,3}(?:\.[0-9]{1,3}){3})_([0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}-[0-9]{2}-[0-9]{2}Z)\.json$`)

	m := re.FindStringSubmatch(filename)
	if m == nil {
		return "", time.Time{}, fmt.Errorf("filename %q does not match expected pattern host_<ip>_<timestamp>.json", filename)
	}

	ipStr := m[1]
	if net.ParseIP(ipStr) == nil {
		return "", time.Time{}, fmt.Errorf("invalid IPv4 address in filename: %q", ipStr)
	}

	tsStr := m[2]

	// Parse the file-safe ISO timestamp (with dashes in the clock) into time.Time.
	// Layout matches: 2025-09-10T03-00-00Z
	t, err := time.Parse("2006-01-02T15-04-05Z", tsStr)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("invalid timestamp in filename: %q: %w", tsStr, err)
	}

	return ipStr, t, nil
}
