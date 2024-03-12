package controllers

import (
	"log/slog"
	"net/http"
)

func (ac *AppController) LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info(
			"request",
			slog.String("remote", r.RemoteAddr),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
		)
		next.ServeHTTP(w, r)
	})
}
