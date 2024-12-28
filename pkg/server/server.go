package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	"fakeapi/pkg/route"
)

type Config struct {
	Port      int16
	Endpoints route.Endpoints
}

func handler(w http.ResponseWriter, r *http.Request, endpoints route.Endpoints) {
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

func setupLog() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func Run(config Config) {
	setupLog()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, config.Endpoints)
	})
	http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)

}
