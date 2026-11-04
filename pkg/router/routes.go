package router

import (
	"crhuber/golinks/pkg/controllers"
	"crhuber/golinks/pkg/database"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"slices"
	"strings"
)

// This "go:embed" directive tells Go to embed static/* (recursively),
// and make it accessible as the staticFS variable.

//go:embed static/*
var staticFS embed.FS

func NewRouter(db *database.DbConnection) http.Handler {

	// server files from static folder
	serverRoot, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatal(err)
	}

	// link controller takes in a pointer to the db
	ac := controllers.NewAppController(db)

	mux := http.NewServeMux()
	// routes links
	mux.HandleFunc("GET /api/v1/links", getOnly(ac.GetLinks, http.MethodPost))
	mux.HandleFunc("POST /api/v1/links", ac.CreateLink)
	mux.HandleFunc("GET /api/v1/link/{id}", getOnly(ac.GetLink, http.MethodDelete, http.MethodPatch))
	mux.HandleFunc("DELETE /api/v1/link/{id}", ac.DeleteLink)
	mux.HandleFunc("PATCH /api/v1/link/{id}", ac.UpdateLink)
	mux.HandleFunc("GET /api/v1/search", getOnly(ac.SearchLinks))
	// tags
	mux.HandleFunc("GET /api/v1/tags", getOnly(ac.GetTags, http.MethodPost))
	mux.HandleFunc("POST /api/v1/tags", ac.CreateTag)
	mux.HandleFunc("GET /api/v1/tag/{id}", getOnly(ac.GetTag, http.MethodDelete, http.MethodPatch))
	mux.HandleFunc("DELETE /api/v1/tag/{id}", ac.DeleteTag)
	mux.HandleFunc("PATCH /api/v1/tag/{id}", ac.UpdateTag)
	//
	mux.HandleFunc("GET /healthz", getOnly(ac.Health))
	fileServer := http.FileServer(http.FS(serverRoot))
	// An all-method "/favicon.ico" overlaps GET /{keyword}: the literal path
	// is more specific, but it matches more methods, so ServeMux treats the
	// pair as a conflict. Register each method chi's Handle covered. GET also
	// matches HEAD, and an explicit HEAD pattern would conflict the same way.
	for _, method := range []string{
		http.MethodConnect,
		http.MethodDelete,
		http.MethodGet,
		http.MethodOptions,
		http.MethodPatch,
		http.MethodPost,
		http.MethodPut,
		http.MethodTrace,
	} {
		mux.Handle(method+" /favicon.ico", fileServer)
	}
	// {$} is an exact match for "/". A bare "/" pattern is a catch-all and
	// would swallow keyword redirects.
	mux.Handle("/{$}", fileServer)
	mux.HandleFunc("GET /{keyword}", getOnly(ac.GetKeyword))
	// Any method, matching the previous HandleFunc registration.
	mux.HandleFunc("/{keyword}/{subkey}", ac.GetKeyword)
	// LogRequest previously wrapped every route, including 404 and 405.
	return ac.LogRequest(mux)
}

// getOnly serves GET. A "GET" ServeMux pattern also matches HEAD, but these
// routes were registered with chi's Get, which does not. HEAD on a keyword
// would otherwise redirect and increment the view count.
// alsoAllowed lists other methods registered on the same path so the 405
// Allow header still names them.
func getOnly(next http.HandlerFunc, alsoAllowed ...string) http.HandlerFunc {
	allowed := append([]string{http.MethodGet}, alsoAllowed...)
	slices.Sort(allowed)
	allow := strings.Join(allowed, ", ")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", allow)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		next(w, r)
	}
}
