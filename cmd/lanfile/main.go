package main

import (
	"log"
	"net/http"
	"os"
	"path"

	"github.com/H-Edward/LANFile/internal/database"
	"github.com/H-Edward/LANFile/internal/files"
	"github.com/H-Edward/LANFile/internal/httpapi"
	"github.com/H-Edward/LANFile/internal/web"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
		
	}
	
	dataDir := getenv("LANFILE_DATA", "./data")
	addr := getenv("LANFILE_PORT", "8022")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("failed to create data directory: %v", err)
	}
	filesDir := path.Join(dataDir, "files")
	if err := os.MkdirAll(filesDir, 0755); err != nil {
		log.Fatalf("failed to create files directory: %v", err)
	}

	db, err := database.Open(getenv("LANFILE_DB", "./data.db"))
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	if err := database.Ping(db); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	r := chi.NewRouter()

	fileService := files.NewService(db)

	webRouter := web.NewWebRouter(fileService)
	r.Mount("/", webRouter)

	APIRouter := httpapi.NewAPIRouter(fileService, dataDir)
	r.Mount("/api", APIRouter)

	log.Printf("Starting server on %s...", addr)
	addr = ":" + addr
	http.ListenAndServe(addr, r)
}
