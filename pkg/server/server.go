package server

import (
	"crhuber/golinks/pkg/controllers"
	"crhuber/golinks/pkg/database"
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// This "go:embed" directive tells Go to embed static/* (recursively),
// and make it accessible as the staticFS variable.

//go:embed static/*
var staticFS embed.FS

func NewRouter(db *database.DbConnection) *mux.Router {

	// server files from static folder
	serverRoot, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatal(err)
	}

	// link controller takes in a pointer to the db
	ac := controllers.NewAppController(db)

	router := mux.NewRouter().StrictSlash(true)
	// routes links
	router.Use(ac.LogRequest)
	router.HandleFunc("/api/v1/links", ac.GetLinks).Methods("GET")
	router.HandleFunc("/api/v1/links", ac.CreateLink).Methods("POST")
	router.HandleFunc("/api/v1/link/{id}", ac.GetLink).Methods("GET")
	router.HandleFunc("/api/v1/link/{id}", ac.DeleteLink).Methods("DELETE")
	router.HandleFunc("/api/v1/link/{id}", ac.UpdateLink).Methods("PATCH")
	router.HandleFunc("/api/v1/search", ac.SearchTags).Methods("GET")
	// tags
	router.HandleFunc("/api/v1/tags", ac.GetTags).Methods("GET")
	router.HandleFunc("/api/v1/tags", ac.CreateTag).Methods("POST")
	router.HandleFunc("/api/v1/tag/{id}", ac.GetTag).Methods("GET")
	router.HandleFunc("/api/v1/tag/{id}", ac.DeleteTag).Methods("DELETE")
	router.HandleFunc("/api/v1/tag/{id}", ac.UpdateTag).Methods("PATCH")
	//
	router.HandleFunc("/healthz", ac.Health)
	router.PathPrefix("/favicon.ico").Handler(http.FileServer(http.FS(serverRoot)))
	router.HandleFunc("/{keyword}", ac.GetKeyword).Methods("GET")
	router.HandleFunc("/{keyword}/{subkey}", ac.GetKeyword).Methods("GET")
	router.PathPrefix("/").Handler(http.FileServer(http.FS(serverRoot)))
	return router
}
