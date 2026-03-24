package domain

import "fmt"

type Assertion struct {
	Path    string
	Method  string
	Count   int
	Body    map[string]string
	Headers map[string]string
}

func (a *Assertion) Empty() bool {
	return a.Path == "" || a.Method == ""
}

type AssertionStatus string

const (
	AssertionStatusOK     AssertionStatus = "ok"
	AssertionStatusFailed AssertionStatus = "failed"
)

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

func NewAssertionResultAllPendingAssertedOK(assertedRequestsIDs []int) AssertionResult {
	return AssertionResult{
		Title:      fmt.Sprintf("There were %d pending assertions", len(assertedRequestsIDs)),
		Success:    true,
		RequestIDs: assertedRequestsIDs,
	}
}

func NewAssertionResultAllSucceededOK(assertedRequestsIDs []int) AssertionResult {
	return AssertionResult{
		Title:      "All assertions succeeded!",
		Success:    true,
		RequestIDs: assertedRequestsIDs,
	}
}

func NewAssertionResultNoPendingAssertionsError(assertedRequestsIDs []int) AssertionResult {
	return AssertionResult{
		Title:      "There are no pending assertions",
		RequestIDs: assertedRequestsIDs,
	}
}

func NewAssertionResultNoMatchingRequestsError() AssertionResult {
	return AssertionResult{
		Title: "The assertion did not match any request",
	}
}

func NewAssertionResultFailedError(assertedRequestsIDs []int, assertionsMessages []string) AssertionResult {
	return AssertionResult{
		Title:      "Assertions failed",
		Messages:   assertionsMessages,
		RequestIDs: assertedRequestsIDs,
	}
}

func NewAssertionResultCountMismatchError(assertedRequestsIDs []int, expected, got int) AssertionResult {
	return AssertionResult{
		Title:      fmt.Sprintf("Occurence count mismatch: expected %d, got %d", expected, got),
		RequestIDs: assertedRequestsIDs,
	}
}
