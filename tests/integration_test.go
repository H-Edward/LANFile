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
	"time"

	"github.com/H-Edward/LANFile/internal/database"
	"github.com/H-Edward/LANFile/internal/files"
	"github.com/H-Edward/LANFile/internal/httpapi"
	"github.com/H-Edward/LANFile/internal/web"
	"github.com/go-chi/chi/v5"
)

func newIntegrationServer(t *testing.T) (*httptest.Server, string) {
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
	return server, isolatedDir
}

func TestFileLifecycleOverHTTP(t *testing.T) {
	server, dataDir := newIntegrationServer(t)
	client := server.Client()

	upload := func(path, content string) map[string]any {
		t.Helper()
		request, err := http.NewRequest(http.MethodPut, server.URL+path, bytes.NewBufferString(content))
		if err != nil {
			t.Fatalf("create upload request: %v", err)
		}
		response, err := client.Do(request)
		if err != nil {
			t.Fatalf("upload request: %v", err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(response.Body)
			t.Fatalf("upload status = %d, body = %q", response.StatusCode, body)
		}
		var file map[string]any
		if err := json.NewDecoder(response.Body).Decode(&file); err != nil {
			t.Fatalf("decode upload response: %v", err)
		}
		return file
	}

	first := upload("/api/u/name/report.txt", "first contents")
	fileID, ok := first["ID"].(string)
	if !ok || fileID == "" {
		t.Fatalf("upload response did not contain an ID: %#v", first)
	}
	storageKey, ok := first["StorageKey"].(string)
	if !ok || storageKey == "" {
		t.Fatalf("upload response did not contain a storage key: %#v", first)
	}
	if _, err := os.Stat(filepath.Join(dataDir, "files", storageKey)); err != nil {
		t.Fatalf("uploaded file was not stored in isolated data directory: %v", err)
	}

	response, err := client.Get(server.URL + "/api/s/name/report.txt?exact=true")
	if err != nil {
		t.Fatalf("search request: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("search status = %d", response.StatusCode)
	}
	var searchResults []map[string]any
	if err := json.NewDecoder(response.Body).Decode(&searchResults); err != nil {
		t.Fatalf("decode search response: %v", err)
	}
	if len(searchResults) != 1 || searchResults[0]["ID"] != fileID {
		t.Fatalf("unexpected search results: %#v", searchResults)
	}
	response.Body.Close()

	response, err = client.Get(server.URL + "/api/d/id/" + fileID)
	if err != nil {
		t.Fatalf("download request: %v", err)
	}
	body, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		t.Fatalf("read download response: %v", err)
	}
	if response.StatusCode != http.StatusOK || string(body) != "first contents" {
		t.Fatalf("download status/body = %d/%q", response.StatusCode, body)
	}

	request, err := http.NewRequest(http.MethodPut, server.URL+"/api/u/id/"+fileID+"?encrypted=password", bytes.NewBufferString("updated contents"))
	if err != nil {
		t.Fatalf("create overwrite request: %v", err)
	}
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("overwrite request: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("overwrite status = %d", response.StatusCode)
	}

	request, err = http.NewRequest(http.MethodHead, server.URL+"/api/d/id/"+fileID, nil)
	if err != nil {
		t.Fatalf("create HEAD request: %v", err)
	}
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("HEAD request: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK || response.Header.Get("Content-Disposition") == "" {
		t.Fatalf("HEAD status/headers = %d/%q", response.StatusCode, response.Header.Get("Content-Disposition"))
	}

	response, err = client.Get(server.URL + "/")
	if err != nil {
		t.Fatalf("web index request: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("web index status = %d", response.StatusCode)
	}

	request, err = http.NewRequest(http.MethodDelete, server.URL+"/api/delete/id/"+fileID, nil)
	if err != nil {
		t.Fatalf("create delete request: %v", err)
	}
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("delete request: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status = %d", response.StatusCode)
	}
}
