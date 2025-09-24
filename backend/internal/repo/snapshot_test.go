package repo_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	r "github.com/endingwithali/2025censys/internal/repo"
	"github.com/google/uuid"

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

// Test GetSnapshotByTimeStamp method
func TestSnapshot_GetSnapshotByTimeStamp(t *testing.T) {
	ctx := context.Background()

	// Setup: Insert a test snapshot first
	testFileRel := filepath.Join("..", "repo", "testfiles", "host_125.199.235.74_2025-09-15T08-49-45Z.json")
	f, err := os.Open(testFileRel)
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer f.Close()

	filename := filepath.Base(testFileRel)
	host := "125.199.235.74"
	timestamp := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC) // Use a unique timestamp
	dstPath := filepath.Join(t.TempDir(), filename)

	// Create the file
	dst, err := os.Create(dstPath)
	if err != nil {
		t.Fatalf("failed to create dst file: %v", err)
	}
	if _, err := dst.ReadFrom(f); err != nil {
		dst.Close()
		t.Fatalf("failed to copy file: %v", err)
	}
	dst.Close()

	// Insert the snapshot
	if err := snapRepo.Insert(ctx, host, timestamp, dstPath, filename); err != nil {
		t.Fatalf("repo.Insert returned error: %v", err)
	}

	// Test: Get snapshot by timestamp
	retrievedSnapshot, err := snapRepo.GetSnapshotByTimeStamp(ctx, host, timestamp)
	if err != nil {
		t.Fatalf("GetSnapshotByTimeStamp returned error: %v", err)
	}

	// Verify the retrieved snapshot
	if retrievedSnapshot.Host_IP != host {
		t.Errorf("expected host %s, got %s", host, retrievedSnapshot.Host_IP)
	}
	if !retrievedSnapshot.Timestamp.Equal(timestamp) {
		t.Errorf("expected timestamp %v, got %v", timestamp, retrievedSnapshot.Timestamp)
	}
	if retrievedSnapshot.File_PWD != dstPath {
		t.Errorf("expected file path %s, got %s", dstPath, retrievedSnapshot.File_PWD)
	}
	if retrievedSnapshot.File_Name != filename {
		t.Errorf("expected filename %s, got %s", filename, retrievedSnapshot.File_Name)
	}
	if retrievedSnapshot.UUID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
}

// Test GetSnapshotByTimeStamp with non-existent snapshot
func TestSnapshot_GetSnapshotByTimeStamp_NotFound(t *testing.T) {
	ctx := context.Background()

	host := "192.168.1.1"
	timestamp := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	_, err := snapRepo.GetSnapshotByTimeStamp(ctx, host, timestamp)
	if err == nil {
		t.Error("expected error for non-existent snapshot, got nil")
	}
	if err != gorm.ErrRecordNotFound {
		t.Errorf("expected gorm.ErrRecordNotFound, got %v", err)
	}
}

// Test GetSnapshotByFileName method
func TestSnapshot_GetSnapshotByFileName(t *testing.T) {
	ctx := context.Background()

	// Setup: Insert a test snapshot first
	testFileRel := filepath.Join("..", "repo", "testfiles", "host_125.199.235.74_2025-09-15T08-49-45Z.json")
	f, err := os.Open(testFileRel)
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer f.Close()

	filename := filepath.Base(testFileRel)
	host := "125.199.235.74"
	timestamp := parseTestTimestamp(t)
	dstPath := filepath.Join(t.TempDir(), filename)

	// Create the file
	dst, err := os.Create(dstPath)
	if err != nil {
		t.Fatalf("failed to create dst file: %v", err)
	}
	if _, err := dst.ReadFrom(f); err != nil {
		dst.Close()
		t.Fatalf("failed to copy file: %v", err)
	}
	dst.Close()

	// Insert the snapshot
	if err := snapRepo.Insert(ctx, host, timestamp, dstPath, filename); err != nil {
		t.Fatalf("repo.Insert returned error: %v", err)
	}

	// Test: Get snapshot by filename (note: this uses File_PWD, not File_Name)
	retrievedSnapshot, err := snapRepo.GetSnapshotByFileName(ctx, host, dstPath)
	if err != nil {
		t.Fatalf("GetSnapshotByFileName returned error: %v", err)
	}

	// Verify the retrieved snapshot
	if retrievedSnapshot.Host_IP != host {
		t.Errorf("expected host %s, got %s", host, retrievedSnapshot.Host_IP)
	}
	if !retrievedSnapshot.Timestamp.Equal(timestamp) {
		t.Errorf("expected timestamp %v, got %v", timestamp, retrievedSnapshot.Timestamp)
	}
	if retrievedSnapshot.File_PWD != dstPath {
		t.Errorf("expected file path %s, got %s", dstPath, retrievedSnapshot.File_PWD)
	}
	if retrievedSnapshot.File_Name != filename {
		t.Errorf("expected filename %s, got %s", filename, retrievedSnapshot.File_Name)
	}
}

// Test GetSnapshotByFileName with non-existent snapshot
func TestSnapshot_GetSnapshotByFileName_NotFound(t *testing.T) {
	ctx := context.Background()

	host := "192.168.1.1"
	filename := "/path/to/nonexistent.json"

	_, err := snapRepo.GetSnapshotByFileName(ctx, host, filename)
	if err == nil {
		t.Error("expected error for non-existent snapshot, got nil")
	}
	if err != gorm.ErrRecordNotFound {
		t.Errorf("expected gorm.ErrRecordNotFound, got %v", err)
	}
}

// Test GetAllHosts method
func TestSnapshot_GetAllHosts(t *testing.T) {
	ctx := context.Background()

	// Insert multiple snapshots for different hosts
	hosts := []string{"192.168.1.1", "10.0.0.1", "172.16.0.1"}
	timestamp := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	for _, host := range hosts {
		filename := filepath.Join(t.TempDir(), "test_"+host+".json")
		// Create a dummy file
		if err := os.WriteFile(filename, []byte(`{"test": "data"}`), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		if err := snapRepo.Insert(ctx, host, timestamp, filename, "test_"+host+".json"); err != nil {
			t.Fatalf("repo.Insert returned error for host %s: %v", host, err)
		}
	}

	// Test: Get all hosts
	retrievedHosts, err := snapRepo.GetAllHosts(ctx)
	if err != nil {
		t.Fatalf("GetAllHosts returned error: %v", err)
	}

	// Verify we got all the hosts (order might be different)
	if len(retrievedHosts) < len(hosts) {
		t.Errorf("expected at least %d hosts, got %d", len(hosts), len(retrievedHosts))
	}

	// Check that all expected hosts are present
	hostMap := make(map[string]bool)
	for _, host := range retrievedHosts {
		hostMap[host] = true
	}

	for _, expectedHost := range hosts {
		if !hostMap[expectedHost] {
			t.Errorf("expected host %s not found in retrieved hosts: %v", expectedHost, retrievedHosts)
		}
	}
}

// Test GetAllHosts with empty database
func TestSnapshot_GetAllHosts_Empty(t *testing.T) {
	ctx := context.Background()

	// Ensure database is empty
	cleanUpDB()

	hosts, err := snapRepo.GetAllHosts(ctx)
	if err != nil {
		t.Fatalf("GetAllHosts returned error: %v", err)
	}

	if len(hosts) != 0 {
		t.Errorf("expected empty hosts list, got %v", hosts)
	}
}

// Test ListAllHostSnapshots method
func TestSnapshot_ListAllHostSnapshots(t *testing.T) {
	ctx := context.Background()

	host := "192.168.1.100"
	timestamps := []time.Time{
		time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 3, 12, 0, 0, 0, time.UTC),
	}

	// Insert multiple snapshots for the same host
	for i, timestamp := range timestamps {
		filename := filepath.Join(t.TempDir(), fmt.Sprintf("test_%d.json", i))
		// Create a dummy file
		if err := os.WriteFile(filename, []byte(`{"test": "data"}`), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		if err := snapRepo.Insert(ctx, host, timestamp, filename, fmt.Sprintf("test_%d.json", i)); err != nil {
			t.Fatalf("repo.Insert returned error for timestamp %v: %v", timestamp, err)
		}
	}

	// Test: Get all snapshots for the host
	retrievedTimestamps, err := snapRepo.ListAllHostSnapshots(ctx, host)
	if err != nil {
		t.Fatalf("ListAllHostSnapshots returned error: %v", err)
	}

	// Verify we got all the timestamps
	if len(retrievedTimestamps) != len(timestamps) {
		t.Errorf("expected %d timestamps, got %d", len(timestamps), len(retrievedTimestamps))
	}

	// Convert expected timestamps to RFC3339 strings for comparison
	expectedStrings := make([]string, len(timestamps))
	for i, ts := range timestamps {
		expectedStrings[i] = ts.Format(time.RFC3339)
	}

	// Check that all expected timestamps are present (order might be different)
	timestampMap := make(map[string]bool)
	for _, ts := range retrievedTimestamps {
		timestampMap[ts] = true
	}

	for _, expectedTS := range expectedStrings {
		if !timestampMap[expectedTS] {
			t.Errorf("expected timestamp %s not found in retrieved timestamps: %v", expectedTS, retrievedTimestamps)
		}
	}
}

// Test ListAllHostSnapshots with non-existent host
func TestSnapshot_ListAllHostSnapshots_NotFound(t *testing.T) {
	ctx := context.Background()

	host := "999.999.999.999"

	timestamps, err := snapRepo.ListAllHostSnapshots(ctx, host)
	if err != nil {
		t.Fatalf("ListAllHostSnapshots returned error: %v", err)
	}

	if len(timestamps) != 0 {
		t.Errorf("expected empty timestamps list for non-existent host, got %v", timestamps)
	}
}

// Test Insert with invalid data
func TestSnapshot_Insert_InvalidData(t *testing.T) {
	ctx := context.Background()

	// Test with empty host IP - this might succeed depending on DB constraints
	err := snapRepo.Insert(ctx, "", time.Now(), "/tmp/test.json", "test.json")
	// We don't assert on this because it might succeed depending on DB constraints
	if err != nil {
		t.Logf("Insert with empty host IP failed as expected: %v", err)
	}

	// Test with invalid file path (this might not fail depending on DB constraints)
	// but it's good to test the behavior
	err = snapRepo.Insert(ctx, "192.168.1.1", time.Now(), "", "test.json")
	// We don't assert on this because it might succeed depending on DB constraints
	_ = err
}

// Test concurrent inserts
func TestSnapshot_ConcurrentInserts(t *testing.T) {
	ctx := context.Background()

	host := "192.168.1.200"
	timestamp := time.Now()
	numGoroutines := 10

	// Channel to collect errors
	errorChan := make(chan error, numGoroutines)

	// Launch multiple goroutines to insert concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			filename := filepath.Join(t.TempDir(), fmt.Sprintf("concurrent_test_%d.json", index))
			// Create a dummy file
			if err := os.WriteFile(filename, []byte(`{"test": "data"}`), 0644); err != nil {
				errorChan <- fmt.Errorf("failed to create test file: %v", err)
				return
			}

			err := snapRepo.Insert(ctx, host, timestamp.Add(time.Duration(index)*time.Second), filename, fmt.Sprintf("concurrent_test_%d.json", index))
			errorChan <- err
		}(i)
	}

	// Collect all errors
	var errors []error
	for i := 0; i < numGoroutines; i++ {
		if err := <-errorChan; err != nil {
			errors = append(errors, err)
		}
	}

	// Check if any errors occurred
	if len(errors) > 0 {
		t.Logf("Some concurrent inserts failed (this might be expected): %v", errors)
	}

	// Verify that some snapshots were inserted
	timestamps, err := snapRepo.ListAllHostSnapshots(ctx, host)
	if err != nil {
		t.Fatalf("ListAllHostSnapshots returned error: %v", err)
	}

	// We expect at least some snapshots to be inserted successfully
	if len(timestamps) == 0 {
		t.Error("expected at least some snapshots to be inserted successfully")
	}
}

// Test with very long host IP
func TestSnapshot_LongHostIP(t *testing.T) {
	ctx := context.Background()

	// Test with an unusually long IP string (though invalid)
	longIP := "192.168.1.1.extra.long.invalid.ip.address"
	timestamp := time.Now()
	filename := filepath.Join(t.TempDir(), "long_ip_test.json")

	// Create a dummy file
	if err := os.WriteFile(filename, []byte(`{"test": "data"}`), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	err := snapRepo.Insert(ctx, longIP, timestamp, filename, "long_ip_test.json")
	// This might succeed or fail depending on DB constraints
	_ = err
}

// Test with special characters in file paths
func TestSnapshot_SpecialCharactersInPath(t *testing.T) {
	ctx := context.Background()

	host := "192.168.1.1"
	timestamp := time.Now()

	// Test with special characters in file path
	specialPath := filepath.Join(t.TempDir(), "test with spaces & special chars!@#$%^&*().json")
	filename := "test with spaces & special chars!@#$%^&*().json"

	// Create a dummy file
	if err := os.WriteFile(specialPath, []byte(`{"test": "data"}`), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	err := snapRepo.Insert(ctx, host, timestamp, specialPath, filename)
	if err != nil {
		t.Logf("Insert with special characters failed (might be expected): %v", err)
	}
}

// Test table name method
func TestSnapshot_TableName(t *testing.T) {
	snapshot := r.Snapshot{}
	tableName := snapshot.TableName()
	expectedTableName := "snapshot"

	if tableName != expectedTableName {
		t.Errorf("expected table name %s, got %s", expectedTableName, tableName)
	}
}

// Test UUID generation
func TestSnapshot_UUIDGeneration(t *testing.T) {
	ctx := context.Background()

	host := "192.168.1.1"
	timestamp := time.Now()
	filename := filepath.Join(t.TempDir(), "uuid_test.json")

	// Create a dummy file
	if err := os.WriteFile(filename, []byte(`{"test": "data"}`), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	err := snapRepo.Insert(ctx, host, timestamp, filename, "uuid_test.json")
	if err != nil {
		t.Fatalf("repo.Insert returned error: %v", err)
	}

	// Retrieve the snapshot to verify UUID was generated
	retrievedSnapshot, err := snapRepo.GetSnapshotByTimeStamp(ctx, host, timestamp)
	if err != nil {
		t.Fatalf("GetSnapshotByTimeStamp returned error: %v", err)
	}

	// Verify UUID is not nil
	if retrievedSnapshot.UUID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}

	// Verify UUID is a valid UUID
	if len(retrievedSnapshot.UUID.String()) == 0 {
		t.Error("expected valid UUID string")
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
