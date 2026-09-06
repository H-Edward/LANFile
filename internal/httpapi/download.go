package httpapi

import (
	"net/http"
	"path/filepath"

	"github.com/H-Edward/LANFile/internal/files"
)

type DownloadHandler struct {
	fileService *files.Service
	dataDir     string
}

func NewDownloadHandler(fileService *files.Service, dataDir string) *DownloadHandler {
	return &DownloadHandler{
		fileService: fileService,
		dataDir:     dataDir,
	}
}

func (h *DownloadHandler) GetbyName(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	files, err := h.fileService.GetByName(r.Context(), name)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	if len(files) == 0 {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	if len(files) > 1 {
		http.Error(w, "Multiple files found with the same name, use the file ID", http.StatusConflict)
		return
	}
	file := files[0]
	path := filepath.Join(h.dataDir, "files", file.StorageKey)

	http.ServeFile(w, r, path)
}

func (h *DownloadHandler) GetbyID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	file, err := h.fileService.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	path := filepath.Join(h.dataDir, "files", file.StorageKey)

	http.ServeFile(w, r, path)
}

