package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/tidwall/buntdb"

	"github.com/thinkofher/lalyta/pkg/api"
	"github.com/thinkofher/lalyta/pkg/service/params"
	"github.com/thinkofher/lalyta/pkg/storage"
)

func run() error {
	bunt, err := buntdb.Open(AppConfig.Bolt.StorageFile)
	if err != nil {
		return fmt.Errorf("buntdb.Open: %w", err)
	}
	defer bunt.Close()

	buntStorage := storage.New(bunt)

	chiParams := new(params.Chi)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// workDir, _ := os.Getwd()
	// filesDir := http.Dir(filepath.Join(workDir, "static"))
	// FileServer(r, "/favicon", filesDir)
	r.Get("/", api.FrontPage())
	r.Get("/favicon.ico", api.FaviconHandler())
	r.Get("/favicon.ico/", api.FaviconHandler())
	maxSyncSize := int64(AppConfig.Server.MaxSyncSizeKb * 1024)
	r.Get("/info", api.Info(maxSyncSize, "PL", "Hello World!", "1.1.13"))
	r.Post("/bookmarks", api.CreateBookmarks(buntStorage))
	r.Get("/bookmarks/{id}", api.Bookmarks(buntStorage, chiParams))
	r.Put("/bookmarks/{id}", api.UpdateBookmarks(buntStorage, chiParams))
	r.Get("/bookmarks/{id}/lastUpdated", api.LastUpdated(buntStorage, chiParams))
	r.Get("/bookmarks/{id}/version", api.Version(buntStorage, chiParams))

	port := fmt.Sprintf(":%d", AppConfig.Server.Port)
	log.Println("Starting server at 0.0.0.0" + port)

	return http.ListenAndServe("0.0.0.0"+port, r)
}

func main() {
	LoadConfig()
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	// if strings.ContainsAny(path, "{}*") {
	// 	panic("FileServer does not permit any URL parameters.")
	// }

	// if path != "/" && path[len(path)-1] != '/' {
	// 	r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
	// 	path += "/"
	// }
	// path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
