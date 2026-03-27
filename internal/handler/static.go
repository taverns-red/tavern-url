package handler

import (
	"io/fs"
	"net/http"
)

// StaticFileServer returns an HTTP handler that serves static files.
// If embeddedFS is non-nil, files are served from the embedded filesystem (prod).
// Otherwise, files are served from disk (dev).
func StaticFileServer(staticDir string, embeddedFS fs.FS) http.Handler {
	if embeddedFS != nil {
		return http.StripPrefix("/static/", http.FileServer(http.FS(embeddedFS)))
	}
	return http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir)))
}
