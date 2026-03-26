package service

import (
	"context"
	"fmt"
	"strings"

	"fakeapi/internal/domain"
	"fakeapi/internal/port"

	"github.com/samber/lo"
	"github.com/tidwall/gjson"
)

var ErrNoPendingRequests = fmt.Errorf("no pending requests")

type assertionService struct {
	repo        port.AssertionRepository
	requestRepo port.RequestRepository
}

// Ensure AssertionService implements port.AssertionService
var _ port.AssertionService = (*assertionService)(nil)

// NewAssertionService creates a new instance of AssertionService.
func NewAssertionService(repo port.AssertionRepository, requestRepo port.RequestRepository) port.AssertionService {
	return &assertionService{
		repo:        repo,
		requestRepo: requestRepo,
	}
}

func (s *assertionService) Assert(
	ctx context.Context, a domain.Assertion,
) (result domain.AssertionResult, err error) {
	requestAsserter, err := s.asserter(ctx, a)
	if err != nil {
		return result, fmt.Errorf("failed to process assertions: %w", err)
	}

	result, err = requestAsserter(a)
	if err != nil {
		return result, fmt.Errorf("failed to assert requests: %w", err)
	}

	if len(result.RequestIDs) == 0 {
		return result, nil
	}

	err = s.createAssertionResults(ctx, result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (s assertionService) createAssertionResults(
	ctx context.Context, result domain.AssertionResult,
) error {
	if result.Success {
		return s.repo.CreateAsOK(ctx, result.RequestIDs)
	}
	return s.repo.CreateAsFailed(ctx, result.RequestIDs)
}

func (s assertionService) asserter(
	ctx context.Context, assertion domain.Assertion,
) (func(domain.Assertion) (domain.AssertionResult, error), error) {
	onlyCountMode := assertion.Empty()
	if onlyCountMode {
		return func(a domain.Assertion) (result domain.AssertionResult, err error) {
			requests, err := s.requestRepo.FetchPending(ctx)
			if err != nil {
				return result, fmt.Errorf("failed to fetch assertions: %w", err)
			}
			return s.assertOnlyRequestCount(a, requests), nil
		}, nil
	}

	return func(a domain.Assertion) (result domain.AssertionResult, err error) {
		requests, err := s.requestRepo.FetchPendingByMethodAndPath(ctx, assertion.Method, assertion.Path)
		if err != nil {
			return result, fmt.Errorf("failed to fetch assertions: %w", err)
		}
		return s.assert(a, requests), nil
	}, nil
}

func (s *assertionService) assert(
	assertion domain.Assertion, requests []domain.Request,
) domain.AssertionResult {
	if len(requests) == 0 && assertion.Count == 0 {
		return domain.NewAssertionResultNoPendingAssertionsOK()
	}

	if len(requests) == 0 {
		return domain.NewAssertionResultNoMatchingRequestsError()
	}

	failedAssertionsMessages := make([]string, 0,
		len(requests)*len(assertion.Headers)*len(assertion.Body))
	allRequestIDs := make([]int, 0, len(requests))
	matchedRequestIDs := make([]int, 0, len(requests))
	for _, r := range requests {
		allRequestIDs = append(allRequestIDs, r.ID)

		failedAssertionsMessagesForRequest := make([]string, 0,
			cap(failedAssertionsMessages)/2)

		failedAssertionsMessagesForRequest = append(failedAssertionsMessagesForRequest,
			s.assertHeaders(r.RequestHeaders, assertion.Headers)...)
		failedAssertionsMessagesForRequest = append(failedAssertionsMessagesForRequest,
			s.assertQueryStrings(r.RequestQueryStrings, assertion.QueryStrings)...)
		failedAssertionsMessagesForRequest = append(failedAssertionsMessagesForRequest,
			s.assertPayload(r.RequestBody, assertion.Body)...)

		if len(failedAssertionsMessagesForRequest) == 0 {
			matchedRequestIDs = append(matchedRequestIDs, r.ID)
		}
	}

	failedAssertionsMessages = lo.Uniq(failedAssertionsMessages)
	matchedRequestCount := len(matchedRequestIDs)

	if matchedRequestCount != assertion.Count {
		if matchedRequestCount > 0 {
			return domain.NewAssertionResultCountMismatchError(allRequestIDs, assertion.Count, matchedRequestCount)
		}
		return domain.NewAssertionResultFailedError(allRequestIDs, failedAssertionsMessages)
	}

	return domain.NewAssertionResultAllSucceededOK(matchedRequestIDs)
}

func (s *assertionService) assertOnlyRequestCount(
	assertion domain.Assertion, requests []domain.Request,
) domain.AssertionResult {
	c := len(requests)
	requestsIDs := lo.Map(requests,
		func(r domain.Request, _ int) int {
			return r.ID
		})
	if c != assertion.Count {
		return domain.NewAssertionResultCountMismatchError(requestsIDs, assertion.Count, c)
	}
	if c == 0 {
		return domain.NewAssertionResultNoPendingAssertionsOK()
	}
	return domain.NewAssertionResultAllPendingAssertedOK(requestsIDs)
}

func (s *assertionService) assertPayload(jsonPayload *string, expectedAttrs map[string]string) (failures []string) {
	if len(expectedAttrs) == 0 {
		return failures
	}

	if jsonPayload == nil || *jsonPayload == "" {
		if len(expectedAttrs) > 0 {
			failures = append(failures, "Recorded request has no body")
		}
		return failures
	}

	payload := *jsonPayload
	if !gjson.Valid(payload) {
		return append(failures, "Payload from database is not valid JSON")
	}

	for p, v := range expectedAttrs {
		result := gjson.Get(payload, p)

		if !result.Exists() {
			failures = append(failures, fmt.Sprintf("Attribute '%s' not found in payload", p))
			continue
		}

		actualVal := result.String()
		if actualVal != v {
			failures = append(failures, fmt.Sprintf("Mismatch for '%s': expected '%s', got '%s'", p, v, actualVal))
		}
	}

	return failures
}

func (s *assertionService) assertHeaders(jsonHeaders *string, expectedHeaders map[string]string) (failures []string) {
	if len(expectedHeaders) == 0 {
		return failures
	}

	if jsonHeaders == nil || *jsonHeaders == "" {
		if len(expectedHeaders) > 0 {
			failures = append(failures, "Recorded request has no headers")
		}
		return failures
	}

	headers := *jsonHeaders
	if !gjson.Valid(headers) {
		return append(failures, "Headers from database are not valid JSON")
	}

	parsedDBHeaders := gjson.Parse(headers).Map()
	normalizedDBHeaders := make(map[string]string)

	for k, v := range parsedDBHeaders {
		normalizedDBHeaders[strings.ToLower(k)] = v.String()
	}

	for k, v := range expectedHeaders {
		actualVal, exists := normalizedDBHeaders[k]

		if !exists {
			failures = append(failures, fmt.Sprintf("Header '%s' not found in request", k))
			continue
		}

		if actualVal != v {
			failures = append(failures,
				fmt.Sprintf("Mismatch for header '%s': expected '%s', got '%s'", k, v, actualVal))
		}
	}

	return failures
}

func (s *assertionService) assertQueryStrings(
	jsonQueryStrings *string, expectedQueryStrings map[string]string,
) (failures []string) {
	if len(expectedQueryStrings) == 0 {
		return failures
	}

	if jsonQueryStrings == nil || *jsonQueryStrings == "" {
		if len(expectedQueryStrings) > 0 {
			failures = append(failures, "Recorded request has no query strings")
		}
		return failures
	}

	queryStrings := *jsonQueryStrings
	if !gjson.Valid(queryStrings) {
		return append(failures, "Query string from database are not valid JSON")
	}

	for p, v := range expectedQueryStrings {
		result := gjson.Get(queryStrings, p)

		if !result.Exists() {
			failures = append(failures, fmt.Sprintf("Attribute '%s' not found in query strings", p))
			continue
		}

		actualVal := result.String()
		if actualVal != v {
			failures = append(failures, fmt.Sprintf("Mismatch for '%s': expected '%s', got '%s'", p, v, actualVal))
		}
	}

	return failures
}
