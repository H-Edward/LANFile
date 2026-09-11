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

// createNewFileRecord handles the creation of a new file record in the database and saves the uploaded file to disk. It also handles overwriting existing files if specified.
func (h *UploadHandler) createNewFileRecord(w http.ResponseWriter, r *http.Request, overwrite string, name string, encrypted string, matchingFiles []database.File) {

	// Auth
	if overwrite == "true" && len(matchingFiles) == 1 {
		file := matchingFiles[0]
		hasAuth, err := ReturnRequestHasAuth(r, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if !hasAuth {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	storageKey := uuid.New().String()
	path := filepath.Join(h.dataDir, "files", storageKey)

	dst, err := os.Create(path)
	if err != nil {
		fmt.Println("Failed to create file:", err)
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	buf := make([]byte, 512)
	n, err := io.ReadFull(r.Body, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		_ = os.Remove(path)
		http.Error(w, "Upload failed", http.StatusInternalServerError)
		return
	}
	mimeType := http.DetectContentType(buf[:n])
	if _, err := dst.Write(buf[:n]); err != nil {
		_ = os.Remove(path)
		http.Error(w, "Upload failed", http.StatusInternalServerError)
		return
	}

	written, err := io.Copy(dst, r.Body)
	if err != nil {
		_ = os.Remove(path)
		http.Error(w, "Upload failed", http.StatusInternalServerError)
		return
	}

	size := int64(n) + written

	var authHash string = "none"

	// At this stage, the user has been authed
	if overwrite == "true" && len(matchingFiles) == 1 {
		file := matchingFiles[0]
		authHash = file.AuthorisationHash
	}
	if overwrite == "false" {
		_, secret, ok := r.BasicAuth()
		if ok && secret != "" {
			authHash, err = files.GenerateHash(secret)
			if err != nil {
				_ = os.Remove(path)
				http.Error(w, "Failed to generate authorization hash", http.StatusInternalServerError)
				return
			}
		}
	}

	input := files.CreateFileInput{
		OriginalName:      name,
		ContentType:       mimeType,
		Size:              size,
		StorageKey:        storageKey,
		Encrypted:         encrypted,
		NeedsAuth:         authHash != "none",
		AuthorisationHash: authHash,
	}
	var file *database.File
	// If overwrite is true and there is exactly one matching file, overwrite it.
	if overwrite == "true" && len(matchingFiles) == 1 {
		oldFile := matchingFiles[0]
		input.NeedsAuth = oldFile.NeedsAuth
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
	encrypted, err := ReturnEncryptionType(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Lets the user overwrite a file via its pretty name, if there is only one
	// file with that name. If there are multiple files with the same name,
	// the user must specify the ID of the file to overwrite.
	overwrite, err := ReturnOverwriteValue(r)
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
	encrypted, err := ReturnEncryptionType(r)
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
