// nolint: errcheck
package rest

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"

	//"net/http/httputil"
	"strings"
	"testing"

	"golang.org/x/oauth2"
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

var configExample = `{"web":{"client_id":"client-xyz",
"project_id":"","auth_uri":"","token_uri":"","auth_provider_x509_cert_url":"",
"client_secret":"secretexyz","redirect_uris":null,"javascript_origins":null,"scopes":null}}
`

// DiscardCloser wraps an io.Writer and provides a no-op Close method
type DiscardCloser struct {
	io.Writer
}

// Close implements the io.Closer interface with a no-op implementation.
func (dc DiscardCloser) Close() error {
	return nil
}

type ConfigStorageTest struct {
	ConfigStorageBase
	testContent string
}

func (ct ConfigStorageTest) GetReader() (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(ct.testContent)), nil
}

func (ct ConfigStorageTest) GetWriter() (io.WriteCloser, error) {
	return DiscardCloser{io.Discard}, nil
}

func init() {
	tsMicrosoft = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.RequestURI() {
		case "/v1.0/me/sendMail":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{}"))
			//w.WriteHeader(http.StatusForbidden)
		case "/consumers/oauth2/v2.0/token":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(token))
		}
		//fmt.Printf("path: %s\n", r.URL.RequestURI())
		w.Write([]byte("Mock response Microsoft"))
	}))
	tsGoogle = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.RequestURI() {
		case "/gmail/v1/users/me/messages/send":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{}"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}))
}

func setupTest() *http.Client {
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
			case "gmail.googleapis.com:443":
				fallthrough
			case "api.google.com:443":
				ts = tsGoogle
			default:
				panic(fmt.Sprintf("server %s not supported by mock", addr))
			}
			return (&net.Dialer{}).DialContext(ctx, network, ts.Listener.Addr().String())
		},
	}
	http.DefaultClient.Transport = transport

	return &http.Client{
		Transport: transport,
	}
}

func TestSendMailMicrosoft(t *testing.T) {
	httpClient := setupTest()
	ctx := context.Background()
	ctxWithClient := context.WithValue(ctx, oauth2.HTTPClient, httpClient)
	storage := ConfigStorageTest{testContent: token}
	storage.Ctx = ctxWithClient
	if oauthConfig, err := OpenConfig(storage, "outlook"); err != nil {
		t.Log(err)
		t.Fail()
	} else if p, err := NewProviderOutlook(oauthConfig, testEmail, storage); err != nil {
		t.Log("error accessing token")
		t.Log(err)
		t.Fail()
	} else {
		if err := p.SendMessage(strings.NewReader(message), []string{testEmail}); err != nil {
			t.Logf("send failure: %s", err)
			t.Fail()
		}
	}
}

func TestSendMailGoogle(t *testing.T) {
	httpClient := setupTest()
	ctx := context.Background()
	ctxWithClient := context.WithValue(ctx, oauth2.HTTPClient, httpClient)
	storage := ConfigStorageTest{testContent: token}
	storage.Ctx = ctxWithClient
	if _, err := OpenConfig(storage, "gmail"); err != nil {
		t.Log(err)
		t.Fail()
	} else if p, err := NewProviderGoogle("gmail", testEmail, storage); err != nil {
		t.Log("error accessing token")
		t.Log(err)
		t.Fail()
	} else {
		if err := p.SendMessage(strings.NewReader(message), []string{testEmail}); err != nil {
			t.Logf("send failure: %s", err)
			t.Fail()
		}
	}
}
