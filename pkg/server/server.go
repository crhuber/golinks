package server

import (
	"crhuber/golinks/pkg/controllers"
	"crhuber/golinks/pkg/database"
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// This "go:embed" directive tells Go to embed static/* (recursively),
// and make it accessible as the staticFS variable.

//go:embed static/*
var staticFS embed.FS

func NewRouter(db *database.DbConnection) *chi.Mux {

	// server files from static folder
	serverRoot, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatal(err)
	}

	// link controller takes in a pointer to the db
	ac := controllers.NewAppController(db)

	router := chi.NewRouter()
	// routes links
	router.Use(ac.LogRequest)
	router.Get("/api/v1/links", ac.GetLinks)
	router.Post("/api/v1/links", ac.CreateLink)
	router.Get("/api/v1/link/{id}", ac.GetLink)
	router.Delete("/api/v1/link/{id}", ac.DeleteLink)
	router.Patch("/api/v1/link/{id}", ac.UpdateLink)
	router.Get("/api/v1/search", ac.SearchTags)
	// tags
	router.Get("/api/v1/tags", ac.GetTags)
	router.Post("/api/v1/tags", ac.CreateTag)
	router.Get("/api/v1/tag/{id}", ac.GetTag)
	router.Delete("/api/v1/tag/{id}", ac.DeleteTag)
	router.Patch("/api/v1/tag/{id}", ac.UpdateTag)
	//
	router.Get("/healthz", ac.Health)
	router.Handle("/favicon.ico", http.FileServer(http.FS(serverRoot)))
	router.Get("/{keyword}", ac.GetKeyword)
	router.HandleFunc("/{keyword}/{subkey}", ac.GetKeyword)
	router.Handle("/", http.FileServer(http.FS(serverRoot)))
	return router
}
