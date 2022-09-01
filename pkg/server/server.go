package server

import (
	"crhuber/golinks/pkg/controllers"
	"crhuber/golinks/pkg/database"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

func NewRouter(db *database.DbConnection, staticPath string) *mux.Router {

	// serve static files relative to location of executable path
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	staticPath = filepath.Join(exeDir, staticPath)
	// link controller takes in a pointer to the db
	ac := controllers.NewAppController(db, staticPath)

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
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))))
	router.HandleFunc("/favicon.ico", ac.HandleFavicon)
	router.HandleFunc("/{keyword}", ac.GetKeyword).Methods("GET")
	router.HandleFunc("/{keyword}/{subkey}", ac.GetKeyword).Methods("GET")
	router.HandleFunc("/", ac.HandleHome)
	return router
}
