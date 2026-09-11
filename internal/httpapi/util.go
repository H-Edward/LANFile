package httpapi

import (
	"fmt"
	"net/http"

	"github.com/H-Edward/LANFile/internal/database"
	"github.com/H-Edward/LANFile/internal/files"
)

func ReturnEncryptionType(r *http.Request) (string, error) {
	encrypted := r.URL.Query().Get("encrypted")
	if encrypted != "key" && encrypted != "password" && encrypted != "" {
		return "", fmt.Errorf("Invalid encryption method")
	}
	if encrypted == "" {
		return "none", nil
	}
	return encrypted, nil
}

func ReturnOverwriteValue(r *http.Request) (string, error) {
	overwrite := r.URL.Query().Get("overwrite")
	if overwrite != "true" && overwrite != "false" && overwrite != "" {
		return "", fmt.Errorf("Invalid overwrite value")
	}
	if overwrite == "" {
		return "false", nil
	}
	return overwrite, nil
}

func ReturnRequestHasAuth(r *http.Request, file database.File) (bool, error) {
	if file.AuthorisationHash == "none" {
		return true, nil
	}

	_, secret, ok := r.BasicAuth()
	if !ok {
		return false, fmt.Errorf("Missing authorization header")
	}
	provided_secrets_hash := files.CheckPassword(secret, file.AuthorisationHash)
	if !provided_secrets_hash {
		return false, fmt.Errorf("Invalid authorization header")
	}

	return true, nil
}
