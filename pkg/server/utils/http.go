package utils

import (
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

const MaxUploadSize = 50 * 1024 * 1024 // 50MB

func SanitizePath(rawPath string) (string, error) {
	path := strings.TrimPrefix(filepath.Clean(rawPath), "/")
	if path == "" || strings.Contains(path, "..") {
		return "", fmt.Errorf("invalid path: %s", rawPath)
	}
	return path, nil
}

func ParseBearerToken(r *http.Request) (string, error) {
	if token := r.URL.Query().Get("token"); token != "" {
		return token, nil
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", fmt.Errorf("invalid authorization format")
	}
	return parts[1], nil
}

func DetermineContentType(headerType, path string) string {
	if headerType != "" {
		return headerType
	}
	if ct := mime.TypeByExtension(filepath.Ext(path)); ct != "" {
		return ct
	}
	return "application/octet-stream"
}

func WriteError(w http.ResponseWriter, message string, status int, err error) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprintf(w, "%s: %v", message, err)
}
