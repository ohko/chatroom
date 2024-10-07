package srv

import (
	"embed"
	"net/http"
	"os"
	"runtime"
	"strings"
)

func HandleStatic(indexFile embed.FS) {
	rootPath := "./public"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if runtime.GOOS == "darwin" {
			if _, err := os.Stat(rootPath + r.URL.Path); os.IsNotExist(err) {
				http.ServeFile(w, r, rootPath+"/index.html")
			} else {
				http.ServeFile(w, r, rootPath+r.URL.Path)
			}
		} else {
			if _, err := indexFile.Open(strings.TrimPrefix(r.URL.Path, "/")); err != nil {
				http.ServeFileFS(w, r, indexFile, "index.html")
			} else {
				http.FileServer(http.FS(indexFile)).ServeHTTP(w, r)
			}
		}
	})
}
