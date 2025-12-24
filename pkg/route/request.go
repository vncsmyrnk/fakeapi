package route

import (
	"fmt"
	"net/http"
	"regexp"
)

type Request struct {
	Path   string
	Method string
}

func NewRequestFromHTTPRequest(r *http.Request) Request {
	return Request{Path: r.URL.Path, Method: r.Method}
}

func (r Request) String() string {
	return fmt.Sprintf("%s %s", r.Method, r.Path)
}

func (r Request) Endpoint(endpoints []Endpoint) (*Endpoint, error) {
	for _, endpoint := range endpoints {
		matches, err := regexp.MatchString(
			r.endpointInputPathForRegexpMatch(endpoint.InputPath), r.Path)
		if err != nil {
			continue
		}

		if matches {
			return &endpoint, nil
		}
	}
	return nil, fmt.Errorf("request did not match any endpoint")
}

func (r Request) endpointInputPathForRegexpMatch(endpointInputPath string) string {
	return fmt.Sprintf("%s$", endpointInputPath)
}
