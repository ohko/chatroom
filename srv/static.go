package srv

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"runtime"
	"strings"
)

func HandleStatic(publicFolder embed.FS) {
	rootPath := "./public"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if runtime.GOOS == "darwin" {
			if _, err := os.Stat(rootPath + r.URL.Path); os.IsNotExist(err) {
				http.ServeFile(w, r, rootPath+"/index.html")
			} else {
				http.ServeFile(w, r, rootPath+r.URL.Path)
			}
		} else {
			fsys, _ := fs.Sub(publicFolder, "public")
			if _, err := fsys.Open(strings.TrimPrefix(r.URL.Path, "/")); err != nil {
				http.ServeFileFS(w, r, fsys, "index.html")
			} else {
				http.FileServer(http.FS(fsys)).ServeHTTP(w, r)
			}
		}
	})
}
