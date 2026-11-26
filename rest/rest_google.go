package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GmailSendEndpoint is the Gmail API endpoint for sending messages.
var GmailSendEndpoint = "https://gmail.googleapis.com/gmail/v1/users/me/messages/send"

// google config

var googleOAuth2Config = &oauth2.Config{
	RedirectURL: "http://localhost:5000/",
	Scopes: []string{
		"https://www.googleapis.com/auth/gmail.send",
	},
	Endpoint: google.Endpoint,
}

func (p *GoogleProvider) SendMessage(messageReader io.Reader, args []string) error {
	return p.SendMessageOpt(messageReader, args, false)
}

func (p *GoogleProvider) SendMessageOpt(messageReader io.Reader, args []string, parseHeaders bool) error {
	if encodedMessage, err := encodeMessageOpt(messageReader, args, parseHeaders); err != nil {
		return err
	} else if _, err := p.sendMessageRest(encodedMessage); err != nil {
		return err
	}
	return nil
}

func NewProviderGoogle(provider, sender string, storage ConfigStorage) (IProvider, error) {
	st := &SavedToken{config: googleOAuth2Config, provider: provider, id: sender, Storage: storage}
	if err := st.Open(); err != nil {
		return nil, err
	}
	// oauth2.NewClient respects oauth2.HTTPClient in context (used in tests) and wraps st as TokenSource.
	p := &GoogleProvider{
		config: googleOAuth2Config,
		client: oauth2.NewClient(storage.Context(), st),
	}
	return p, nil
}

// GoogleAPIError represents an error returned by the Gmail REST API and implements ExitCoder and StatusCode.
type GoogleAPIError struct {
	HTTPStatusCode int
	Message        string
}

func (e *GoogleAPIError) StatusCode() int {
	return e.HTTPStatusCode
}

func (e *GoogleAPIError) ExitCode() int {
	return exitCodeForHTTPStatus(e.HTTPStatusCode)
}

func (e *GoogleAPIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("googleapi: Error %d: %s", e.HTTPStatusCode, e.Message)
	}
	return fmt.Sprintf("googleapi: Error %d", e.HTTPStatusCode)
}

func parseGoogleAPIError(res *http.Response) *GoogleAPIError {
	gErr := &GoogleAPIError{HTTPStatusCode: res.StatusCode}
	var parsed struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(res.Body).Decode(&parsed); err == nil {
		if parsed.Error.Code != 0 {
			gErr.HTTPStatusCode = parsed.Error.Code
		}
		gErr.Message = parsed.Error.Message
	}
	return gErr
}

func (p *GoogleProvider) sendMessageRest(bodyBuf *bytes.Buffer) (*http.Response, error) {
	var payload bytes.Buffer
	if err := json.NewEncoder(&payload).Encode(struct {
		Raw string `json:"raw"`
	}{
		Raw: bodyBuf.String(),
	}); err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, GmailSendEndpoint, &payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := p.client.Do(req)
	if err != nil {
		return &http.Response{}, err
	}
	defer res.Body.Close()
	if res.StatusCode > 299 {
		return &http.Response{}, parseGoogleAPIError(res)
	}
	return res, nil
}
