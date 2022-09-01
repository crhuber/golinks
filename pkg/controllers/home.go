package controllers

import (
	"io"
	"net/http"
	"os"
)

func (c *AppController) HandleHome(w http.ResponseWriter, r *http.Request) {
	indexFile, err := os.Open("./static/index.html")
	if err != nil {
		io.WriteString(w, "error reading index")
		return
	}
	defer indexFile.Close()

	io.Copy(w, indexFile)
}

func (c *AppController) HandleFavicon(w http.ResponseWriter, r *http.Request) {
	indexFile, err := os.Open("./static/favicon.ico")
	if err != nil {
		io.WriteString(w, "error reading favicon")
		return
	}
	defer indexFile.Close()

	io.Copy(w, indexFile)
}
