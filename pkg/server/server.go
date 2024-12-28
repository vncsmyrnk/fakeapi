package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	"fakeapi/pkg/route"
)

type server struct {
	Port      int16
	Endpoints route.Endpoints
}

func NewServer(port int16, endpoints route.Endpoints) server {
	return server{
		Port:      port,
		Endpoints: endpoints,
	}
}

func (s server) Run() {
	setupLog()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveHTTP(w, r, s.Endpoints)
	})
	http.ListenAndServe(fmt.Sprintf(":%d", s.Port), nil)
}

func setupLog() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func serveHTTP(w http.ResponseWriter, r *http.Request, endpoints route.Endpoints) {
	request := route.NewRequestFromHTTPRequest(r)
	activeEndpoint, err := endpoints.Requested(request)
	if err != nil {
		log.Error(fmt.Sprintf("%s %s not found", request.Method, request.Path))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	log.Info(fmt.Sprintf("%s %s", request.Method, request.Path))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activeEndpoint.OutputContent)
}
