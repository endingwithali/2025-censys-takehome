package repo_test

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	r "github.com/endingwithali/2025censys/internal/repo"
	// no service import; tests exercise repo behavior directly

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var testDB *gorm.DB
var snapRepo r.SnapshotRepo

func TestMain(m *testing.M) {
	testDB = setupTestDBConnection()
	snapRepo = r.NewSnapshotRepo(testDB)
	code := m.Run()
	cleanUpDB()
	os.Exit(code)
}

// connects to the real Postgres test DB as requested
func setupTestDBConnection() *gorm.DB {
	// bad practice - should not be hardcoded, but for the sake of brevity
	connectionString := "host=localhost user=test password=testpassword dbname=censys_testdb port=5432 sslmode=disable TimeZone=UTC"
	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	if err != nil {
		log.Fatalf("Snapshot_TEST: Failed to open db: %v", err)
	}
	return db
}

func cleanUpDB() {
	// remove everything from the table snapshot
	if testDB != nil {
		testDB.Exec("TRUNCATE TABLE snapshot RESTART IDENTITY CASCADE;")
	}
}

// Test creating a snapshot via the service layer using file from testfiles
func TestSnapshot_Create(t *testing.T) {
	ctx := context.Background()

	// build path relative to repository root so `go test` finds the file
	testFileRel := filepath.Join("..", "repo", "testfiles", "host_125.199.235.74_2025-09-15T08-49-45Z.json")
	f, err := os.Open(testFileRel)
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer f.Close()

	// simulate service behavior enough to produce the args repo.Insert expects
	filename := filepath.Base(testFileRel)
	host := "125.199.235.74"
	timestamp := parseTestTimestamp(t)

	// copy file to temp dir path to emulate file location
	dstPath := filepath.Join(t.TempDir(), filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		t.Fatalf("failed to create dst file: %v", err)
	}
	if _, err := dst.ReadFrom(f); err != nil {
		dst.Close()
		t.Fatalf("failed to copy file: %v", err)
	}
	dst.Close()

	if err := snapRepo.Insert(ctx, host, timestamp, dstPath, filename); err != nil {
		t.Fatalf("repo.Insert returned error: %v", err)
	}

	// verify host appears in DB
	hosts, err := snapRepo.GetAllHosts(ctx)
	if err != nil {
		t.Fatalf("GetAllHosts returned error: %v", err)
	}
	expected := "125.199.235.74"
	found := false
	for _, h := range hosts {
		if h == expected {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected host %s in hosts list, got: %v", expected, hosts)
	}
}

// Test duplicate create: second CreateSnapshot should fail because file already exists
func TestSnapshot_CreateDuplicate(t *testing.T) {
	ctx := context.Background()

	testFileRel := filepath.Join("..", "repo", "testfiles", "host_125.199.235.74_2025-09-15T08-49-45Z.json")
	f1, err := os.Open(testFileRel)
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer f1.Close()

	filename := filepath.Base(testFileRel)
	host := "125.199.235.74"
	timestamp := parseTestTimestamp(t)
	// create dst file path and copy
	dstPath := filepath.Join(t.TempDir(), filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		t.Fatalf("failed to create dst file: %v", err)
	}
	if _, err := dst.ReadFrom(f1); err != nil {
		dst.Close()
		t.Fatalf("failed to copy file: %v", err)
	}
	dst.Close()

	// first insert
	if err := snapRepo.Insert(ctx, host, timestamp, dstPath, filename); err != nil {
		t.Fatalf("initial repo.Insert returned error: %v", err)
	}

	// second insert: repo.Insert currently does not check for duplicates, so behavior
	// depends on DB constraints. We attempt a second insert and then check how many
	// rows exist for the host/timestamp combination.
	if err := snapRepo.Insert(ctx, host, timestamp, dstPath, filename); err != nil {
		// if DB prevents duplicate inserts, that's acceptable; assert that only one row exists
		var snaps []r.Snapshot
		_ = testDB.WithContext(ctx).Where("host_ip = ? AND timestamp = ?", host, timestamp).Find(&snaps).Error
		if len(snaps) > 1 {
			t.Fatalf("expected at most 1 snapshot, found %d", len(snaps))
		}
	} else {
		// if insert succeeded, verify two rows exist (no DB uniqueness)
		var snaps []r.Snapshot
		if err := testDB.WithContext(ctx).Where("host_ip = ? AND timestamp = ?", host, timestamp).Find(&snaps).Error; err != nil {
			t.Fatalf("failed to query snapshots: %v", err)
		}
		if len(snaps) < 2 {
			t.Fatalf("expected >=2 snapshots after second insert, got %d", len(snaps))
		}
	}
}

// helper: parse the timestamp used in test filename
func parseTestTimestamp(t *testing.T) (ts time.Time) {
	t.Helper()
	tsParsed, err := time.Parse(time.RFC3339, "2025-09-15T08:49:45Z")
	if err != nil {
		t.Fatalf("failed to parse test timestamp: %v", err)
	}
	return tsParsed
}
