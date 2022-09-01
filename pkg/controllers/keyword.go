package controllers

import (
	"crhuber/golinks/pkg/models"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func (c *AppController) updateViewCount(link *models.Link) {
	link.Views += 1
	c.db.Db.Save(link)
}

func (c *AppController) GetKeyword(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	link := models.Link{}
	keyword := params["keyword"]
	// hack to match urls like foo/bar work
	if params["subkey"] != "" {
		keyword = fmt.Sprintf("%s/%s", keyword, params["subkey"])
	}
	err := c.db.Db.First(&link, "keyword = ?", keyword).Error
	if err != nil {
		// handle programatic links
		log.Info("keyword not found  in exact match. trying wildcard")
		keywordParts := strings.Split(keyword, "/")
		err := c.db.Db.First(&link, "keyword = ?", fmt.Sprintf("%s/{*}", keywordParts[0])).Error
		if err != nil {
			http.Redirect(w, r, fmt.Sprintf("/?q=%s", keyword), http.StatusTemporaryRedirect)
			return
		}
		programmaticDestination := strings.Replace(link.Destination, "{*}", keywordParts[1], 1)
		go c.updateViewCount(&link)
		http.Redirect(w, r, programmaticDestination, http.StatusTemporaryRedirect)
		return
	}
	go c.updateViewCount(&link)
	http.Redirect(w, r, link.Destination, http.StatusTemporaryRedirect)
}
