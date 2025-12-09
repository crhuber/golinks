package controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type jsonErr struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

func JsonResponse(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("Error encoding json", slog.Any("error", err))
		JsonError(w, err, http.StatusBadRequest, "Error encoding json")
	}
}

func JsonError(w http.ResponseWriter, err error, status int, text string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(jsonErr{Code: status, Text: text}); err != nil {
		slog.Error("Error encoding json", slog.Any("error", err))
	}
}
