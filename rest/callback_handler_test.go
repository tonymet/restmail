package rest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCallbackHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		state          string
		query          string
		wantStatus     int
		wantBody       string
		wantCodeInChan string
	}{
		{
			name:           "Success",
			state:          "test-state",
			query:          "code=test-code&state=test-state",
			wantStatus:     http.StatusOK,
			wantBody:       "code: test-code",
			wantCodeInChan: "test-code",
		},
		{
			name:       "State Mismatch",
			state:      "test-state",
			query:      "code=test-code&state=wrong-state",
			wantStatus: http.StatusForbidden,
			wantBody:   "State mismatch error",
		},
		{
			name:       "Missing State",
			state:      "test-state",
			query:      "code=test-code",
			wantStatus: http.StatusForbidden,
			wantBody:   "State mismatch error",
		},
		{
			name:       "OAuth Error",
			state:      "test-state",
			query:      "error=access_denied&state=test-state",
			wantStatus: http.StatusBadRequest,
			wantBody:   "OAuth Error: access_denied",
		},
		{
			name:       "Missing Code",
			state:      "test-state",
			query:      "state=test-state",
			wantStatus: http.StatusBadRequest,
			wantBody:   "expecting query param: code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &callbackHandler{
				codeChan: make(chan string, 1),
				state:    tt.state,
			}

			req := httptest.NewRequest("GET", "/?"+tt.query, nil)
			rr := httptest.NewRecorder()

			h.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("got status %v, want %v", rr.Code, tt.wantStatus)
			}

			if !strings.Contains(rr.Body.String(), tt.wantBody) {
				t.Errorf("got body %q, want it to contain %q", rr.Body.String(), tt.wantBody)
			}

			if tt.wantCodeInChan != "" {
				select {
				case code := <-h.codeChan:
					if code != tt.wantCodeInChan {
						t.Errorf("got code %q from channel, want %q", code, tt.wantCodeInChan)
					}
				default:
					t.Error("expected code in channel, but it was empty")
				}
			} else {
				if len(h.codeChan) > 0 {
					t.Errorf("expected channel to be empty, but got code %q", <-h.codeChan)
				}
			}
		})
	}
}
