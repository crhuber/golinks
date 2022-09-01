package controllers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func (c *AppController) HandleHome(w http.ResponseWriter, r *http.Request) {
	index := filepath.Join(c.staticPath, "index.html")
	indexFile, err := os.Open(index)
	if err != nil {
		io.WriteString(w, "error reading index")
		return
	}
	defer indexFile.Close()

	io.Copy(w, indexFile)
}

func (c *AppController) HandleFavicon(w http.ResponseWriter, r *http.Request) {
	favicon := filepath.Join(c.staticPath, "favicon.ico")
	faviconFile, err := os.Open(favicon)
	if err != nil {
		io.WriteString(w, "error reading favicon")
		return
	}
	defer faviconFile.Close()

	io.Copy(w, faviconFile)
}
