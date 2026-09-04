package main

import (
	"os"
	"net/http"

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
	dataDir := getenv("LANFILE_DATA", "/data")
	addr := getenv("LANFILE_ADDR", ":8021")

	http.ListenAndServe(addr, nil)
}
