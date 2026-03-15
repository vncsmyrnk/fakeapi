package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"github.com/tidwall/gjson"

	apiHTTP "fakeapi/internal/handler/http"
)

const (
	serverBaseURL  = "http://localhost"
	getRequestsURI = "/requests"
)

type assertionExpectedValues struct {
	Method         string
	URI            string
	Headers        map[string]string
	BodyAttributes map[string]string
}

func main() {
	port := flag.Int("port", 8080, "Port the target server is running on")
	quiet := flag.Bool("q", false, "Quiet mode")

	var (
		headers, attrs []string
	)
	flag.StringArrayVarP(&headers, "header", "H", []string{}, "Expected headers in 'Key: Value' format")
	flag.StringArrayVarP(&attrs, "body-attributes", "b", []string{}, "Expected body attributes in 'key=value' format")
	flag.Parse()

	url := requestURL(serverBaseURL, *port)
	requests, err := getRequests(url)
	if err != nil {
		log.Fatalf("failed to fetch requests: %v", err)
	}

	args := flag.Args()
	if len(args) < 2 {
		log.Fatal("missing arguments")
	}

	expectedHeaders := headersParsed(headers)
	expectedBodyAttrs := attributesParsed(attrs)

	method := args[0]
	uri := args[1]

	var (
		match                                 bool
		succeededAssertions, failedAssertions []string
	)
	for _, r := range requests {
		match, succeededAssertions, failedAssertions =
			assertRequest(r, assertionExpectedValues{
				Method:         method,
				URI:            uri,
				Headers:        expectedHeaders,
				BodyAttributes: expectedBodyAttrs,
			})

		if match {
			break
		}
	}

	p := quietAwarePrintLnGenerator(*quiet)
	showAssertions(p, succeededAssertions, failedAssertions)

	if len(failedAssertions) > 0 {
		os.Exit(1)
	}
}

func assertRequest(
	r apiHTTP.RequestResponse,
	a assertionExpectedValues,
) (match bool, succeededAssertions, failedAssertions []string) {
	if r.Method != a.Method || r.URI != a.URI {
		return
	}

	h, _ := json.Marshal(r.RequestHeaders)
	succeededHeaderAssertions, failedHeaderAssertions :=
		assertHeaders(string(h), a.Headers)

	b, _ := json.Marshal(r.RequestBody)
	succeededBodyAssertions, failedBodyAssertions :=
		assertPayload(string(b), a.BodyAttributes)

	succeededAssertions = append(succeededHeaderAssertions, succeededBodyAssertions...)
	failedAssertions = append(failedHeaderAssertions, failedBodyAssertions...)
	return true, succeededAssertions, failedAssertions
}

func assertPayload(jsonPayload string, expectedAttrs map[string]string) (successes []string, failures []string) {
	if len(expectedAttrs) == 0 {
		return successes, failures
	}

	if !gjson.Valid(jsonPayload) {
		return successes, append(failures, "❌ Payload from database is not valid JSON")
	}

	for p, v := range expectedAttrs {
		result := gjson.Get(jsonPayload, p)

		if !result.Exists() {
			failures = append(failures, fmt.Sprintf("❌ Attribute '%s' not found in payload", p))
			continue
		}

		actualVal := result.String()
		if actualVal != v {
			failures = append(failures, fmt.Sprintf("❌ Mismatch for '%s': expected '%s', got '%s'", p, v, actualVal))
		} else {
			successes = append(successes, fmt.Sprintf("✅ Attribute '%s' matches ('%s')", p, actualVal))
		}
	}

	return successes, failures
}

func assertHeaders(jsonHeaders string, expectedHeaders map[string]string) (successes []string, failures []string) {
	if len(expectedHeaders) == 0 {
		return successes, failures
	}

	if !gjson.Valid(jsonHeaders) {
		return successes, append(failures, "❌ Headers from database are not valid JSON")
	}

	parsedDBHeaders := gjson.Parse(jsonHeaders).Map()
	normalizedDBHeaders := make(map[string]string)

	for k, v := range parsedDBHeaders {
		normalizedDBHeaders[strings.ToLower(k)] = v.String()
	}

	for k, v := range expectedHeaders {
		actualVal, exists := normalizedDBHeaders[k]

		if !exists {
			failures = append(failures, fmt.Sprintf("❌ Header '%s' not found in request", k))
			continue
		}

		if actualVal != v {
			failures = append(failures,
				fmt.Sprintf("❌ Mismatch for header '%s': expected '%s', got '%s'", k, v, actualVal))
		} else {
			successes = append(successes,
				fmt.Sprintf("✅ Header '%s' matches ('%s')", k, actualVal))
		}
	}

	return successes, failures
}

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

func showAssertions(printLn func(args ...any), succeededAssertions, failedAssertions []string) {
	if len(succeededAssertions) == 0 && len(failedAssertions) == 0 {
		printLn("❌ The assertion did not match any request")
		return
	}

	for _, s := range succeededAssertions {
		printLn(s)
	}

	if len(succeededAssertions) > 0 {
		printLn()
	}

	if len(failedAssertions) > 0 {
		printLn("--- Assertion failures ---")
		for _, f := range failedAssertions {
			printLn(f)
		}

		if len(succeededAssertions) == 0 {
			printLn("\n❌ All assertions failed")
		}
		return
	}

	printLn("🎉 All assertions succeeded!")
}

func quietAwarePrintLnGenerator(quiet bool) func(args ...any) {
	return func(args ...any) {
		if !quiet {
			fmt.Println(args...)
		}
	}
}
