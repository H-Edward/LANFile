package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/H-Edward/LANFile/internal/database"
	"github.com/H-Edward/LANFile/internal/files"
)

type SearchHandler struct {
	fileService *files.Service
	dataDir     string
}

func NewSearchHandler(fileService *files.Service, dataDir string) *SearchHandler {
	return &SearchHandler{
		fileService: fileService,
		dataDir:     dataDir,
	}
}

func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if r.URL.Path == "/api/s/getall" {
			h.GetAllFiles(w, r)
		} else if r.PathValue("name") != "" {
			h.SearchByName(w, r)
		} else if r.PathValue("id") != "" {
			h.SearchByID(w, r)
		} else {
			http.Error(w, "Missing name or id in path", http.StatusBadRequest)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *SearchHandler) SearchByName(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	exact := r.URL.Query().Get("exact")
	var files []database.File
	var err error
	if exact == "true" {
		files, err = h.fileService.GetByName(r.Context(), name)
		if err != nil {
			http.Error(w, "Error searching for files", http.StatusNotFound)
			return
		}
		if len(files) == 0 {
			http.Error(w, "No files found", http.StatusNotFound)
			return
		} else if len(files) > 1 {
			http.Error(w, "Multiple files found with the same name, use the file ID", http.StatusConflict)
			return
		}

	} else {
		files, err = h.fileService.SearchByName(r.Context(), name)
		if err != nil {
			http.Error(w, "Error searching for files", http.StatusNotFound)
			return
		}
		if len(files) == 0 {
			http.Error(w, "No files found", http.StatusNotFound)
			return
		}
	}

	// Return the list of files as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func (h *SearchHandler) SearchByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	file, err := h.fileService.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Error searching for files", http.StatusNotFound)
		return
	}
	if file == nil {
		http.Error(w, "No files found", http.StatusNotFound)
		return
	}

	// Return the list of files as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(file)
}

func (h *SearchHandler) GetAllFiles(w http.ResponseWriter, r *http.Request) {
	files, err := h.fileService.GetAll(r.Context())
	if err != nil {
		http.Error(w, "Error retrieving files", http.StatusNotFound)
		return
	}
	// Return the list of files as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}
