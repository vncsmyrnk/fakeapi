package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	apiHTTP "fakeapi/internal/handler/http"
)

const (
	serverBaseURL  = "http://localhost"
	getRequestsURI = "/requests"
)

func requestURL(baseURL string, port int) string {
	return fmt.Sprintf("%s:%d", baseURL, port)
}

func getRequests(url string) ([]apiHTTP.RequestResponse, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", url, getRequestsURI), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-FakeAPI-Control", "meta")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var requests []apiHTTP.RequestResponse
	err = json.Unmarshal(b, &requests)
	if err != nil {
		return nil, err
	}

	return requests, nil
}

func loggerGenerator(verbose bool) func(log.Level, log.Fields, string, ...any) {
	l := log.StandardLogger()
	return func(level log.Level, fields log.Fields, format string, args ...any) {
		switch level {
		case log.FatalLevel:
			fmt.Println(fmt.Sprintf(format, args...))
			os.Exit(1)
		case log.InfoLevel:
			fmt.Println(fmt.Sprintf(format, args...))
		default:
			if verbose {
				l.WithFields(fields).Logln(level, fmt.Sprintf(format, args...))
			}
		}
	}
}

func main() {
	port := flag.Int("port", 8080, "Port the target server is running on")
	verbose := flag.Bool("v", false, "Verbose")
	flag.Parse()

	logger := loggerGenerator(*verbose)

	url := requestURL(serverBaseURL, *port)
	logger(log.DebugLevel, log.Fields{"baseURL": url}, "requesting API")
	requests, err := getRequests(url)
	if err != nil {
		logger(log.FatalLevel, nil, "failed to fetch requests: %v", err)
	}

	args := flag.Args()
	if len(args) < 2 {
		logger(log.FatalLevel, nil, "missing required arguments")
	}

	method := args[0]
	uri := args[1]

	for _, r := range requests {
		if r.Method != method || r.URI != uri {
			continue
		}

		logger(log.InfoLevel, nil, "assertion succeeded")
		os.Exit(0)
	}

	logger(log.FatalLevel, log.Fields{"method": method, "uri": uri}, "assert failed")
}
