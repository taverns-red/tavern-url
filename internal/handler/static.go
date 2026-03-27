package handler

import (
	"net/http"
)

// StaticFileServer returns an HTTP handler that serves static files from disk.
// In production, static files will be embedded. For now, serve from the filesystem.
func StaticFileServer(staticDir string) http.Handler {
	return http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir)))
}
