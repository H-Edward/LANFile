package main

import (
	"log"
	"net/http"
	"os"

	"github.com/H-Edward/LANFile/internal/database"
	"github.com/H-Edward/LANFile/internal/files"
	"github.com/H-Edward/LANFile/internal/httpapi"
	"github.com/joho/godotenv"
)

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
func main() {
	godotenv.Load()
	dataDir := getenv("LANFILE_DATA", "./data")
	addr := getenv("LANFILE_ADDR", ":8021")

	db, err := database.Open(getenv("LANFILE_DB", "./data.db"))
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	if err := database.Ping(db); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	fileService := files.NewService(db)
	router := httpapi.NewRouter(fileService, dataDir)

	log.Printf("Starting server on %s...", addr)

	http.ListenAndServe(addr, router)
}
