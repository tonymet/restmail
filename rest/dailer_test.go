// nolint: errcheck
package rest

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"

	//"net/http/httputil"
	"strings"
	"testing"
)

var (
	tsMicrosoft, tsGoogle *httptest.Server
	testEmail             = os.Getenv("TEST_EMAIL")
)

var message = `subject: test subject

test message`

var token = `{
    "access_token": "xxxxx",
    "expires_in": 3599,
    "ext_expires_in": 3599,
    "refresh_token": "yyyy",
    "scope": "Mail.Send",
    "token_type": "Bearer"
}
`

func init() {
	tsMicrosoft = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.RequestURI() {
		case "/v1.0/me/sendMail":
			w.WriteHeader(http.StatusOK)
			//w.WriteHeader(http.StatusForbidden)
		case "/consumers/oauth2/v2.0/token":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(token))
		}
		fmt.Printf("path: %s\n", r.URL.RequestURI())
		w.Write([]byte("Mock response Microsoft"))
	}))
	tsGoogle = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Mock response Google"))
	}))
}

func setupTest() {
	// Create transport that redirects all connections to test server's listener
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, network, tsMicrosoft.Listener.Addr().String())
		},
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Redirect all connections to test server's listener
			// Redirect all connections to test server's listener
			var ts *httptest.Server
			switch addr {
			case "login.microsoftonline.com:443":
				fallthrough
			case "graph.microsoft.com:443":
				ts = tsMicrosoft
			case "api.google.com:443":
				ts = tsGoogle
			default:
				panic(fmt.Sprintf("server %s not supported by mock", addr))
			}
			return (&net.Dialer{}).DialContext(ctx, network, ts.Listener.Addr().String())
		},
	}
	http.DefaultClient.Transport = transport
}

func TestSendMailMicrosoft(t *testing.T) {
	t.Skip()
	setupTest()
	storage := &ConfigStorageOS{ConfigStorageBase: ConfigStorageBase{Provider: "outlook"}}
	if oauthConfig, err := OpenConfig(storage, "outlook"); err != nil {
		panic(err)
	} else if p, err := NewProviderOutlook(oauthConfig, testEmail, storage); err != nil {
		t.Log("error accessing token")
		t.Log(err)
		t.Fail()
	} else {
		if err := p.SendMessage(strings.NewReader(message), []string{testEmail}); err != nil {
			t.Log("send failure")
			t.Fail()
		}
	}
}
