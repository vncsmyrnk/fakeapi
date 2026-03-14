package domain

import (
	"time"
)

// Request represents a recorded HTTP request that hit a mocked endpoint.
type Request struct {
	ID             int
	EndpointID     int
	URI            string
	Method         string
	HitTime        time.Time
	UserAgent      *string
	RequestHeaders *string
	RequestBody    *string
}
