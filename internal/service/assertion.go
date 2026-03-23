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

	if !result.Success || len(result.RequestIDs) == 0 {
		return result, nil
	}

	err = s.repo.Create(ctx, result.RequestIDs)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (s assertionService) asserter(
	ctx context.Context, assertion domain.Assertion,
) (func(domain.Assertion) (domain.AssertionResult, error), error) {
	onlyCountMode := assertion.EmptyMethodAndURI()
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
		requests, err := s.requestRepo.FetchPendingByMethodAndURI(ctx, assertion.Method, assertion.URI)
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
	matchedRequestIDs := make([]int, 0, len(requests))
	for _, r := range requests {
		failedHeaderAssertions :=
			s.assertHeaders(r.RequestHeaders, assertion.Headers)

		failedBodyAssertions :=
			s.assertPayload(r.RequestBody, assertion.Body)

		failedAssertionsMessages = append(failedAssertionsMessages,
			failedBodyAssertions...)
		failedAssertionsMessages = append(failedAssertionsMessages,
			failedHeaderAssertions...)

		if len(failedBodyAssertions) == 0 && len(failedHeaderAssertions) == 0 {
			matchedRequestIDs = append(matchedRequestIDs, r.ID)
		}
	}

	failedAssertionsMessages = lo.Uniq(failedAssertionsMessages)
	matchedRequestCount := len(matchedRequestIDs)

	if matchedRequestCount != assertion.Count {
		if matchedRequestCount > 0 {
			return domain.NewAssertionResultCountMismatchError(assertion.Count, matchedRequestCount)
		}
		return domain.NewAssertionResultFailedError(failedAssertionsMessages)
	}

	return domain.NewAssertionResultAllSucceededOK(matchedRequestIDs)
}

func (s *assertionService) assertOnlyRequestCount(
	assertion domain.Assertion, requests []domain.Request,
) domain.AssertionResult {
	c := len(requests)
	if c != assertion.Count {
		return domain.NewAssertionResultCountMismatchError(assertion.Count, c)
	}
	if c == 0 {
		return domain.NewAssertionResultNoPendingAssertionsOK()
	}

	assertedRequestsIDs := lo.Map(requests,
		func(item domain.Request, _ int) int {
			return item.ID
		})
	return domain.NewAssertionResultAllPendingAssertedOK(assertedRequestsIDs)
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
