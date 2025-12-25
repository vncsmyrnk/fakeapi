package route

import (
	"encoding/json"
	"errors"
	"fakeapi/internal/customregexp"
	"fakeapi/internal/customtemplate"
	"fmt"
	"io"
	"os"
)

// EndpointContent represents what an endpoint must return.
type EndpointContent any

// EndpointPossibleContent represents all content data an endpoint can return,
// which can be set on a JSON data file. It can be dynamically filtered by
// named regexp groups
type EndpointPossibleContent struct {
	Paths   []string        `json:"paths"`
	Methods []string        `json:"methods"`
	Data    EndpointContent `json:"data"`
}

func (e EndpointPossibleContent) Matches(endpoint Endpoint) bool {
	var matchedPath bool
	for _, path := range e.Paths {
		if endpoint.InputPath == path {
			matchedPath = true
			break
		}
	}

	if !matchedPath {
		return false
	}

	for _, method := range e.Methods {
		if endpoint.InputMethod == method {
			return true
		}
	}
	return false
}

// Endpoint represents a custom endpoint set by the user.
type Endpoint struct {
	InputPath          string `json:"path"`
	InputMethod        string `json:"method"`
	OutputStatus       int    `json:"status"`
	OutputDelaySeconds int    `json:"delay_in_seconds"`
	// OutputContent overrides filterable data content
	OutputContent EndpointContent `json:"content"`
}

func (e Endpoint) Content(request Request, dataFilePath string) (EndpointContent, error) {
	c, err := e.content(request, dataFilePath)
	if err != nil {
		return nil, err
	}

	return e.applyTemplateWithRequestBody(c, request)
}

func (e Endpoint) content(request Request, dataFilePath string) (EndpointContent, error) {
	if e.OutputContent != nil {
		return e.OutputContent, nil
	}

	endpointContent, err := e.getPossibleContent(dataFilePath)
	if err != nil && !errors.Is(err, ErrEndpointPossibleContentFileNotFound) {
		return nil, err
	}

	if endpointContent == nil {
		return nil, nil
	}

	endpointFilteredContent := e.filterEnpointContentByRegexpNames(request.Path, *endpointContent)
	if endpointFilteredContent == nil {
		return nil, ErrEndpointFilteredPossibleContentNotFound
	}

	return endpointFilteredContent, nil
}

func (e Endpoint) applyTemplateWithRequestBody(
	content EndpointContent, request Request,
) (EndpointContent, error) {
	if len(request.Body) == 0 {
		return content, nil
	}

	templateVariables, err := e.requestBodyToTemplateVariables(request)
	if err != nil {
		return nil, fmt.Errorf("failed to build template variables: %w", err)
	}

	return customtemplate.Execute(content, templateVariables)
}

func (e Endpoint) requestBodyToTemplateVariables(request Request) (map[string]any, error) {
	var templateVariables map[string]any
	if err := json.Unmarshal(request.Body, &templateVariables); err != nil {
		return nil, fmt.Errorf("failed to build template variables: %w", err)
	}

	return templateVariables, nil
}

func (e Endpoint) filterEnpointContentByRegexpNames(
	requestPath string, endpointContent EndpointPossibleContent,
) EndpointContent {
	return customregexp.FilterBySubexpNames(e.InputPath, requestPath, endpointContent.Data)
}

func (e Endpoint) getPossibleContent(
	possibleContentFilePath string,
) (*EndpointPossibleContent, error) {
	file, err := os.Open(possibleContentFilePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrEndpointPossibleContentFileNotFound, err)
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrEndpointPossibleContentFileNotFound, err)
	}

	var endpointsPossibleContent []EndpointPossibleContent
	err = json.Unmarshal(byteValue, &endpointsPossibleContent)
	if err != nil {
		return nil, err
	}

	for _, endpointPossibleContent := range endpointsPossibleContent {
		if endpointPossibleContent.Matches(e) {
			return &endpointPossibleContent, nil
		}
	}

	return nil, nil
}

func (e Endpoint) String() string {
	return fmt.Sprintf("%s %s", e.InputMethod, e.InputPath)
}

// NewEndpointsFromFile creates endpoints read from a file.
func NewEndpointsFromFile(filePath string) ([]Endpoint, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var endpoints []Endpoint
	err = json.Unmarshal(byteValue, &endpoints)
	if err != nil {
		return nil, err
	}

	return endpoints, nil
}
