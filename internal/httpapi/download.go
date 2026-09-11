package httpapi

import (
	"mime"
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

func (h *DownloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		if r.PathValue("name") != "" {
			h.GetByName(w, r)
		} else if r.PathValue("id") != "" {
			h.GetByID(w, r)
		} else {
			http.Error(w, "Missing name or id in path", http.StatusBadRequest)
			return
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func (h *DownloadHandler) GetByName(w http.ResponseWriter, r *http.Request) {
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

	hasAuth, err := ReturnRequestHasAuth(r, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if !hasAuth {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	setDownloadFilename(w, file.OriginalName)

	http.ServeFile(w, r, path)
}

func (h *DownloadHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	file, err := h.fileService.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	path := filepath.Join(h.dataDir, "files", file.StorageKey)

	hasAuth, err := ReturnRequestHasAuth(r, *file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if !hasAuth {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	setDownloadFilename(w, file.OriginalName)
	http.ServeFile(w, r, path)
}

func setDownloadFilename(w http.ResponseWriter, name string) {
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
		"filename": name,
	}))
}
