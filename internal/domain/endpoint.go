package domain

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
)

var (
	ErrEndpointNotFound = errors.New("endpoint not found")
)

// Endpoint represents a mocked API endpoint.
type Endpoint struct {
	ID         int
	Path       string
	Method     string
	StatusCode int
	Response   *string
}

// MatchesRequest checks if the Endpoint matches the given HTTP request.
func (e Endpoint) MatchesRequest(r *http.Request) bool {
	if r.Method != e.Method {
		return false
	}

	match, _ := regexp.MatchString(fmt.Sprintf("%s$", e.Path), r.URL.Path)
	return match
}
