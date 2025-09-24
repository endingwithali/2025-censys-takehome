package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/endingwithali/2025censys/internal/repo"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockSnapshotRepo implements the SnapshotRepo interface for testing
type MockSnapshotRepo struct {
	mock.Mock
}

func (m *MockSnapshotRepo) Insert(ctx context.Context, host_ip string, timestamp time.Time, file_pwd string, file_name string) error {
	args := m.Called(ctx, host_ip, timestamp, file_pwd, file_name)
	return args.Error(0)
}

func (m *MockSnapshotRepo) GetSnapshotByTimeStamp(ctx context.Context, host_ip string, timestamp time.Time) (repo.Snapshot, error) {
	args := m.Called(ctx, host_ip, timestamp)
	return args.Get(0).(repo.Snapshot), args.Error(1)
}

func (m *MockSnapshotRepo) GetSnapshotByFileName(ctx context.Context, host_ip string, filename string) (repo.Snapshot, error) {
	args := m.Called(ctx, host_ip, filename)
	return args.Get(0).(repo.Snapshot), args.Error(1)
}

func (m *MockSnapshotRepo) GetAllHosts(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockSnapshotRepo) ListAllHostSnapshots(ctx context.Context, host_ip string) ([]string, error) {
	args := m.Called(ctx, host_ip)
	return args.Get(0).([]string), args.Error(1)
}

// Helper function to create a multipart file for testing
func createMultipartFile(content string) multipart.File {
	reader := strings.NewReader(content)
	return &mockMultipartFile{
		reader: reader,
	}
}

type mockMultipartFile struct {
	reader io.Reader
}

func (m *mockMultipartFile) Read(p []byte) (n int, err error) {
	return m.reader.Read(p)
}

func (m *mockMultipartFile) ReadAt(p []byte, off int64) (n int, err error) {
	// Not implemented for this test
	return 0, fmt.Errorf("ReadAt not implemented")
}

func (m *mockMultipartFile) Seek(offset int64, whence int) (int64, error) {
	// Not implemented for this test
	return 0, fmt.Errorf("Seek not implemented")
}

func (m *mockMultipartFile) Close() error {
	return nil
}

func TestNewSnapshotService(t *testing.T) {
	mockRepo := &MockSnapshotRepo{}
	fileLocation := "/tmp/test"

	service := NewSnapshotService(mockRepo, fileLocation)

	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.snapshotRepo)
	assert.Equal(t, fileLocation, service.FileLocation)
}

func TestSnapshotService_CreateSnapshot(t *testing.T) {
	tests := []struct {
		name           string
		filename       string
		fileContent    string
		expectedStatus error
		repoError      error
		fileExists     bool
		expectedIP     string
		expectedTime   time.Time
	}{
		{
			name:           "successful snapshot creation",
			filename:       "host_192.168.1.1_2025-01-01T12-00-00Z.json",
			fileContent:    `{"test": "data"}`,
			expectedStatus: nil,
			repoError:      nil,
			fileExists:     false,
			expectedIP:     "192.168.1.1",
			expectedTime:   time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			name:           "invalid filename format",
			filename:       "invalid.json",
			fileContent:    `{"test": "data"}`,
			expectedStatus: fmt.Errorf("Failed to parse file name"),
			repoError:      nil,
			fileExists:     false,
			expectedIP:     "",
			expectedTime:   time.Time{},
		},
		{
			name:           "duplicate file",
			filename:       "host_192.168.1.1_2025-01-01T12-00-00Z.json",
			fileContent:    `{"test": "data"}`,
			expectedStatus: fmt.Errorf("Attempting to add duplicate file for host"),
			repoError:      nil,
			fileExists:     true,
			expectedIP:     "",
			expectedTime:   time.Time{},
		},
		{
			name:           "repository error",
			filename:       "host_192.168.1.1_2025-01-01T12-00-00Z.json",
			fileContent:    `{"test": "data"}`,
			expectedStatus: fmt.Errorf("Failed to write file to DB"),
			repoError:      fmt.Errorf("database error"),
			fileExists:     false,
			expectedIP:     "192.168.1.1",
			expectedTime:   time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tempDir := t.TempDir()
			mockRepo := &MockSnapshotRepo{}
			service := NewSnapshotService(mockRepo, tempDir)
			ctx := context.Background()

			// Create existing file if needed
			if tt.fileExists {
				existingFilePath := filepath.Join(tempDir, tt.filename)
				err := os.WriteFile(existingFilePath, []byte("existing content"), 0644)
				require.NoError(t, err)
			}

			// Setup mock expectations for successful cases
			if tt.expectedStatus == nil || (tt.repoError != nil && tt.expectedIP != "") {
				expectedFilePath := filepath.Join(tempDir, tt.filename)
				mockRepo.On("Insert", ctx, tt.expectedIP, tt.expectedTime, expectedFilePath, tt.filename).Return(tt.repoError)
			}

			// Test
			file := createMultipartFile(tt.fileContent)
			err := service.CreateSnapshot(ctx, file, tt.filename)

			// Assertions
			if tt.expectedStatus != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedStatus.Error())
			} else {
				require.NoError(t, err)

				// Verify file was created
				expectedFilePath := filepath.Join(tempDir, tt.filename)
				assert.FileExists(t, expectedFilePath)

				// Verify file content
				content, err := os.ReadFile(expectedFilePath)
				require.NoError(t, err)
				assert.Equal(t, tt.fileContent, string(content))
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSnapshotService_GetSnapshotByTimestamp(t *testing.T) {
	tests := []struct {
		name           string
		hostIP         string
		timestampStr   string
		expectedPath   string
		expectedStatus error
		repoError      error
	}{
		{
			name:           "successful retrieval",
			hostIP:         "192.168.1.1",
			timestampStr:   "2025-01-01T12:00:00Z",
			expectedPath:   "/path/to/snapshot.json",
			expectedStatus: nil,
			repoError:      nil,
		},
		{
			name:           "invalid timestamp format",
			hostIP:         "192.168.1.1",
			timestampStr:   "invalid-timestamp",
			expectedPath:   "",
			expectedStatus: fmt.Errorf("Incorrectly formatted timestamp string"),
			repoError:      nil,
		},
		{
			name:           "snapshot not found",
			hostIP:         "192.168.1.1",
			timestampStr:   "2025-01-01T12:00:00Z",
			expectedPath:   "",
			expectedStatus: fmt.Errorf("snapshot not found"),
			repoError:      fmt.Errorf("snapshot not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := &MockSnapshotRepo{}
			service := NewSnapshotService(mockRepo, "/tmp")
			ctx := context.Background()

			// Setup mock expectations
			if tt.expectedStatus == nil || tt.repoError != nil {
				expectedTime, err := time.Parse(time.RFC3339, tt.timestampStr)
				require.NoError(t, err)

				mockSnapshot := repo.Snapshot{
					UUID:      uuid.New(),
					Host_IP:   tt.hostIP,
					Timestamp: expectedTime,
					File_PWD:  tt.expectedPath,
					File_Name: "snapshot.json",
				}

				mockRepo.On("GetSnapshotByTimeStamp", ctx, tt.hostIP, expectedTime).Return(mockSnapshot, tt.repoError)
			}

			// Test
			result, err := service.GetSnapshotByTimestamp(ctx, tt.hostIP, tt.timestampStr)

			// Assertions
			if tt.expectedStatus != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedStatus.Error())
				assert.Equal(t, "", result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedPath, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSnapshotService_GetAllHosts(t *testing.T) {
	tests := []struct {
		name           string
		expectedHosts  []string
		expectedStatus error
		repoError      error
	}{
		{
			name:           "successful retrieval",
			expectedHosts:  []string{"192.168.1.1", "10.0.0.1"},
			expectedStatus: nil,
			repoError:      nil,
		},
		{
			name:           "empty hosts list",
			expectedHosts:  []string{},
			expectedStatus: nil,
			repoError:      nil,
		},
		{
			name:           "repository error",
			expectedHosts:  nil,
			expectedStatus: fmt.Errorf("database error"),
			repoError:      fmt.Errorf("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := &MockSnapshotRepo{}
			service := NewSnapshotService(mockRepo, "/tmp")
			ctx := context.Background()

			mockRepo.On("GetAllHosts", ctx).Return(tt.expectedHosts, tt.repoError)

			// Test
			result, err := service.GetAllHosts(ctx)

			// Assertions
			if tt.expectedStatus != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedStatus.Error())
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedHosts, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSnapshotService_ListAllSnapshotsForHost(t *testing.T) {
	tests := []struct {
		name              string
		hostIP            string
		expectedSnapshots []string
		expectedStatus    error
		repoError         error
	}{
		{
			name:              "successful retrieval",
			hostIP:            "192.168.1.1",
			expectedSnapshots: []string{"2025-01-01T12:00:00Z", "2025-01-02T12:00:00Z"},
			expectedStatus:    nil,
			repoError:         nil,
		},
		{
			name:              "empty snapshots list",
			hostIP:            "192.168.1.1",
			expectedSnapshots: []string{},
			expectedStatus:    nil,
			repoError:         nil,
		},
		{
			name:              "repository error",
			hostIP:            "192.168.1.1",
			expectedSnapshots: nil,
			expectedStatus:    fmt.Errorf("database error"),
			repoError:         fmt.Errorf("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := &MockSnapshotRepo{}
			service := NewSnapshotService(mockRepo, "/tmp")
			ctx := context.Background()

			mockRepo.On("ListAllHostSnapshots", ctx, tt.hostIP).Return(tt.expectedSnapshots, tt.repoError)

			// Test
			result, err := service.ListAllSnapshotsForHost(ctx, tt.hostIP)

			// Assertions
			if tt.expectedStatus != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedStatus.Error())
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedSnapshots, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSnapshotService_parseFileName(t *testing.T) {
	tests := []struct {
		name          string
		filename      string
		expectedIP    string
		expectedTime  time.Time
		expectedError string
	}{
		{
			name:          "valid filename",
			filename:      "host_192.168.1.1_2025-01-01T12-00-00Z.json",
			expectedIP:    "192.168.1.1",
			expectedTime:  time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			expectedError: "",
		},
		{
			name:          "valid filename with different IP",
			filename:      "host_10.0.0.1_2023-12-25T23-59-59Z.json",
			expectedIP:    "10.0.0.1",
			expectedTime:  time.Date(2023, 12, 25, 23, 59, 59, 0, time.UTC),
			expectedError: "",
		},
		{
			name:          "invalid filename - wrong format",
			filename:      "invalid.json",
			expectedIP:    "",
			expectedTime:  time.Time{},
			expectedError: "does not match expected pattern",
		},
		{
			name:          "invalid filename - wrong prefix",
			filename:      "file_192.168.1.1_2025-01-01T12-00-00Z.json",
			expectedIP:    "",
			expectedTime:  time.Time{},
			expectedError: "does not match expected pattern",
		},
		{
			name:          "invalid filename - wrong extension",
			filename:      "host_192.168.1.1_2025-01-01T12-00-00Z.txt",
			expectedIP:    "",
			expectedTime:  time.Time{},
			expectedError: "does not match expected pattern",
		},
		{
			name:          "invalid filename - invalid IP",
			filename:      "host_999.999.999.999_2025-01-01T12-00-00Z.json",
			expectedIP:    "",
			expectedTime:  time.Time{},
			expectedError: "invalid IPv4 address",
		},
		{
			name:          "invalid filename - invalid timestamp",
			filename:      "host_192.168.1.1_2025-13-01T12-00-00Z.json",
			expectedIP:    "",
			expectedTime:  time.Time{},
			expectedError: "invalid timestamp",
		},
		{
			name:          "invalid filename - missing timezone",
			filename:      "host_192.168.1.1_2025-01-01T12-00-00.json",
			expectedIP:    "",
			expectedTime:  time.Time{},
			expectedError: "does not match expected pattern",
		},
		{
			name:          "invalid filename - wrong timezone format",
			filename:      "host_192.168.1.1_2025-01-01T12-00-00+05:30.json",
			expectedIP:    "",
			expectedTime:  time.Time{},
			expectedError: "does not match expected pattern",
		},
		{
			name:          "invalid filename - wrong time separator",
			filename:      "host_192.168.1.1_2025-01-01T12:00:00Z.json",
			expectedIP:    "",
			expectedTime:  time.Time{},
			expectedError: "does not match expected pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			service := &SnapshotService{}

			// Test
			ip, timestamp, err := service.parseFileName(tt.filename)

			// Assertions
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Equal(t, "", ip)
				assert.Equal(t, time.Time{}, timestamp)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedIP, ip)
				assert.Equal(t, tt.expectedTime, timestamp)
			}
		})
	}
}

// Test file cleanup on repository error
func TestSnapshotService_CreateSnapshot_CleanupOnRepoError(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	mockRepo := &MockSnapshotRepo{}
	service := NewSnapshotService(mockRepo, tempDir)
	ctx := context.Background()

	filename := "host_192.168.1.1_2025-01-01T12-00-00Z.json"
	fileContent := `{"test": "data"}`

	// Setup mock to return error
	expectedFilePath := filepath.Join(tempDir, filename)
	expectedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	mockRepo.On("Insert", ctx, "192.168.1.1", expectedTime, expectedFilePath, filename).Return(fmt.Errorf("database error"))

	// Test
	file := createMultipartFile(fileContent)
	err := service.CreateSnapshot(ctx, file, filename)

	// Assertions
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to write file to DB")

	// Verify file was cleaned up
	assert.NoFileExists(t, expectedFilePath)

	mockRepo.AssertExpectations(t)
}

// Test file cleanup on file write error
func TestSnapshotService_CreateSnapshot_CleanupOnFileWriteError(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	mockRepo := &MockSnapshotRepo{}
	service := NewSnapshotService(mockRepo, tempDir)
	ctx := context.Background()

	filename := "host_192.168.1.1_2025-01-01T12-00-00Z.json"

	// Create a mock file that will cause an error during copy
	file := &mockMultipartFileWithError{}

	// Test
	err := service.CreateSnapshot(ctx, file, filename)

	// Assertions
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to write contents of file to file on OS")

	// Verify no file was created
	expectedFilePath := filepath.Join(tempDir, filename)
	assert.NoFileExists(t, expectedFilePath)
}

type mockMultipartFileWithError struct{}

func (m *mockMultipartFileWithError) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("read error")
}

func (m *mockMultipartFileWithError) ReadAt(p []byte, off int64) (n int, err error) {
	return 0, fmt.Errorf("ReadAt not implemented")
}

func (m *mockMultipartFileWithError) Seek(offset int64, whence int) (int64, error) {
	return 0, fmt.Errorf("Seek not implemented")
}

func (m *mockMultipartFileWithError) Close() error {
	return nil
}

// Test edge case with very long filename
func TestSnapshotService_CreateSnapshot_LongFilename(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	mockRepo := &MockSnapshotRepo{}
	service := NewSnapshotService(mockRepo, tempDir)
	ctx := context.Background()

	// Create a very long filename
	longIP := strings.Repeat("1", 100) + ".1.1.1"
	filename := fmt.Sprintf("host_%s_2025-01-01T12-00-00Z.json", longIP)
	fileContent := `{"test": "data"}`

	// Test
	file := createMultipartFile(fileContent)
	err := service.CreateSnapshot(ctx, file, filename)

	// Assertions
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not match expected pattern")
}

// Test concurrent access to same file
func TestSnapshotService_CreateSnapshot_ConcurrentAccess(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	mockRepo := &MockSnapshotRepo{}
	service := NewSnapshotService(mockRepo, tempDir)
	ctx := context.Background()

	filename := "host_192.168.1.1_2025-01-01T12-00-00Z.json"
	fileContent := `{"test": "data"}`

	// Setup mock expectations
	expectedFilePath := filepath.Join(tempDir, filename)
	expectedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	mockRepo.On("Insert", ctx, "192.168.1.1", expectedTime, expectedFilePath, filename).Return(nil).Once()

	// Test
	file := createMultipartFile(fileContent)
	err := service.CreateSnapshot(ctx, file, filename)

	// Assertions
	require.NoError(t, err)
	assert.FileExists(t, expectedFilePath)

	// Try to create the same file again - should fail
	file2 := createMultipartFile(fileContent)
	err2 := service.CreateSnapshot(ctx, file2, filename)

	// Assertions
	require.Error(t, err2)
	assert.Contains(t, err2.Error(), "Attempting to add duplicate file")

	mockRepo.AssertExpectations(t)
}

// Test with special characters in file content
func TestSnapshotService_CreateSnapshot_SpecialCharacters(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	mockRepo := &MockSnapshotRepo{}
	service := NewSnapshotService(mockRepo, tempDir)
	ctx := context.Background()

	filename := "host_192.168.1.1_2025-01-01T12-00-00Z.json"
	fileContent := `{"special": "chars: !@#$%^&*()_+-=[]{}|;':\",./<>?` + "`" + `", "unicode": "测试", "newlines": "line1\nline2\r\nline3"}`

	// Setup mock expectations
	expectedFilePath := filepath.Join(tempDir, filename)
	expectedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	mockRepo.On("Insert", ctx, "192.168.1.1", expectedTime, expectedFilePath, filename).Return(nil)

	// Test
	file := createMultipartFile(fileContent)
	err := service.CreateSnapshot(ctx, file, filename)

	// Assertions
	require.NoError(t, err)
	assert.FileExists(t, expectedFilePath)

	// Verify file content
	content, err := os.ReadFile(expectedFilePath)
	require.NoError(t, err)
	assert.Equal(t, fileContent, string(content))

	mockRepo.AssertExpectations(t)
}
