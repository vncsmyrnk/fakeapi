package http

import (
	"bytes"
	"encoding/json"
	"fmt"

	"fakeapi/internal/domain"
)

type endpointResponse struct {
	ID         int    `json:"id"`
	Path       string `json:"path"`
	Method     string `json:"method"`
	StatusCode int    `json:"statusCode"`
	Response   any    `json:"response"`
}

type EndpointRequest struct {
	Path       string          `json:"path"`
	Method     string          `json:"method"`
	StatusCode int             `json:"statusCode"`
	Response   json.RawMessage `json:"response"`
}

func newEndpointResponse(e domain.Endpoint) endpointResponse {
	var c any
	if e.Response != nil {
		_ = json.Unmarshal([]byte(*e.Response), &c)
	}
	return endpointResponse{
		ID:         e.ID,
		Path:       e.Path,
		Method:     e.Method,
		StatusCode: e.StatusCode,
		Response:   c,
	}
}

func NewDomainEndpointFromRequest(er EndpointRequest) domain.Endpoint {
	var (
		c *string
		b bytes.Buffer
	)
	_ = json.Compact(&b, er.Response)

	s := b.String()
	if s == "" || s == "null" {
		c = nil
	} else {
		c = &s
	}

	return domain.Endpoint{
		Path:       er.Path,
		Method:     er.Method,
		StatusCode: er.StatusCode,
		Response:   c,
	}
}

type RequestResponse struct {
	ID              int     `json:"id"`
	EndpointID      int     `json:"endpointId"`
	AssertionStatus string  `json:"assertion_status"`
	Path            string  `json:"path"`
	Method          string  `json:"method"`
	HitTime         string  `json:"hitTime"`
	UserAgent       *string `json:"userAgent,omitempty"`
	RequestHeaders  any     `json:"requestHeaders,omitempty"`
	RequestBody     any     `json:"requestBody,omitempty"`
}

func newRequestResponse(r domain.Request) RequestResponse {
	var headers any
	if r.RequestHeaders != nil {
		_ = json.Unmarshal([]byte(*r.RequestHeaders), &headers)
	}

	var body any
	if r.RequestBody != nil {
		err := json.Unmarshal([]byte(*r.RequestBody), &body)
		if err != nil {
			// If not a JSON string, just return it as a string
			body = *r.RequestBody
		}
	}

	return RequestResponse{
		ID:              r.ID,
		EndpointID:      r.EndpointID,
		AssertionStatus: string(r.AssertionStatus),
		Path:            r.Path,
		Method:          r.Method,
		HitTime:         r.HitTime.Format("2006-01-02T15:04:05Z07:00"),
		UserAgent:       r.UserAgent,
		RequestHeaders:  headers,
		RequestBody:     body,
	}
}

type AssertionRequest struct {
	Path    string            `json:"path"`
	Method  string            `json:"method"`
	Count   *int              `json:"count"`
	Body    map[string]string `json:"body"`
	Headers map[string]string `json:"headers"`
}

func newDomainAssertionFromAssertionRequest(ar AssertionRequest) (assertion domain.Assertion, err error) {
	if ar.Count == nil && (ar.Path == "" || ar.Method == "") {
		return assertion, fmt.Errorf("count or method/path are required")
	}

	count := 1
	if ar.Count != nil {
		count = *ar.Count
	}
	return domain.Assertion{
		Method:  ar.Method,
		Path:    ar.Path,
		Body:    ar.Body,
		Headers: ar.Headers,
		Count:   count,
	}, nil
}

type AssertionResponse struct {
	Title    string   `json:"title"`
	Messages []string `json:"messages"`
}

func newAssertionResponse(ar domain.AssertionResult) AssertionResponse {
	return AssertionResponse{
		Title:    ar.Title,
		Messages: ar.Messages,
	}
}
