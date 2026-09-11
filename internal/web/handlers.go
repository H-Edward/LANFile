package web

import "net/http"

type IndexData struct {
	Title      string
	Page       string
	Files      []FileData
	TargetID   string
	TargetName string
	NeedsAuth  bool
}

type FileData struct {
	ID        string
	Name      string
	Size      int64
	Encrypted string
	NeedsAuth bool
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	files, err := h.files.GetAll(r.Context())
	if err != nil {
		http.Error(w, "Failed to load files", http.StatusInternalServerError)
		return
	}

	fileData := make([]FileData, 0, len(files))
	for _, file := range files {
		needsAuth := file.NeedsAuth || file.AuthorisationHash != "none"
		fileData = append(fileData, FileData{ID: file.ID, Name: file.OriginalName, Size: file.Size, Encrypted: file.Encrypted, NeedsAuth: needsAuth})
	}

	data := IndexData{
		Title: "LANFile",
		Page:  "index",
		Files: fileData,
	}

	if err := h.templates.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	targetID := r.URL.Query().Get("id")
	targetName := r.URL.Query().Get("name")
	needsAuth := false
	if targetID != "" {
		file, err := h.files.GetByID(r.Context(), targetID)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		needsAuth = file.NeedsAuth || file.AuthorisationHash != "none"
	}
	data := IndexData{Title: "Upload | LANFile", Page: "upload", TargetID: targetID, TargetName: targetName, NeedsAuth: needsAuth}
	if err := h.templates.ExecuteTemplate(w, "upload-page", data); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

func (h *Handler) Files(w http.ResponseWriter, r *http.Request) {
	h.Index(w, r)
}

func formatSize(size int64) string {
	if size < 1024 {
		return formatNumber(size) + " B"
	}
	units := []string{"KB", "MB", "GB", "TB"}
	value := float64(size)
	for _, unit := range units {
		value /= 1024
		if value < 1024 || unit == units[len(units)-1] {
			return formatDecimal(value) + " " + unit
		}
	}
	return formatNumber(size) + " B"
}
