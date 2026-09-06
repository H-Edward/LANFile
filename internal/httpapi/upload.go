package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"uuid"

	"github.com/H-Edward/LANFile/internal/database"
	"github.com/H-Edward/LANFile/internal/files"
)

type UploadHandler struct {
	fileService *files.Service
	dataDir     string
}

func NewUploadHandler(fileService *files.Service, dataDir string) *UploadHandler {
	return &UploadHandler{
		fileService: fileService,
		dataDir:     dataDir,
	}
}

func returnEncryptionType(r *http.Request) (string, error) {
	encrypted := r.URL.Query().Get("encrypted")
	if encrypted != "gpg" && encrypted != "pass" && encrypted != "" {
		return "", fmt.Errorf("Invalid encryption method")
	}
	if encrypted == "" {
		return "false", nil
	}
	return encrypted, nil
}

func returnOverwriteValue(r *http.Request) (string, error) {
	overwrite := r.URL.Query().Get("overwrite")
	if overwrite != "true" && overwrite != "false" && overwrite != "" {
		return "", fmt.Errorf("Invalid overwrite value")
	}
	if overwrite == "" {
		return "false", nil
	}
	return overwrite, nil
}

// createNewFileRecord handles the creation of a new file record in the database and saves the uploaded file to disk. It also handles overwriting existing files if specified.
func (h *UploadHandler) createNewFileRecord(w http.ResponseWriter, r *http.Request, overwrite string, name string, encrypted string, matchingFiles []database.File) {

	storageKey := uuid.New().String()
	path := filepath.Join(h.dataDir, "files", storageKey)

	dst, err := os.Create(path)
	if err != nil {
		fmt.Println("Failed to create file:", err)
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	size, err := io.Copy(dst, r.Body)
	if err != nil {
		_ = os.Remove(path)
		http.Error(w, "Upload failed", http.StatusInternalServerError)
		return
	}

	input := files.CreateFileInput{
		OriginalName: name,
		ContentType:  r.Header.Get("Content-Type"),
		Size:         size,
		StorageKey:   storageKey,
		Encrypted:    encrypted,
	}
	var file *database.File
	// If overwrite is true and there is exactly one matching file, overwrite it.
	if overwrite == "true" && len(matchingFiles) == 1 {
		oldFile := matchingFiles[0]

		file, err = h.fileService.OverwriteByID(
			r.Context(),
			oldFile.ID,
			input,
		)

		if err == nil {
			oldPath := filepath.Join(
				h.dataDir,
				"files",
				oldFile.StorageKey,
			)

			if removeErr := os.Remove(oldPath); removeErr != nil && !os.IsNotExist(removeErr) {
				fmt.Println("Failed to remove old file:", removeErr)
			}
		}
	} else { // If overwrite is false, a new file is simply being created
		file, err = h.fileService.Create(
			r.Context(),
			input,
		)
	}

	if err != nil {
		_ = os.Remove(path)
		http.Error(w, "Failed to save metadata", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(file); err != nil {
		fmt.Println("Failed to write response:", err)
	}
}

func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		if r.PathValue("name") != "" {
			h.handleUploadByName(w, r)
		} else if r.PathValue("id") != "" {
			h.handleUploadById(w, r)
		} else {
			http.Error(w, "Missing name or id in path", http.StatusBadRequest)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *UploadHandler) handleUploadByName(w http.ResponseWriter, r *http.Request) {
	// Pretty name for the user, choose filename or user specified name.
	name := files.SanitizeFilename(r.PathValue("name"))
	if r.URL.Query().Get("name") != "" {
		name = files.SanitizeFilename(r.URL.Query().Get("name"))
	}

	// Unfunctional encryption, lets the user specify if the file is encrypted.
	// This is just for metadata purposes, the file is not necessarily encrypted.
	encrypted, err := returnEncryptionType(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Lets the user overwrite a file via its pretty name, if there is only one
	// file with that name. If there are multiple files with the same name,
	// the user must specify the ID of the file to overwrite.
	overwrite, err := returnOverwriteValue(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	matchingFiles := []database.File{}
	// Since the user is using a name, and there can be two of the same name+
	// We need to make sure that we are 100% overwritting the correct file.
	// Hence if there are more than 1, abort.
	if overwrite == "true" {
		var err error

		matchingFiles, err = h.fileService.GetByName(r.Context(), name)
		if err != nil {
			http.Error(w, "Failed to check for existing files", http.StatusInternalServerError)
			return
		}

		if len(matchingFiles) > 1 {
			http.Error(
				w,
				// TODO: Make this message friendlier, include possible ids etc.
				"At least two files with the same name exist. Unsure which to overwrite, aborting!",
				http.StatusConflict,
			)
			return
		}
	}
	h.createNewFileRecord(w, r, overwrite, name, encrypted, matchingFiles)

}

// This function is exclusively for overwriting, since the user isn't able to set a custom ID
func (h *UploadHandler) handleUploadById(w http.ResponseWriter, r *http.Request) {
	encrypted, err := returnEncryptionType(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file_to_overwrite, err := h.fileService.GetByID(r.Context(), r.PathValue("id"))
	if err != nil {
		http.Error(w, "Failed to get file by ID", http.StatusInternalServerError)
		return
	}

	matching_file := []database.File{*file_to_overwrite}

	unsanitised_name := r.URL.Query().Get("name")
	var new_name string
	if unsanitised_name != "" {
		new_name = files.SanitizeFilename(unsanitised_name)
	} else {
		new_name = file_to_overwrite.OriginalName
	}

	h.createNewFileRecord(w, r, "true", new_name, encrypted, matching_file)
}
