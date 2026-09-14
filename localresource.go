package main

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// localResourcePath is the URL under which the preview loads local images
// referenced by the open document, e.g. /local-resource?path=E%3A%5Cdocs%5Cimg.png.
// It is relative to the app origin, so it works with every platform's scheme
// (wails://localhost on macOS/Linux, http://wails.localhost on Windows).
const localResourcePath = "/local-resource"

// localImageTypes lists the file types the preview may load from disk.
var localImageTypes = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".webp": "image/webp",
	".avif": "image/avif",
	".svg":  "image/svg+xml",
	".bmp":  "image/bmp",
	".ico":  "image/x-icon",
}

// assetHandler serves local images under localResourcePath and the embedded
// frontend bundle for everything else.
func (a *App) assetHandler() http.Handler {
	fileServer := application.AssetFileServerFS(assets)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == localResourcePath {
			a.serveLocalResource(w, r)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}

// serveLocalResource serves a local image file, but only from directories the
// user has whitelisted and only for image file types.
func (a *App) serveLocalResource(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" || !filepath.IsAbs(path) {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	contentType, ok := localImageTypes[strings.ToLower(filepath.Ext(path))]
	if !ok {
		http.Error(w, "unsupported file type", http.StatusForbidden)
		return
	}
	if a.config == nil || !a.config.IsPathAllowed(path) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if contentType == "image/svg+xml" {
		// SVGs can carry scripts; never let one run if it is opened directly.
		w.Header().Set("Content-Security-Policy", "script-src 'none'")
	}
	http.ServeFile(w, r, filepath.Clean(path))
}
