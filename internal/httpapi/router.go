package httpapi

import (
	"log"
	"net/http"

	"github.com/H-Edward/LANFile/internal/files"
	"github.com/go-chi/chi/v5"
)

func NewRouter(fileService *files.Service, dataDir string) http.Handler {
	r := chi.NewRouter()
	log.Println("Setting up routes...")

	upload := NewUploadHandler(fileService, dataDir)
	download := NewDownloadHandler(fileService, dataDir)

	r.Method(http.MethodPut, "/u/name/{name}", upload)
	r.Method(http.MethodPut, "/u/key/{key}", upload)

	r.Get("/d/name/{name}", download.GetbyName)
	r.Head("/d/name/{name}", download.GetbyName)

	r.Get("/d/id/{id}", download.GetbyID)
	r.Head("/d/id/{id}", download.GetbyID)


	r.Delete("/delete/name/{name}", deleteFileHandler)
	r.Delete("/delete/key/{key}", deleteFileHandler)


	return r
}
