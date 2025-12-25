package route

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
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

func (e Endpoint) filterEnpointContentByRegexpNames(
	requestPath string, endpointContent EndpointPossibleContent,
) EndpointContent {
	contentList, ok := endpointContent.Data.([]any)
	if !ok {
		return nil
	}

	re := regexp.MustCompile(e.InputPath)
	match := re.FindStringSubmatch(requestPath)
	if match == nil {
		return nil
	}

	if re.NumSubexp() == 0 {
		return endpointContent
	}

	regexpGroupNames := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i == 0 {
			continue
		}
		regexpGroupNames[name] = match[i]
	}

	for _, content := range contentList {
		c, ok := content.(map[string]any)
		if !ok {
			break
		}

		var matches int
		for k, v := range regexpGroupNames {
			contentPropertyValue, ok := c[k]
			if !ok {
				continue
			}

			strContentPropertyValue := fmt.Sprintf("%v", contentPropertyValue)
			strPossibleValue := fmt.Sprintf("%v", v)
			if strContentPropertyValue == strPossibleValue {
				matches++
			}
		}

		if matches == len(regexpGroupNames) {
			return c
		}
	}

	return nil
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
