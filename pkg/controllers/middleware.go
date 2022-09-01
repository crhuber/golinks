package controllers

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func (ac *AppController) LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logString := fmt.Sprintf("%v %v %v", r.RemoteAddr, r.Method, r.URL)
		log.Info(logString)
		next.ServeHTTP(w, r)
	})
}
