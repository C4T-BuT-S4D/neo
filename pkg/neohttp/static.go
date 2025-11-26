package neohttp

import (
	"net/http"
	"os"
	"path/filepath"
)

func StaticHandler(dir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join(dir, r.URL.Path)

		if r.URL.Path == "" || r.URL.Path == "/" {
			filePath = filepath.Join(dir, "index.html")
		} else if _, err := os.Stat(filePath); err != nil {
			filePath = filepath.Join(dir, "index.html")
		}

		http.ServeFile(w, r, filePath)
	}
}
