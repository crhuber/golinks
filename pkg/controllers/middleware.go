package controllers

import (
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

func (ac *AppController) LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		escapedUrl := strings.Replace(r.URL.String(), "\n", "", -1)
		escapedUrl = strings.Replace(escapedUrl, "\r", "", -1)
		logString := fmt.Sprintf("%v %v %v", r.RemoteAddr, r.Method, escapedUrl)
		log.Info(logString)
		next.ServeHTTP(w, r)
	})
}
