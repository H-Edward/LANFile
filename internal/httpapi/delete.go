package httpapi

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/H-Edward/LANFile/internal/files"
)

type DeleteHandler struct {
	fileService *files.Service
	dataDir     string
}

func NewDeleteHandler(fileService *files.Service, dataDir string) *DeleteHandler {
	return &DeleteHandler{
		fileService: fileService,
		dataDir:     dataDir,
	}
}

func (h *DeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodDelete:
		if r.PathValue("id") != "" {
			h.DeleteByID(w, r)
		} else {
			http.Error(w, "Missing id in path", http.StatusBadRequest)
			return
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func (h *DeleteHandler) DeleteByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	file, err := h.fileService.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	storage_key := file.StorageKey

	if file.NeedsAuth {
		hasAuth, err := ReturnRequestHasAuth(r, *file)
		if !hasAuth || err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	err = h.fileService.DeleteByID(r.Context(), id)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	// Delete the file from disk
	path := filepath.Join(h.dataDir, "files", storage_key)
	err = os.Remove(path)
	if err != nil {
		http.Error(w, "Failed to delete file from disk", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
