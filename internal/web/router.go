package web

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/H-Edward/LANFile/internal/files"
	"github.com/go-chi/chi/v5"
)

//go:embed templates/*.html templates/partials/*.html
var templateFS embed.FS

type Handler struct {
	files     *files.Service
	templates *template.Template
}

func NewWebRouter(fileService *files.Service) http.Handler {
	h := &Handler{
		files: fileService,
		templates: template.Must(
			template.New("base.html").Funcs(template.FuncMap{
				"formatSize": formatSize,
			}).ParseFS(templateFS, "templates/*.html", "templates/partials/*.html"),
		),
	}

	r := chi.NewRouter()

	r.Get("/", h.Index)
	r.Get("/upload", h.Upload)
	r.Get("/files", h.Files)

	return r
}
