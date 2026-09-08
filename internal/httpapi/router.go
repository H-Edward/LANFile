package httpapi

import (
	"log"
	"net/http"

	"github.com/H-Edward/LANFile/internal/files"
	"github.com/go-chi/chi/v5"
)

func NewAPIRouter(fileService *files.Service, dataDir string) http.Handler {
	r := chi.NewRouter()
	log.Println("Setting up routes...")

	upload := NewUploadHandler(fileService, dataDir)
	download := NewDownloadHandler(fileService, dataDir)
	search := NewSearchHandler(fileService, dataDir)
	delete := NewDeleteHandler(fileService, dataDir)

	r.Method(http.MethodPut, "/u/name/{name}", upload)
	r.Method(http.MethodPut, "/u/id/{id}", upload) // Purely for overwritting, since the ID is something the user can't set

	r.Method(http.MethodGet, "/d/name/{name}", download)
	r.Method(http.MethodHead, "/d/name/{name}", download)

	r.Method(http.MethodGet, "/d/id/{id}", download)
	r.Method(http.MethodHead, "/d/id/{id}", download)

	r.Method(http.MethodGet, "/s/name/{name}", search)
	r.Method(http.MethodGet, "/s/id/{id}", search)
	r.Method(http.MethodGet, "/s/getall", search)

	r.Method(http.MethodDelete, "/delete/id/{id}", delete)

	return r
}
