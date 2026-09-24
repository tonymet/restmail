package rest

import (
	"errors"
	"fmt"
	"net"
	"net/http"
)

// Standard sysexits.h status codes and OpenBSD smtpd MDA delivery conventions:
// - Status 0 (EX_OK): Successful delivery.
// - Status 71 (EX_OSERR) and 75 (EX_TEMPFAIL): Temporary failures (smtpd retries).
// - All other exit status: Permanent failures (smtpd bounces).
const (
	ExitOk          = 0  // EX_OK: Successful delivery
	ExitUsage       = 64 // EX_USAGE: Command line usage error
	ExitDataErr     = 65 // EX_DATAERR: Data format error (e.g. malformed email headers/body)
	ExitUnavailable = 69 // EX_UNAVAILABLE: Service unavailable or permanent delivery failure
	ExitOSErr       = 71 // EX_OSERR: Operating system error / local storage / token failure
	ExitTempFail    = 75 // EX_TEMPFAIL: Temporary failure; retry later
)

// ExitCoder is implemented by errors that provide their own MDA exit code.
type ExitCoder interface {
	ExitCode() int
}

// HTTPStatusError represents an error carrying an HTTP status code.
type HTTPStatusError struct {
	Code int
	Err  error
}

func (e *HTTPStatusError) StatusCode() int {
	return e.Code
}

func (e *HTTPStatusError) ExitCode() int {
	if e.Code == http.StatusTooManyRequests || e.Code >= 500 {
		return ExitTempFail
	}
	if e.Code >= 400 && e.Code < 500 {
		return ExitUnavailable
	}
	return ExitTempFail
}

func (e *HTTPStatusError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("HTTP error %d: %v", e.Code, e.Err)
	}
	return fmt.Sprintf("error sending mail: statusCode = %d", e.Code)
}

func (e *HTTPStatusError) Unwrap() error {
	return e.Err
}

// DataError represents an error with the input mail data format (e.g. malformed headers).
type DataError struct {
	Err error
}

func (e *DataError) ExitCode() int {
	return ExitDataErr
}

func (e *DataError) Error() string {
	return fmt.Sprintf("mail data error: %v", e.Err)
}

func (e *DataError) Unwrap() error {
	return e.Err
}

// ClassifyDeliveryError inspects an error encountered during mail delivery
// and returns the appropriate MDA exit code.
func ClassifyDeliveryError(err error) int {
	if err == nil {
		return ExitOk
	}

	// 1. Check if the error implements ExitCoder directly
	var exitCoder ExitCoder
	if errors.As(err, &exitCoder) {
		return exitCoder.ExitCode()
	}

	// 2. Check for any type exposing StatusCode() int (e.g., HTTP clients or API wrappers)
	type hasStatusCode interface {
		StatusCode() int
	}
	var scErr hasStatusCode
	if errors.As(err, &scErr) {
		code := scErr.StatusCode()
		if code == http.StatusTooManyRequests || code >= 500 {
			return ExitTempFail
		}
		if code >= 400 && code < 500 {
			return ExitUnavailable
		}
	}

	// 4. Check for network timeouts / temporary network errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return ExitTempFail
		}
	}

	// Default unknown errors to ExitTempFail so smtpd can retry transient glitches
	return ExitTempFail
}
