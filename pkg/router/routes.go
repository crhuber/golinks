package router

import (
	"crhuber/golinks/pkg/controllers"
	"crhuber/golinks/pkg/database"
	"embed"
	"io/fs"
	"log"
	"net/http"
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
	mux.HandleFunc("GET /api/v1/links", ac.GetLinks)
	mux.HandleFunc("POST /api/v1/links", ac.CreateLink)
	mux.HandleFunc("GET /api/v1/link/{id}", ac.GetLink)
	mux.HandleFunc("DELETE /api/v1/link/{id}", ac.DeleteLink)
	mux.HandleFunc("PATCH /api/v1/link/{id}", ac.UpdateLink)
	mux.HandleFunc("GET /api/v1/search", ac.SearchLinks)
	// tags
	mux.HandleFunc("GET /api/v1/tags", ac.GetTags)
	mux.HandleFunc("POST /api/v1/tags", ac.CreateTag)
	mux.HandleFunc("GET /api/v1/tag/{id}", ac.GetTag)
	mux.HandleFunc("DELETE /api/v1/tag/{id}", ac.DeleteTag)
	mux.HandleFunc("PATCH /api/v1/tag/{id}", ac.UpdateTag)
	//
	mux.HandleFunc("GET /healthz", ac.Health)
	fileServer := http.FileServer(http.FS(serverRoot))
	// GET also matches HEAD. Other methods 405 instead of serving the file.
	mux.Handle("GET /favicon.ico", fileServer)
	// {$} is an exact match for "/". A bare "/" pattern is a catch-all and
	// would swallow keyword redirects.
	mux.Handle("/{$}", fileServer)
	mux.HandleFunc("GET /{keyword}", ac.GetKeyword)
	// Any method, matching the previous HandleFunc registration.
	mux.HandleFunc("/{keyword}/{subkey}", ac.GetKeyword)
	// LogRequest previously wrapped every route, including 404 and 405.
	return ac.LogRequest(mux)
}
