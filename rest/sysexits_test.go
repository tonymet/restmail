package rest

import (
	"errors"
	"net"
	"testing"
	"time"

	"google.golang.org/api/googleapi"
)

type dummyTimeoutErr struct{}

func (e dummyTimeoutErr) Error() string   { return "i/o timeout" }
func (e dummyTimeoutErr) Timeout() bool   { return true }
func (e dummyTimeoutErr) Temporary() bool { return true }

func TestClassifyDeliveryError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "nil error is ExitOk",
			err:      nil,
			expected: ExitOk,
		},
		{
			name:     "data error is ExitDataErr",
			err:      &DataError{Err: errors.New("malformed headers")},
			expected: ExitDataErr,
		},
		{
			name:     "outlook 429 rate limit is ExitTempFail",
			err:      &HTTPStatusError{Code: 429},
			expected: ExitTempFail,
		},
		{
			name:     "outlook 500 server error is ExitTempFail",
			err:      &HTTPStatusError{Code: 500},
			expected: ExitTempFail,
		},
		{
			name:     "outlook 503 service unavailable is ExitTempFail",
			err:      &HTTPStatusError{Code: 503},
			expected: ExitTempFail,
		},
		{
			name:     "outlook 400 bad request is ExitUnavailable",
			err:      &HTTPStatusError{Code: 400},
			expected: ExitUnavailable,
		},
		{
			name:     "outlook 404 not found is ExitUnavailable",
			err:      &HTTPStatusError{Code: 404},
			expected: ExitUnavailable,
		},
		{
			name:     "graph error 429 is ExitTempFail",
			err:      &GraphError{HTTPStatusCode: 429, Code: "ApplicationThrottled"},
			expected: ExitTempFail,
		},
		{
			name:     "graph error 503 is ExitTempFail",
			err:      &GraphError{HTTPStatusCode: 503, Code: "ServiceUnavailable"},
			expected: ExitTempFail,
		},
		{
			name:     "graph error 400 is ExitUnavailable",
			err:      &GraphError{HTTPStatusCode: 400, Code: "BadRequest"},
			expected: ExitUnavailable,
		},
		{
			name:     "graph error 404 is ExitUnavailable",
			err:      &GraphError{HTTPStatusCode: 404, Code: "ResourceNotFound"},
			expected: ExitUnavailable,
		},
		{
			name:     "googleapi 500 error is ExitTempFail",
			err:      &GoogleAPIError{Err: &googleapi.Error{Code: 500, Message: "Internal Server Error"}},
			expected: ExitTempFail,
		},
		{
			name:     "googleapi 429 rate limit is ExitTempFail",
			err:      &GoogleAPIError{Err: &googleapi.Error{Code: 429, Message: "Quota Exceeded"}},
			expected: ExitTempFail,
		},
		{
			name:     "googleapi 400 error is ExitUnavailable",
			err:      &GoogleAPIError{Err: &googleapi.Error{Code: 400, Message: "Invalid Argument"}},
			expected: ExitUnavailable,
		},
		{
			name:     "net timeout is ExitTempFail",
			err:      &net.OpError{Op: "dial", Net: "tcp", Err: dummyTimeoutErr{}},
			expected: ExitTempFail,
		},
		{
			name:     "unknown generic error defaults to ExitTempFail",
			err:      errors.New("unexpected socket issue"),
			expected: ExitTempFail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyDeliveryError(tt.err)
			if got != tt.expected {
				t.Errorf("ClassifyDeliveryError(%v) = %d; want %d", tt.err, got, tt.expected)
			}
		})
	}
}

// verify timeout dummy implements net.Error
var _ net.Error = dummyTimeoutErr{}
var _ = time.Second
