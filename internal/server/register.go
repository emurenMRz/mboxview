package server

import (
	"mime"
	"net/http"
	"path/filepath"
)

// RegisterHandlers registers HTTP handlers for the server. Call this before ListenAndServe.
func RegisterHandlers(path string) {
	basePath = path

	http.HandleFunc("/api/mailboxes/", handleMailboxRoutes)

	// Serve static files under /static/ (API lives under /api/)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Ensure common MIME types are set (some platforms lack .css/.js by default)
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".js", "application/javascript")

	// Serve index at root
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "index.html"))
	})
}

// ListenAndServe is a thin wrapper to allow main to call server.ListenAndServe
func ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, nil)
}

// SetEditMode enables or disables edit mode for the server.
func SetEditMode(v bool) {
	editMode = v
}
