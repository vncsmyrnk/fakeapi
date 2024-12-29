package route

import (
	"fmt"
	"net/http"
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
