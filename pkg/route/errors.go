package route

import "fmt"

var (
	ErrEndpointFilteredPossibleContentNotFound = fmt.Errorf(
		"endpoint filtered possible content not found")
	ErrEndpointPossibleContentFileNotFound = fmt.Errorf(
		"endpoint possible content file not found")
)
