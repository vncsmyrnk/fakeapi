package http

import (
	"bytes"
	"encoding/json"

	"fakeapi/internal/domain"
)

type endpointResponse struct {
	ID         int    `json:"id"`
	URI        string `json:"uri"`
	Method     string `json:"method"`
	StatusCode int    `json:"statusCode"`
	Content    any    `json:"content"`
}

type EndpointRequest struct {
	ID         int             `json:"id"`
	URI        string          `json:"uri"`
	Method     string          `json:"method"`
	StatusCode int             `json:"statusCode"`
	Content    json.RawMessage `json:"content"`
}

func newEndpointResponse(e domain.Endpoint) endpointResponse {
	var c any
	if e.Content != nil {
		_ = json.Unmarshal([]byte(*e.Content), &c)
	}
	return endpointResponse{
		ID:         e.ID,
		URI:        e.URI,
		Method:     e.Method,
		StatusCode: e.StatusCode,
		Content:    c,
	}
}

func NewDomainEndpointFromRequest(er EndpointRequest) domain.Endpoint {
	var (
		c *string
		b bytes.Buffer
	)
	_ = json.Compact(&b, er.Content)

	s := b.String()
	if s == "" || s == "null" {
		c = nil
	} else {
		c = &s
	}

	return domain.Endpoint{
		ID:         er.ID,
		URI:        er.URI,
		Method:     er.Method,
		StatusCode: er.StatusCode,
		Content:    c,
	}
}

type RequestResponse struct {
	ID             int     `json:"id"`
	EndpointID     int     `json:"endpointId"`
	URI            string  `json:"uri"`
	Method         string  `json:"method"`
	HitTime        string  `json:"hitTime"`
	UserAgent      *string `json:"userAgent,omitempty"`
	RequestHeaders any     `json:"requestHeaders,omitempty"`
	RequestBody    any     `json:"requestBody,omitempty"`
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
		ID:             r.ID,
		EndpointID:     r.EndpointID,
		URI:            r.URI,
		Method:         r.Method,
		HitTime:        r.HitTime.Format("2006-01-02T15:04:05Z07:00"),
		UserAgent:      r.UserAgent,
		RequestHeaders: headers,
		RequestBody:    body,
	}
}
