package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/Masterminds/semver/v3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	apiHTTP "fakeapi/internal/handler/http"
	"fakeapi/internal/version"
)

const (
	serverBaseURL        = "http://localhost"
	postAssertionsURI    = "/assertions"
	requestsURI          = "/requests"
	minimumServerVersion = ""
)

var client = &http.Client{
	Timeout: 10 * time.Second,
}

var serverVersionConstraint, _ = semver.NewConstraint("~0.7")

type assertionResult struct {
	Success  bool
	Title    string
	Messages []string
}

func main() {
	rootCmd := &cobra.Command{
		Use:     "fakeapi",
		Version: version.CliVersion,
	}
	rootCmd.SetVersionTemplate("{{.Version}}\n")

	port := rootCmd.PersistentFlags().IntP("port", "p", 8080, "Port the target server is running on")
	quiet := rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Quiet mode")

	cmdClear := &cobra.Command{
		Use:   "clear",
		Short: "Deletes fake API configuration or recorded data",
		Long:  "Deletes fake API configuration or recorded data",
		Args:  cobra.ExactArgs(1),
	}

	cmdClearRequests := &cobra.Command{
		Use:   "requests",
		Short: "Deletes all recorded requests",
		Long:  "Deletes all recorded requests",
		Args:  cobra.NoArgs,
	}
	cmdClear.AddCommand(cmdClearRequests)
	rootCmd.AddCommand(cmdClear)

	url := requestURL(serverBaseURL, *port)
	r := newRequester(url, serverResponseVersionCheck)

	cmdClearRequests.Run = func(_ *cobra.Command, _ []string) {
		err := deleteRequests(r)
		if err != nil {
			log.Fatalf("failed to delete requests: %v", err)
		}
	}

	cmdList := &cobra.Command{
		Use:   "list",
		Short: "Lists fake API configuration or recorded data",
		Long:  "Lists fake API configuration or recorded data",
		Args:  cobra.ExactArgs(1),
	}

	cmdListRequests := &cobra.Command{
		Use:   "requests",
		Short: "Lists all recorded requests",
		Long:  "Lists all recorded requests",
		Args:  cobra.NoArgs,
	}
	cmdList.AddCommand(cmdListRequests)
	rootCmd.AddCommand(cmdList)
	pendingRequests := cmdListRequests.Flags().Bool("pending", false, "Pending requests")
	jsonRequests := cmdListRequests.Flags().Bool("json", false, "Return a JSON response")

	cmdListRequests.Run = func(_ *cobra.Command, _ []string) {
		requests, err := getRequests(r, *pendingRequests)
		if err != nil {
			log.Fatalf("failed to fetch requests: %v", err)
		}

		if *jsonRequests {
			jsonRequests, err := json.Marshal(requests)
			if err != nil {
				log.Fatalf("failed to format JSON output: %v", err)
			}

			var prettyJSON bytes.Buffer
			err = json.Indent(&prettyJSON, []byte(jsonRequests), "", " ")
			if err != nil {
				log.Fatalf("failed to format JSON output: %v", err)
			}
			fmt.Println(prettyJSON.String())
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "METHOD\tURI\tASSERTED")
		for _, r := range requests {
			fmt.Fprintf(w, "%s\t%s\t%v\n", r.Method, r.URI, assertionStatus(r))
		}
		w.Flush()
	}

	cmdAssert := &cobra.Command{
		Use:   "assert [method] [URI]",
		Short: "Assert a request made to the fake API",
		Long:  "Assert can ensure a request was made to the server using headers and body filters.",
		Args:  cobra.ArbitraryArgs,
	}

	var (
		headers, attrs []string
	)
	cmdAssert.Flags().StringArrayVarP(&headers, "header", "H", []string{}, "Expected headers in 'Key: Value' format")
	cmdAssert.Flags().StringArrayVarP(&attrs, "body-attributes", "b", []string{}, "Expected body attributes in 'key=value' format")
	requestCount := cmdAssert.Flags().IntP("count", "c", -1, "Match occurrence count")

	cmdAssert.Run = func(cmd *cobra.Command, args []string) {
		expectedHeaders := headersParsed(headers)
		expectedBodyAttrs := attributesParsed(attrs)

		var method, uri string
		if len(args) >= 2 {
			method = args[0]
			uri = args[1]
		} else if *requestCount == -1 {
			_ = cmd.Usage()
			os.Exit(1)
		}

		if *requestCount == -1 {
			requestCount = nil
		}

		assertion := apiHTTP.AssertionRequest{
			Method:  method,
			URI:     uri,
			Headers: expectedHeaders,
			Body:    expectedBodyAttrs,
			Count:   requestCount,
		}

		result, err := postAssertions(r, assertion)
		if err != nil {
			log.Fatalf("failed to post assertions: %v", err)
		}

		p := quietAwarePrintfGenerator(*quiet)
		showAssertions(p, result)

		if !result.Success {
			os.Exit(1)
		}
	}
	rootCmd.AddCommand(cmdAssert)

	_ = rootCmd.Execute()
}

func serverResponseVersionCheck(r *http.Response) error {
	v := r.Header.Get("x-fakeapi-version")
	if v == "dev" {
		return nil
	}
	serverVersion, err := semver.NewVersion(v)
	if err != nil {
		return err
	}

	if constraintOK := serverVersionConstraint.Check(serverVersion); !constraintOK {
		return fmt.Errorf(
			"server version not compatible with the current CLI. The current CLI supports the following server versions: %s",
			serverVersionConstraint.String())
	}
	return nil
}

func requestURL(baseURL string, port int) string {
	return fmt.Sprintf("%s:%d", baseURL, port)
}

func postAssertions(r requester, assertion apiHTTP.AssertionRequest) (result assertionResult, err error) {
	var assertionResponse apiHTTP.AssertionResponse
	reqOptions := requestOptions{method: http.MethodPost, path: postAssertionsURI, body: assertion, ignoreResponseStatus: true}
	responseStatusCode, err := r(reqOptions, &assertionResponse)
	if err != nil {
		log.Fatalf("failed to post assertions: %v", err)
	}

	result = assertionResult{
		Title:    assertionResponse.Title,
		Messages: assertionResponse.Messages,
	}
	if responseStatusCode >= 200 && responseStatusCode <= 299 {
		result.Success = true
	}

	return result, nil
}

func getRequests(r requester, pending bool) (requests []apiHTTP.RequestResponse, err error) {
	reqOptions := requestOptions{method: http.MethodGet, path: fmt.Sprintf("%s?pending=%v", requestsURI, pending)}
	_, err = r(reqOptions, &requests)
	if err != nil {
		log.Fatalf("failed to fetch requests: %v", err)
	}

	return requests, nil
}

func deleteRequests(r requester) error {
	reqOptions := requestOptions{method: http.MethodDelete, path: requestsURI}
	_, err := r(reqOptions, nil)
	if err != nil {
		log.Fatalf("failed to delete requests: %v", err)
	}

	return nil
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

func assertionStatus(r apiHTTP.RequestResponse) string {
	var assertionStatusIcon string
	switch r.AssertionStatus {
	case "ok":
		assertionStatusIcon = "✅"
	case "failed":
		assertionStatusIcon = "❌"
	case "pending":
		assertionStatusIcon = "⏳"
	default:
		assertionStatusIcon = "❔"
	}
	assertionStatus := fmt.Sprintf("%s %s", assertionStatusIcon, cases.Title(language.AmericanEnglish).String(r.AssertionStatus))
	return assertionStatus
}

func quietAwarePrintfGenerator(quiet bool) func(s string, args ...any) {
	return func(s string, args ...any) {
		if !quiet {
			fmt.Printf(s, args...)
		}
	}
}

type requestOptions struct {
	method               string
	path                 string
	body                 any
	ignoreResponseStatus bool
}

type requester func(r requestOptions, responseBodyTarget any) (statusCode int, err error)

func newRequester(baseURL string, postRequest func(*http.Response) error) requester {
	return requester(func(r requestOptions, target any) (statusCode int, err error) {
		d, _ := json.Marshal(r.body)
		req, err := http.NewRequest(r.method, fmt.Sprintf("%s%s", baseURL, r.path), bytes.NewBuffer(d))
		if err != nil {
			return statusCode, err
		}

		req.Header.Set("Accept", "application/json")
		req.Header.Set("x-fakeapi-control", "meta")

		resp, err := client.Do(req)
		if err != nil {
			return statusCode, err
		}
		defer resp.Body.Close()

		statusCode = resp.StatusCode
		if !r.ignoreResponseStatus && resp.StatusCode >= 299 {
			return statusCode, fmt.Errorf("unexpected respose status code")
		}

		if target != nil {
			if err := json.NewDecoder(resp.Body).Decode(&target); err != nil {
				return statusCode, fmt.Errorf("failed to decode JSON: %v", err)
			}
		}

		return statusCode, postRequest(resp)
	})
}
