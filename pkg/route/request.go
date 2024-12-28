package route

import "net/http"

type Request struct {
	Path   string
	Method string
}

func NewRequestFromHTTPRequest(r *http.Request) Request {
	return Request{Path: r.URL.Path, Method: r.Method}
}
