package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/H-Edward/LANFile/internal/database"
	"github.com/H-Edward/LANFile/internal/files"
	"github.com/H-Edward/LANFile/internal/httpapi"
	"github.com/H-Edward/LANFile/internal/web"
	"github.com/go-chi/chi/v5"
)

type testServer struct {
	*httptest.Server
	DataDir string
	Client  *http.Client
}

func newIntegrationServer(t *testing.T) *testServer {
	t.Helper()

	_, testFile, _, _ := runtime.Caller(0)
	repoRoot := filepath.Dir(filepath.Dir(testFile))
	testDataRoot := filepath.Join(repoRoot, "tests", ".test-data")
	if err := os.MkdirAll(testDataRoot, 0o755); err != nil {
		t.Fatalf("create test data root: %v", err)
	}
	isolatedDir, err := os.MkdirTemp(testDataRoot, "integration-")
	if err != nil {
		t.Fatalf("create isolated test directory: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(isolatedDir) })
	if err := os.MkdirAll(filepath.Join(isolatedDir, "files"), 0o755); err != nil {
		t.Fatalf("create isolated files directory: %v", err)
	}

	db, err := database.Open(filepath.Join(isolatedDir, "data.db"))
	if err != nil {
		t.Fatalf("open isolated database: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, dbErr := db.DB()
		if dbErr == nil {
			_ = sqlDB.Close()
		}
	})

	service := files.NewService(db)
	router := chi.NewRouter()
	router.Mount("/", web.NewWebRouter(service))
	router.Mount("/api", httpapi.NewAPIRouter(service, isolatedDir))

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	return &testServer{
		Server:  server,
		DataDir: isolatedDir,
		Client:  server.Client(),
	}
}

// Helper to handle requests easily inside subtests
func (ts *testServer) doRequest(t *testing.T, method, path string, body string, user, password string) (*http.Response, string) {
	t.Helper()
	var bodyReader io.Reader
	if body != "" {
		bodyReader = bytes.NewBufferString(body)
	}

	req, err := http.NewRequest(method, ts.URL+path, bodyReader)
	if err != nil {
		t.Fatalf("failed to create request [%s %s]: %v", method, path, err)
	}

	if user != "" || password != "" {
		req.SetBasicAuth(user, password)
	}

	resp, err := ts.Client.Do(req)
	if err != nil {
		t.Fatalf("failed to execute request [%s %s]: %v", method, path, err)
	}

	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	return resp, string(respBody)
}

func TestUploadAPI(t *testing.T) {
	ts := newIntegrationServer(t)

	t.Run("Upload New File By Name", func(t *testing.T) {
		resp, body := ts.doRequest(t, http.MethodPut, "/api/u/name/hello.txt", "Hello World", "", "")
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected status 201 Created, got %d. Body: %s", resp.StatusCode, body)
		}

		var file map[string]any
		if err := json.Unmarshal([]byte(body), &file); err != nil {
			t.Fatalf("failed to parse JSON response: %v", err)
		}

		fileID, ok := file["ID"].(string)
		if !ok || fileID == "" {
			t.Fatalf("response missing ID field: %s", body)
		}

		storageKey, ok := file["StorageKey"].(string)
		if !ok || storageKey == "" {
			t.Fatalf("response missing StorageKey field: %s", body)
		}

		// Verify file actually written to disk
		if _, err := os.Stat(filepath.Join(ts.DataDir, "files", storageKey)); err != nil {
			t.Fatalf("file not found on disk storage: %v", err)
		}
	})

	t.Run("Allow Duplicate Upload With Overwrite Flag", func(t *testing.T) {
		ts.doRequest(t, http.MethodPut, "/api/u/name/overwrite.txt", "Initial Content", "", "")

		resp, body := ts.doRequest(t, http.MethodPut, "/api/u/name/overwrite.txt?overwrite=true", "New Content", "", "")
		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200/201 on overwrite, got %d. Body: %s", resp.StatusCode, body)
		}
	})

	t.Run("Upload With Encryption Metadata Query Param", func(t *testing.T) {
		resp, body := ts.doRequest(t, http.MethodPut, "/api/u/name/secret.enc?encrypted=password", "Encrypted Data", "", "")
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d", resp.StatusCode)
		}

		var file map[string]any
		_ = json.Unmarshal([]byte(body), &file)
		if file["Encrypted"] != "password" {
			t.Errorf("expected encryption param metadata to be saved/returned, got: %v", body)
		}
	})
}

func TestDownloadAndAuthAPI(t *testing.T) {
	ts := newIntegrationServer(t)

	// Seed unauthenticated file
	ts.doRequest(t, http.MethodPut, "/api/u/name/public.txt", "Public Info", "", "")

	// Seed password-protected file via HTTP Basic Auth as described in README
	_, uploadBody := ts.doRequest(t, http.MethodPut, "/api/u/name/protected.txt", "Confidential Data", "", "secretpass")
	var protectedFile map[string]any
	_ = json.Unmarshal([]byte(uploadBody), &protectedFile)
	protectedID := protectedFile["ID"].(string)

	t.Run("Download Public File By Name", func(t *testing.T) {
		resp, body := ts.doRequest(t, http.MethodGet, "/api/d/name/public.txt", "", "", "")
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
		}
		if body != "Public Info" {
			t.Errorf("expected body 'Public Info', got %q", body)
		}
	})

	t.Run("Download Protected File Without Auth Fails", func(t *testing.T) {
		resp, _ := ts.doRequest(t, http.MethodGet, "/api/d/id/"+protectedID, "", "", "")
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized, got %d", resp.StatusCode)
		}
	})

	t.Run("Download Protected File With Wrong Password Fails", func(t *testing.T) {
		resp, _ := ts.doRequest(t, http.MethodGet, "/api/d/id/"+protectedID, "", "", "wrongpass")
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized, got %d", resp.StatusCode)
		}
	})

	t.Run("Download Protected File With Correct Password Succeeds", func(t *testing.T) {
		resp, body := ts.doRequest(t, http.MethodGet, "/api/d/id/"+protectedID, "", "", "secretpass")
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
		}
		if body != "Confidential Data" {
			t.Errorf("expected body 'Confidential Data', got %q", body)
		}
	})

	t.Run("Download Nonexistent File Returns 404", func(t *testing.T) {
		resp, _ := ts.doRequest(t, http.MethodGet, "/api/d/name/doesnotexist.txt", "", "", "")
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected 404 Not Found, got %d", resp.StatusCode)
		}
	})
}

func TestSearchAPI(t *testing.T) {
	ts := newIntegrationServer(t)

	ts.doRequest(t, http.MethodPut, "/api/u/name/alpha_report.txt", "data 1", "", "")
	ts.doRequest(t, http.MethodPut, "/api/u/name/beta_report.txt", "data 2", "", "")
	ts.doRequest(t, http.MethodPut, "/api/u/name/gamma_notes.txt", "data 3", "", "")

	t.Run("Substring Search", func(t *testing.T) {
		resp, body := ts.doRequest(t, http.MethodGet, "/api/s/name/report", "", "", "")
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
		}

		var results []map[string]any
		if err := json.Unmarshal([]byte(body), &results); err != nil {
			t.Fatalf("failed to parse search response: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("expected 2 search results for 'report', got %d", len(results))
		}
	})

	t.Run("Exact Match Search - Positive", func(t *testing.T) {
		resp, body := ts.doRequest(t, http.MethodGet, "/api/s/name/gamma_notes.txt?exact=true", "", "", "")
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
		}

		var results []map[string]any
		_ = json.Unmarshal([]byte(body), &results)
		if len(results) != 1 {
			t.Errorf("expected 1 exact match, got %d", len(results))
		}
	})

	t.Run("Exact Match Search - Negative", func(t *testing.T) {
		resp, body := ts.doRequest(t, http.MethodGet, "/api/s/name/gamma_notes?exact=true", "", "", "")
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404 Not Found, got %d", resp.StatusCode)
		}

		var results []map[string]any
		_ = json.Unmarshal([]byte(body), &results)
		if len(results) != 0 {
			t.Errorf("expected 0 exact matches for partial string, got %d", len(results))
		}
	})

	t.Run("Get All Metadata", func(t *testing.T) {
		resp, body := ts.doRequest(t, http.MethodGet, "/api/s/getall", "", "", "")
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
		}

		var results []map[string]any
		_ = json.Unmarshal([]byte(body), &results)
		if len(results) < 3 {
			t.Errorf("expected at least 3 files in getall, got %d", len(results))
		}
	})
}

func TestWebUI(t *testing.T) {
	ts := newIntegrationServer(t)

	t.Run("Serve Home Page", func(t *testing.T) {
		resp, _ := ts.doRequest(t, http.MethodGet, "/", "", "", "")
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200 OK for Web UI, got %d", resp.StatusCode)
		}
	})
}
