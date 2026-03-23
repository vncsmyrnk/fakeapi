package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"

	apiHTTP "fakeapi/internal/handler/http"
)

const (
	serverBaseURL     = "http://localhost"
	postAssertionsURI = "/assertions"
)

var client = &http.Client{
	Timeout: 10 * time.Second,
}

var (
	CliVersion = "dev"
)

type assertionResult struct {
	Success  bool
	Title    string
	Messages []string
}

func main() {
	port := flag.IntP("port", "p", 8080, "Port the target server is running on")
	quiet := flag.BoolP("quiet", "q", false, "Quiet mode")
	version := flag.BoolP("version", "v", false, "Display the current version")

	var (
		headers, attrs []string
	)
	flag.StringArrayVarP(&headers, "header", "H", []string{}, "Expected headers in 'Key: Value' format")
	flag.StringArrayVarP(&attrs, "body-attributes", "b", []string{}, "Expected body attributes in 'key=value' format")
	occurences := flag.IntP("occurrences", "c", -1, "Match occurrence count")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Assert Fake API requests\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  fakeassert <method> <uri> [flags]\n\n")

		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  fakeassert POST /items -H 'user-agent:curl' -a 'status=active'\n\n")

		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()
	if *version {
		fmt.Printf("%s\n", CliVersion)
		os.Exit(0)
	}

	if flag.NArg() == 1 || flag.NArg() > 2 {
		flag.Usage()
		os.Exit(1)
	}

	if flag.NArg() == 0 && *occurences == -1 {
		flag.Usage()
		os.Exit(1)
	}

	url := requestURL(serverBaseURL, *port)

	expectedHeaders := headersParsed(headers)
	expectedBodyAttrs := attributesParsed(attrs)

	var method, uri string
	args := flag.Args()
	if len(args) >= 2 {
		method = args[0]
		uri = args[1]
	}

	if *occurences == -1 {
		occurences = nil
	}

	assertion := apiHTTP.AssertionRequest{
		Method:  method,
		URI:     uri,
		Headers: expectedHeaders,
		Body:    expectedBodyAttrs,
		Count:   occurences,
	}

	result, err := postAssertions(url, assertion)
	if err != nil {
		log.Fatalf("failed to post assertions: %v", err)
	}

	p := quietAwarePrintfGenerator(*quiet)
	showAssertions(p, result)

	if !result.Success {
		os.Exit(1)
	}
}

func requestURL(baseURL string, port int) string {
	return fmt.Sprintf("%s:%d", baseURL, port)
}

func postAssertions(url string, assertion apiHTTP.AssertionRequest) (result assertionResult, err error) {
	d, _ := json.Marshal(assertion)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", url, postAssertionsURI), bytes.NewBuffer(d))
	if err != nil {
		return result, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-FakeAPI-Control", "meta")

	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	var assertionResponse apiHTTP.AssertionResponse
	if err := json.NewDecoder(resp.Body).Decode(&assertionResponse); err != nil {
		return result, fmt.Errorf("failed to decode JSON: %v", err)
	}

	result = assertionResult{
		Title:    assertionResponse.Title,
		Messages: assertionResponse.Messages,
	}
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		result.Success = true
	}

	return result, nil
}

func headersParsed(headers []string) map[string]string {
	expectedHeaders := make(map[string]string)
	for _, h := range headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			key := strings.ToLower(strings.TrimSpace(parts[0]))
			val := strings.TrimSpace(parts[1])
			expectedHeaders[key] = val
		}
	}
	return expectedHeaders
}

func attributesParsed(attributes []string) map[string]string {
	expectedAttrs := make(map[string]string)
	for _, a := range attributes {
		parts := strings.SplitN(a, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			expectedAttrs[key] = val
		}
	}
	return expectedAttrs
}

func showAssertions(printf func(string, ...any), result assertionResult) {
	if result.Success {
		printf("✅ %s\n", result.Title)
		return
	}

	if len(result.Messages) == 0 {
		printf("❌ %s\n", result.Title)
		return
	}

	for _, msg := range result.Messages {
		printf("❌ %s\n", msg)
	}
}

func quietAwarePrintfGenerator(quiet bool) func(s string, args ...any) {
	return func(s string, args ...any) {
		if !quiet {
			fmt.Printf(s, args...)
		}
	}
}
