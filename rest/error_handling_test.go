package rest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

type mockTransport struct {
	roundTrip func(*http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTrip(req)
}

func TestSendMessageOutlookErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		err        error
		wantErr    string
	}{
		{
			name:       "401 Unauthorized",
			statusCode: http.StatusUnauthorized,
			wantErr:    "statusCode = 401",
		},
		{
			name:       "429 Too Many Requests",
			statusCode: http.StatusTooManyRequests,
			wantErr:    "statusCode = 429",
		},
		{
			name:       "500 Internal Server Error",
			statusCode: http.StatusInternalServerError,
			wantErr:    "statusCode = 500",
		},
		{
			name:    "Network Error",
			err:     fmt.Errorf("network down"),
			wantErr: "network down",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &OutlookProvider{
				client: &http.Client{
					Transport: &mockTransport{
						roundTrip: func(r *http.Request) (*http.Response, error) {
							if tt.err != nil {
								return nil, tt.err
							}
							return &http.Response{
								StatusCode: tt.statusCode,
								Body:       io.NopCloser(strings.NewReader("{}")),
							}, nil
						},
					},
				},
			}

			err := p.SendMessage(strings.NewReader("test"), []string{"to:test@example.com"})
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("got error %q, want it to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestSendMessageGoogleErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    string
	}{
		{
			name:       "403 Forbidden",
			statusCode: http.StatusForbidden,
			wantErr:    "403",
		},
		{
			name:       "503 Service Unavailable",
			statusCode: http.StatusServiceUnavailable,
			wantErr:    "503",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &http.Client{
				Transport: &mockTransport{
					roundTrip: func(r *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: tt.statusCode,
							// Google client expects a JSON error response or it might fail differently
							Body:   io.NopCloser(strings.NewReader(`{"error": {"code": ` + fmt.Sprint(tt.statusCode) + `, "message": "mock error"}}`)),
							Header: make(http.Header),
						}, nil
					},
				},
			}

			ctx := context.WithValue(context.Background(), oauth2.HTTPClient, mockClient)
			storage := &ConfigStorageTest{
				testContent: token,
			}
			storage.Ctx = ctx

			// NewProviderGoogle will use the mockClient from context
			p, err := NewProviderGoogle("gmail", testEmail, storage)
			if err != nil {
				t.Fatalf("failed to create provider: %v", err)
			}

			err = p.SendMessage(strings.NewReader("test"), []string{"to:test@example.com"})
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("got error %q, want it to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}
