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
	"github.com/spf13/cobra"

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
	rootCmd := &cobra.Command{
		Version: CliVersion,
	}
	rootCmd.SetVersionTemplate("{{.Version}}\n")

	port := rootCmd.PersistentFlags().IntP("port", "p", 8080, "Port the target server is running on")
	quiet := rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Quiet mode")

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
		url := requestURL(serverBaseURL, *port)

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
	rootCmd.AddCommand(cmdAssert)

	_ = rootCmd.Execute()
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
