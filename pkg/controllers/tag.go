package controllers

import (
	"crhuber/golinks/pkg/models"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// anything of type tag controller has these functions available
func (c *AppController) GetTag(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tag := models.Tag{}
	err := c.db.Db.First(&tag, id).Error
	if err != nil {
		JsonError(w, err, http.StatusNotFound, "Not Found")
		return
	}
	json.NewEncoder(w).Encode(tag)
}

func (c *AppController) GetTags(w http.ResponseWriter, r *http.Request) {
	tags := []models.Tag{}
	c.db.Db.Find(&tags)
	json.NewEncoder(w).Encode(tags)
}

func (c *AppController) CreateTag(w http.ResponseWriter, r *http.Request) {
	tag := models.Tag{}
	json.NewDecoder(r.Body).Decode(&tag)
	err := tag.Validate()
	if err != nil {
		slog.Error("error validating fields", slog.Any("error", err))
		JsonError(w, err, http.StatusBadRequest, err.Error())
		return
	}
	c.db.Db.Create(&tag)
	JsonResponse(w, tag)
}

func (c *AppController) UpdateTag(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tag := models.Tag{}
	err := c.db.Db.First(&tag, id).Error
	if err != nil {
		JsonError(w, err, http.StatusNotFound, "Not Found")
		return
	}
	json.NewDecoder(r.Body).Decode(&tag)
	c.db.Db.Save(&tag)
	JsonResponse(w, tag)
}

func (c *AppController) DeleteTag(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tag := models.Tag{}
	err := c.db.Db.Unscoped().Delete(&tag, id).Error
	if err != nil {
		// If we didn't find it, 404
		JsonError(w, err, http.StatusNotFound, "Not Found")
		return
	}
	JsonResponse(w, "Tag Deleted")
}
