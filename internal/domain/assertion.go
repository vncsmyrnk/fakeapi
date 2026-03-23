package domain

import "fmt"

type Assertion struct {
	URI     string
	Method  string
	Count   int
	Body    map[string]string
	Headers map[string]string
}

func (a *Assertion) EmptyMethodAndURI() bool {
	return a.URI == "" || a.Method == ""
}

type AssertionResult struct {
	Success    bool
	Title      string
	Messages   []string
	RequestIDs []int
}

func NewAssertionResultNoPendingAssertionsOK() AssertionResult {
	return AssertionResult{
		Title:   "There are no pending assertions",
		Success: true,
	}
}

func NewAssertionResultAllSucceededOK(assertedRequestIDs []int) AssertionResult {
	return AssertionResult{
		Title:      "All assertions succeeded!",
		Success:    true,
		RequestIDs: assertedRequestIDs,
	}
}

func NewAssertionResultNoPendingAssertionsError() AssertionResult {
	return AssertionResult{
		Title: "There are no pending assertions",
	}
}

func NewAssertionResultNoMatchingRequestsError() AssertionResult {
	return AssertionResult{
		Title: "The assertion did not match any request",
	}
}

func NewAssertionResultFailedError(assertionsMessages []string) AssertionResult {
	return AssertionResult{
		Title:    "Assertions failed",
		Messages: assertionsMessages,
	}
}

func NewAssertionResultCountMismatchError(expected, got int) AssertionResult {
	return AssertionResult{
		Title: fmt.Sprintf("Occurence count mismatch: expected %d, got %d", expected, got),
	}
}
