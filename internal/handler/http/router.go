package http

import (
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"

	"fakeapi/internal/domain"
	"fakeapi/internal/port"
)

// NewMainHandler returns the main HTTP handler that acts as the fake API
// and routes meta requests if 'x-fakeapi-control' header is present.
func NewMainHandler(endSvc port.EndpointService, reqSvc port.RequestService, metaHandler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Infof("%s %s", r.Method, r.URL.Path)

		if r.Header.Get("x-fakeapi-control") == "meta" {
			metaHandler.ServeHTTP(w, r)
			return
		}

		e, err := endSvc.MatchRequest(r.Context(), r)
		if err != nil {
			if errors.Is(err, domain.ErrEndpointNotFound) {
				log.Errorf("no endpoint matches this request")
				w.WriteHeader(http.StatusNotFound)
			} else {
				log.Errorf("failed to find a endpoint for the request")
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		err = reqSvc.Create(r.Context(), e.ID, r)
		if err != nil {
			log.Errorf("failed to save request history: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(e.StatusCode)
		log.Infof("endpoint returned %d", e.StatusCode)
		if e.Content != nil {
			_, _ = w.Write([]byte(*e.Content))
		}
	}
}
