package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/endingwithali/2025censys/internal/repo"
	"github.com/endingwithali/2025censys/internal/service"
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

// Helper function to create a server for testing
func createTestServer(mockSnapshotRepo *MockSnapshotRepo, maxFileSize int) *Server {
	snapshotService := service.NewSnapshotService(mockSnapshotRepo, "/tmp")
	diffService := service.NewDifferencesServicet()

	return &Server{
		snapshotService:   snapshotService,
		differenceService: diffService,
		MaxFileSize:       maxFileSize,
	}
}

func TestServer_Get(t *testing.T) {
	// Setup
	mockSnapshotRepo := &MockSnapshotRepo{}
	server := createTestServer(mockSnapshotRepo, 1024*1024)

	// Test
	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	server.Get(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "All Connected!", w.Body.String())
}

func TestServer_ListAllHosts(t *testing.T) {
	tests := []struct {
		name           string
		hosts          []string
		expectedStatus int
		expectedBody   []string
		repoError      error
	}{
		{
			name:           "successful request",
			hosts:          []string{"192.168.1.1", "10.0.0.1"},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"192.168.1.1", "10.0.0.1"},
			repoError:      nil,
		},
		{
			name:           "empty hosts list",
			hosts:          []string{},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{},
			repoError:      nil,
		},
		{
			name:           "repository error",
			hosts:          nil,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   nil,
			repoError:      fmt.Errorf("database connection failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockSnapshotRepo := &MockSnapshotRepo{}
			server := createTestServer(mockSnapshotRepo, 1024*1024)

			mockSnapshotRepo.On("GetAllHosts", mock.Anything).Return(tt.hosts, tt.repoError)

			// Test
			req := httptest.NewRequest("GET", "/api/host/all", nil)
			w := httptest.NewRecorder()

			server.ListAllHosts(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != nil {
				var response []string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}
			mockSnapshotRepo.AssertExpectations(t)
		})
	}
}

func TestServer_GetAllSnapshotsForHost(t *testing.T) {
	tests := []struct {
		name           string
		ip             string
		snapshots      []string
		expectedStatus int
		repoError      error
	}{
		{
			name:           "successful request",
			ip:             "192.168.1.1",
			snapshots:      []string{"2025-01-01T12:00:00Z", "2025-01-02T12:00:00Z"},
			expectedStatus: http.StatusOK,
			repoError:      nil,
		},
		{
			name:           "missing ip parameter",
			ip:             "",
			snapshots:      nil,
			expectedStatus: http.StatusNotAcceptable,
			repoError:      nil,
		},
		{
			name:           "repository error",
			ip:             "192.168.1.1",
			snapshots:      nil,
			expectedStatus: http.StatusInternalServerError,
			repoError:      fmt.Errorf("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockSnapshotRepo := &MockSnapshotRepo{}
			server := createTestServer(mockSnapshotRepo, 1024*1024)

			if tt.ip != "" {
				mockSnapshotRepo.On("ListAllHostSnapshots", mock.Anything, tt.ip).Return(tt.snapshots, tt.repoError)
			}

			// Test
			url := "/api/host"
			if tt.ip != "" {
				url += "?ip=" + tt.ip
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			server.GetAllSnapshotsForHost(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response []string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.snapshots, response)
			}
			mockSnapshotRepo.AssertExpectations(t)
		})
	}
}

func TestServer_GetSnapshotForHost(t *testing.T) {
	tests := []struct {
		name           string
		ip             string
		timestamp      string
		filePath       string
		fileContent    string
		expectedStatus int
		repoError      error
		fileExists     bool
	}{
		{
			name:           "successful request",
			ip:             "192.168.1.1",
			timestamp:      "2025-01-01T12:00:00Z",
			filePath:       "/tmp/test.json",
			fileContent:    `{"test": "data"}`,
			expectedStatus: http.StatusOK,
			repoError:      nil,
			fileExists:     true,
		},
		{
			name:           "missing ip parameter",
			ip:             "",
			timestamp:      "2025-01-01T12:00:00Z",
			expectedStatus: http.StatusNotAcceptable,
			repoError:      nil,
			fileExists:     false,
		},
		{
			name:           "missing timestamp parameter",
			ip:             "192.168.1.1",
			timestamp:      "",
			expectedStatus: http.StatusNotAcceptable,
			repoError:      nil,
			fileExists:     false,
		},
		{
			name:           "repository error",
			ip:             "192.168.1.1",
			timestamp:      "2025-01-01T12:00:00Z",
			expectedStatus: http.StatusBadRequest,
			repoError:      fmt.Errorf("snapshot not found"),
			fileExists:     false,
		},
		{
			name:           "file not found",
			ip:             "192.168.1.1",
			timestamp:      "2025-01-01T12:00:00Z",
			filePath:       "/tmp/nonexistent.json",
			expectedStatus: http.StatusNotFound,
			repoError:      nil,
			fileExists:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tempDir := t.TempDir()
			mockSnapshotRepo := &MockSnapshotRepo{}
			snapshotService := service.NewSnapshotService(mockSnapshotRepo, tempDir)
			diffService := service.NewDifferencesServicet()
			server := &Server{
				snapshotService:   snapshotService,
				differenceService: diffService,
				MaxFileSize:       1024 * 1024,
			}

			if tt.ip != "" && tt.timestamp != "" && tt.fileExists {
				// Create test file
				testFilePath := filepath.Join(tempDir, "test.json")
				err := os.WriteFile(testFilePath, []byte(tt.fileContent), 0644)
				require.NoError(t, err)

				parsedTime, err := time.Parse(time.RFC3339, tt.timestamp)
				require.NoError(t, err)

				mockSnapshot := repo.Snapshot{
					UUID:      uuid.New(),
					Host_IP:   tt.ip,
					Timestamp: parsedTime,
					File_PWD:  testFilePath,
					File_Name: "test.json",
				}

				mockSnapshotRepo.On("GetSnapshotByTimeStamp", mock.Anything, tt.ip, parsedTime).Return(mockSnapshot, tt.repoError)
			} else if tt.ip != "" && tt.timestamp != "" {
				parsedTime, err := time.Parse(time.RFC3339, tt.timestamp)
				require.NoError(t, err)
				mockSnapshotRepo.On("GetSnapshotByTimeStamp", mock.Anything, tt.ip, parsedTime).Return(repo.Snapshot{}, tt.repoError)
			}

			// Test
			url := "/api/snapshot"
			if tt.ip != "" && tt.timestamp != "" {
				url += "?ip=" + tt.ip + "&at=" + tt.timestamp
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			server.GetSnapshotForHost(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, tt.fileContent, w.Body.String())
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			}
			mockSnapshotRepo.AssertExpectations(t)
		})
	}
}

func TestServer_CreateSnapshot(t *testing.T) {
	tests := []struct {
		name           string
		filename       string
		fileContent    string
		expectedStatus int
		repoError      error
		fileExists     bool
		maxFileSize    int
	}{
		{
			name:           "successful upload",
			filename:       "host_192.168.1.1_2025-01-01T12-00-00Z.json",
			fileContent:    `{"test": "data"}`,
			expectedStatus: http.StatusOK,
			repoError:      nil,
			fileExists:     false,
			maxFileSize:    1024 * 1024,
		},
		{
			name:           "invalid filename format",
			filename:       "invalid.json",
			fileContent:    `{"test": "data"}`,
			expectedStatus: http.StatusBadRequest,
			repoError:      nil,
			fileExists:     false,
			maxFileSize:    1024 * 1024,
		},
		{
			name:           "duplicate file",
			filename:       "host_192.168.1.1_2025-01-01T12-00-00Z.json",
			fileContent:    `{"test": "data"}`,
			expectedStatus: http.StatusConflict,
			repoError:      nil,
			fileExists:     true,
			maxFileSize:    1024 * 1024,
		},
		{
			name:           "repository error",
			filename:       "host_192.168.1.1_2025-01-01T12-00-00Z.json",
			fileContent:    `{"test": "data"}`,
			expectedStatus: http.StatusConflict,
			repoError:      fmt.Errorf("database error"),
			fileExists:     false,
			maxFileSize:    1024 * 1024,
		},
		{
			name:           "file too large",
			filename:       "host_192.168.1.1_2025-01-01T12-00-00Z.json",
			fileContent:    strings.Repeat("a", 1024*1024+1),
			expectedStatus: http.StatusBadRequest,
			repoError:      nil,
			fileExists:     false,
			maxFileSize:    1024 * 1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tempDir := t.TempDir()
			mockSnapshotRepo := &MockSnapshotRepo{}
			diffService := service.NewDifferencesServicet()
			snapshotService := service.NewSnapshotService(mockSnapshotRepo, tempDir)
			server := &Server{
				snapshotService:   snapshotService,
				differenceService: diffService,
				MaxFileSize:       tt.maxFileSize,
			}

			// Create existing file if needed
			if tt.fileExists {
				existingFilePath := filepath.Join(tempDir, tt.filename)
				err := os.WriteFile(existingFilePath, []byte("existing content"), 0644)
				require.NoError(t, err)
			}

			// Setup mock expectations for cases that will call Insert
			if tt.expectedStatus == http.StatusOK || (tt.expectedStatus == http.StatusConflict && tt.repoError != nil) {
				parsedTime, err := time.Parse("2006-01-02T15-04-05Z", "2025-01-01T12-00-00Z")
				require.NoError(t, err)
				expectedFilePath := filepath.Join(tempDir, tt.filename)
				mockSnapshotRepo.On("Insert", mock.Anything, "192.168.1.1", parsedTime, expectedFilePath, tt.filename).Return(tt.repoError)
			}

			// Test
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			part, err := writer.CreateFormFile("file", tt.filename)
			require.NoError(t, err)
			_, err = part.Write([]byte(tt.fileContent))
			require.NoError(t, err)
			err = writer.Close()
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/snapshot", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			w := httptest.NewRecorder()

			server.CreateSnapshot(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)
			mockSnapshotRepo.AssertExpectations(t)
		})
	}
}

func TestServer_GetSnapshotDiffs(t *testing.T) {
	tests := []struct {
		name           string
		ip             string
		t1             string
		t2             string
		file1Path      string
		file2Path      string
		file1Content   string
		file2Content   string
		diffStatus     string
		diffContent    string
		expectedStatus int
		repoError      error
		diffError      error
	}{
		{
			name:         "successful diff",
			ip:           "192.168.1.1",
			t1:           "2025-01-01T12:00:00Z",
			t2:           "2025-01-02T12:00:00Z",
			file1Path:    "/tmp/file1.json",
			file2Path:    "/tmp/file2.json",
			file1Content: `{"key": "value1"}`,
			file2Content: `{"key": "value2"}`,
			diffStatus:   "NoMatch", // This is what the real service returns
			diffContent: `{
    "key": [0;33m"value1" => "value2"[0m
}`, // This is what the real service returns (with ANSI color codes)
			expectedStatus: http.StatusOK,
			repoError:      nil,
			diffError:      nil,
		},
		{
			name:           "missing ip parameter",
			ip:             "",
			t1:             "2025-01-01T12:00:00Z",
			t2:             "2025-01-02T12:00:00Z",
			expectedStatus: http.StatusNotAcceptable,
			repoError:      nil,
			diffError:      nil,
		},
		{
			name:           "missing t1 parameter",
			ip:             "192.168.1.1",
			t1:             "",
			t2:             "2025-01-02T12:00:00Z",
			expectedStatus: http.StatusNotAcceptable,
			repoError:      nil,
			diffError:      nil,
		},
		{
			name:           "missing t2 parameter",
			ip:             "192.168.1.1",
			t1:             "2025-01-01T12:00:00Z",
			t2:             "",
			expectedStatus: http.StatusNotAcceptable,
			repoError:      nil,
			diffError:      nil,
		},
		{
			name:           "snapshot not found for t1",
			ip:             "192.168.1.1",
			t1:             "2025-01-01T12:00:00Z",
			t2:             "2025-01-02T12:00:00Z",
			expectedStatus: http.StatusNoContent,
			repoError:      fmt.Errorf("snapshot not found"),
			diffError:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tempDir := t.TempDir()
			mockSnapshotRepo := &MockSnapshotRepo{}
			snapshotService := service.NewSnapshotService(mockSnapshotRepo, tempDir)
			diffService := service.NewDifferencesServicet()
			server := &Server{
				snapshotService:   snapshotService,
				differenceService: diffService,
				MaxFileSize:       1024 * 1024,
			}

			// Create test files if needed
			if tt.file1Path != "" {
				testFile1Path := filepath.Join(tempDir, "file1.json")
				err := os.WriteFile(testFile1Path, []byte(tt.file1Content), 0644)
				require.NoError(t, err)
			}
			if tt.file2Path != "" {
				testFile2Path := filepath.Join(tempDir, "file2.json")
				err := os.WriteFile(testFile2Path, []byte(tt.file2Content), 0644)
				require.NoError(t, err)
			}

			// Setup mock expectations
			if tt.ip != "" && tt.t1 != "" && tt.t2 != "" {
				if tt.repoError == nil {
					parsedTime1, err := time.Parse(time.RFC3339, tt.t1)
					require.NoError(t, err)
					parsedTime2, err := time.Parse(time.RFC3339, tt.t2)
					require.NoError(t, err)

					mockSnapshot1 := repo.Snapshot{
						UUID:      uuid.New(),
						Host_IP:   tt.ip,
						Timestamp: parsedTime1,
						File_PWD:  filepath.Join(tempDir, "file1.json"),
						File_Name: "file1.json",
					}
					mockSnapshot2 := repo.Snapshot{
						UUID:      uuid.New(),
						Host_IP:   tt.ip,
						Timestamp: parsedTime2,
						File_PWD:  filepath.Join(tempDir, "file2.json"),
						File_Name: "file2.json",
					}

					mockSnapshotRepo.On("GetSnapshotByTimeStamp", mock.Anything, tt.ip, parsedTime1).Return(mockSnapshot1, nil)
					mockSnapshotRepo.On("GetSnapshotByTimeStamp", mock.Anything, tt.ip, parsedTime2).Return(mockSnapshot2, nil)

					// Note: Using real diff service, so we don't mock expectations
				} else {
					parsedTime1, err := time.Parse(time.RFC3339, tt.t1)
					require.NoError(t, err)
					mockSnapshotRepo.On("GetSnapshotByTimeStamp", mock.Anything, tt.ip, parsedTime1).Return(repo.Snapshot{}, tt.repoError)
				}
			}

			// Test
			url := "/api/snapshot/diff"
			if tt.ip != "" && tt.t1 != "" && tt.t2 != "" {
				url += "?ip=" + tt.ip + "&t1=" + tt.t1 + "&t2=" + tt.t2
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			server.GetSnapshotDiffs(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response diffResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.diffStatus, response.DiffStatus)
				// Don't check exact diff content as it includes ANSI color codes
				assert.NotEmpty(t, response.Differences)
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			}
			mockSnapshotRepo.AssertExpectations(t)
		})
	}
}

func TestServer_NotFound(t *testing.T) {
	// Setup
	mockSnapshotRepo := &MockSnapshotRepo{}
	server := createTestServer(mockSnapshotRepo, 1024*1024)

	// Test
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()

	server.NotFound(w, req)

	// Assertions
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "Not Found", w.Body.String())
}

func TestServer_MethodNotAllowed(t *testing.T) {
	// Setup
	mockSnapshotRepo := &MockSnapshotRepo{}
	server := createTestServer(mockSnapshotRepo, 1024*1024)

	// Test
	req := httptest.NewRequest("POST", "/api/health", nil)
	w := httptest.NewRecorder()

	server.MethodNotAllowed(w, req)

	// Assertions
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Equal(t, "Method Not Allowed", w.Body.String())
}

func TestServer_validateFileNameFormat(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		valid    bool
	}{
		{
			name:     "valid filename",
			filename: "host_192.168.1.1_2025-01-01T12-00-00Z.json",
			valid:    true,
		},
		{
			name:     "valid filename with fractional seconds",
			filename: "host_192.168.1.1_2025-01-01T12-00-00.123Z.json",
			valid:    true,
		},
		{
			name:     "valid filename with timezone offset",
			filename: "host_192.168.1.1_2025-01-01T12-00-00+05-30.json",
			valid:    true,
		},
		{
			name:     "invalid filename - wrong format",
			filename: "invalid.json",
			valid:    false,
		},
		{
			name:     "invalid filename - wrong prefix",
			filename: "file_192.168.1.1_2025-01-01T12-00-00Z.json",
			valid:    false,
		},
		{
			name:     "invalid filename - wrong extension",
			filename: "host_192.168.1.1_2025-01-01T12-00-00Z.txt",
			valid:    false,
		},
		{
			name:     "invalid filename - invalid IP",
			filename: "host_999.999.999.999_2025-01-01T12-00-00Z.json",
			valid:    true, // The regex only checks format, not IP validity
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockSnapshotRepo := &MockSnapshotRepo{}
			server := createTestServer(mockSnapshotRepo, 1024*1024)

			// Test
			result := server.validateFileNameFormat(tt.filename)

			// Assertions
			assert.Equal(t, tt.valid, result)
		})
	}
}

// Integration test to verify the router setup
func TestRouterSetup(t *testing.T) {
	// Setup
	mockSnapshotRepo := &MockSnapshotRepo{}
	snapshotService := service.NewSnapshotService(mockSnapshotRepo, "/tmp")
	diffService := service.NewDifferencesServicet()
	router := New(snapshotService, diffService, 1024*1024)

	// Setup mock expectations for the host/all endpoint
	mockSnapshotRepo.On("GetAllHosts", mock.Anything).Return([]string{}, fmt.Errorf("database error"))

	tests := []struct {
		method string
		path   string
		status int
	}{
		{"GET", "/api/health", http.StatusOK},
		{"GET", "/api/host/all", http.StatusInternalServerError}, // Will fail due to mock returning error
		{"GET", "/api/host", http.StatusNotAcceptable},
		{"GET", "/api/snapshot", http.StatusNotAcceptable},
		{"POST", "/api/snapshot", http.StatusBadRequest}, // Missing file
		{"GET", "/api/snapshot/diff", http.StatusNotAcceptable},
		{"GET", "/nonexistent", http.StatusNotFound},
		{"POST", "/api/health", http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s %s", tt.method, tt.path), func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.status, w.Code)
		})
	}

	mockSnapshotRepo.AssertExpectations(t)
}
