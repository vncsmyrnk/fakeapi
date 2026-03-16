package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"github.com/tidwall/gjson"

	apiHTTP "fakeapi/internal/handler/http"
)

const (
	serverBaseURL     = "http://localhost"
	getRequestsURI    = "/requests?pending=true"
	postAssertionsURI = "/assertions"
)

var client = &http.Client{
	Timeout: 10 * time.Second,
}

var (
	CliVersion = "dev"
)

type assertionExpectedValues struct {
	Method         string
	URI            string
	Headers        map[string]string
	BodyAttributes map[string]string
	Occurrences    int
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

	if *occurences == -1 {
		occurences = &[]int{1}[0]
	}

	url := requestURL(serverBaseURL, *port)
	requests, err := getRequests(url)
	if err != nil {
		log.Fatalf("failed to fetch requests: %v", err)
	}

	expectedHeaders := headersParsed(headers)
	expectedBodyAttrs := attributesParsed(attrs)

	var method, uri string
	args := flag.Args()
	if len(args) >= 2 {
		method = args[0]
		uri = args[1]
	}

	assertion := assertionExpectedValues{
		Method:         method,
		URI:            uri,
		Headers:        expectedHeaders,
		BodyAttributes: expectedBodyAttrs,
		Occurrences:    *occurences,
	}

	var (
		succeededAssertions, failedAssertions []string
		matchedRequestIDs                     []int
	)
	for _, r := range requests {
		match, s, f := assertRequest(r, assertion)
		if match {
			succeededAssertions = append(succeededAssertions, s...)
			failedAssertions = append(failedAssertions, f...)
			if len(f) == 0 {
				matchedRequestIDs = append(matchedRequestIDs, r.ID)
			}
		}
	}

	matchedRequestIDsCount := len(matchedRequestIDs)
	if matchedRequestIDsCount >= 1 {
		failedAssertions = []string{}
	}

	actualRequestCountMatchesAssertion := matchedRequestIDsCount == assertion.Occurrences
	noPendingRequests := matchedRequestIDsCount == 0 && len(succeededAssertions) == 0
	if matchedRequestIDsCount == assertion.Occurrences || (noPendingRequests && assertion.Occurrences == 0) {
		actualRequestCountMatchesAssertion = true
	}

	if !actualRequestCountMatchesAssertion {
		failedAssertions = append(failedAssertions,
			fmt.Sprintf("❌ Occurence count mismatch: expected %d, got %d",
				assertion.Occurrences, len(matchedRequestIDs)))
		matchedRequestIDs = []int{}
	} else if noPendingRequests {
		m := fmt.Sprintf("✅ There are still %d pending assertions", matchedRequestIDsCount)
		if matchedRequestIDsCount == 0 {
			m = "✅ There are no pending assertions"
		}
		succeededAssertions = []string{m}
	}

	p := quietAwarePrintLnGenerator(*quiet)
	showAssertions(p, succeededAssertions, failedAssertions)

	err = postAssertedRequestIDs(url, matchedRequestIDs)
	if err != nil {
		log.Warnf("failed to assert request IDs: %v", err)
	}

	if len(failedAssertions) > 0 {
		os.Exit(1)
	}
}

func assertRequest(
	r apiHTTP.RequestResponse,
	a assertionExpectedValues,
) (match bool, succeededAssertions, failedAssertions []string) {
	emptyMethodAndURI := a.Method == "" && a.URI == ""
	if emptyMethodAndURI {
		return true, succeededAssertions, failedAssertions
	}

	requestAndExpectedMethodURIMatch := r.Method == a.Method &&
		r.URI == a.URI
	if !requestAndExpectedMethodURIMatch {
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

	if len(succeededAssertions) == 0 && len(failedAssertions) == 0 {
		succeededAssertions = append(succeededAssertions,
			fmt.Sprintf("✅ Method and URI matches (%s %s)", r.Method, r.URI))
	}

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

func postAssertedRequestIDs(url string, ids []int) error {
	if len(ids) == 0 {
		return nil
	}

	d, _ := json.Marshal(ids)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", url, postAssertionsURI), bytes.NewBuffer(d))
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-FakeAPI-Control", "meta")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return nil
	}

	return fmt.Errorf("failed to request assertions")
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

	s := lo.Uniq(succeededAssertions)
	f := lo.Uniq(failedAssertions)
	for _, assertionText := range s {
		printLn(assertionText)
	}

	if len(s) > 0 {
		printLn()
	}

	if len(f) > 0 {
		printLn("--- Assertion failures ---")
		for _, assertionText := range f {
			printLn(assertionText)
		}

		if len(s) == 0 {
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
