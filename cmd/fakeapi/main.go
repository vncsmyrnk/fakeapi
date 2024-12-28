package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
)

type cliArgs struct {
	HTTPMethod   string
	EndpointName string
}

type Response struct {
	Message string `json:"message"`
}

const (
	MethodGet     = "GET"
	MethodHead    = "HEAD"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodPatch   = "PATCH"
	MethodDelete  = "DELETE"
	MethodConnect = "CONNECT"
	MethodOptions = "OPTIONS"
	MethodTrace   = "TRACE"
)

type Endpoint struct {
	Method   string
	Response map[string]string
}

func handler(w http.ResponseWriter, r *http.Request, args *cliArgs) {
	if r.Method != args.HTTPMethod {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path != fmt.Sprintf("/%s", args.EndpointName) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	response := Response{Message: "Hello, World!"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func newCLIArgs() *cliArgs {
	httpMethod := flag.String("method", "GET", "HTTP Method for the fake endpoint")
	endpointName := flag.String("endpoint", "", "Fake endpoint's name")
	flag.Parse()

	return &cliArgs{
		HTTPMethod:   *httpMethod,
		EndpointName: *endpointName,
	}
}

func main() {
	cliArgs := newCLIArgs()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, cliArgs)
	})
	http.ListenAndServe(":8080", nil)
}
