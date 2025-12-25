package route

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
)

type Request struct {
	Path   string
	Method string
	Body   []byte
}

func NewRequestFromHTTPRequest(r *http.Request) (Request, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return Request{}, err
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	return Request{
		Path:   r.URL.Path,
		Method: r.Method,
		Body:   body,
	}, nil
}

func (r Request) String() string {
	return fmt.Sprintf("%s %s", r.Method, r.Path)
}

func (r Request) Endpoint(endpoints []Endpoint) (*Endpoint, error) {
	for _, endpoint := range endpoints {
		requestMatchesEndpoint, err := r.Matches(endpoint)
		if err != nil {
			log.Printf("endpoint %s has regexp syntax errors\n", endpoint.String())
			continue
		}

		if requestMatchesEndpoint {
			return &endpoint, nil
		}
	}
	return nil, fmt.Errorf("request did not match any endpoint")
}

func (r Request) Matches(endpoint Endpoint) (bool, error) {
	pathMatches, err := regexp.MatchString(
		r.endpointInputPathForRegexpMatch(endpoint.InputPath), r.Path)
	if err != nil {
		return false, err
	}

	if pathMatches && r.Method == endpoint.InputMethod {
		return true, nil
	}

	return false, nil
}

func (r Request) endpointInputPathForRegexpMatch(endpointInputPath string) string {
	return fmt.Sprintf("%s$", endpointInputPath)
}
