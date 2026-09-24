package rest

import (
	"errors"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

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
	var (
		p   = &GoogleProvider{config: googleOAuth2Config}
		err error
		st  = &SavedToken{config: googleOAuth2Config, provider: provider, id: sender, Storage: storage}
	)
	if err := st.Open(); err != nil {
		return nil, err
	}
	// use our tokensource and override httpClient for testing
	opts := make([]option.ClientOption, 0)
	opts = append(opts, option.WithTokenSource(st))
	if hc, ok := storage.Context().Value(oauth2.HTTPClient).(*http.Client); ok {
		opts = append(opts, option.WithHTTPClient(hc))
	}
	if p.srv, err = gmail.NewService(storage.Context(), opts...); err != nil {
		return p, err
	}
	return p, nil
}

// GoogleAPIError wraps a googleapi.Error and implements ExitCoder and StatusCode.
type GoogleAPIError struct {
	Err *googleapi.Error
}

func (e *GoogleAPIError) StatusCode() int {
	if e.Err != nil {
		return e.Err.Code
	}
	return 0
}

func (e *GoogleAPIError) ExitCode() int {
	if e.Err == nil {
		return ExitOk
	}
	if e.Err.Code == http.StatusTooManyRequests || e.Err.Code >= 500 {
		return ExitTempFail
	}
	if e.Err.Code >= 400 && e.Err.Code < 500 {
		return ExitUnavailable
	}
	return ExitTempFail
}

func (e *GoogleAPIError) Error() string {
	return e.Err.Error()
}

func (e *GoogleAPIError) Unwrap() error {
	return e.Err
}

func (p *GoogleProvider) sendMessageRest(bodyReader io.Reader) (*http.Response, error) {
	if requestBody, err := io.ReadAll(bodyReader); err != nil {
		panic(err)
	} else {
		gmsg := &gmail.Message{
			Raw: string(requestBody),
		}
		googleResponse, err := p.srv.Users.Messages.Send("me", gmsg).Do()
		if err != nil {
			var gErr *googleapi.Error
			if errors.As(err, &gErr) {
				return &http.Response{}, &GoogleAPIError{Err: gErr}
			}
			return &http.Response{}, err
		}
		return &http.Response{StatusCode: googleResponse.HTTPStatusCode, Header: googleResponse.Header}, err
	}
}
