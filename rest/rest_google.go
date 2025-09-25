package rest

import (
	"context"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"google.golang.org/api/gmail/v1"
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

func (p *GoogleProvider) SendMessage(messageReader io.Reader) error {
	if encodedMessage, err := encodeMessage(messageReader); err != nil {
		panic(err)
	} else if _, err := p.sendMessageRest(encodedMessage); err != nil {
		panic(err)
	}
	return nil
}

func NewProviderGoogle(provider, sender string) (IProvider, error) {
	var (
		p   = &GoogleProvider{config: googleOAuth2Config}
		err error
	)
	ctx := context.Background()
	var st = &SavedToken{provider: provider, id: sender, config: googleOAuth2Config}
	if err := st.Open(); err != nil {
		return nil, err
	}
	// use refresh tokensource
	if p.srv, err = gmail.NewService(ctx, option.WithTokenSource(st)); err != nil {
		panic(err)
	}
	return p, nil
}

func (p *GoogleProvider) sendMessageRest(bodyReader io.ReadCloser) (*http.Response, error) {
	if requestBody, err := io.ReadAll(bodyReader); err != nil {
		panic(err)
	} else {
		gmsg := &gmail.Message{
			Raw: string(requestBody),
		}
		googleResponse, err := p.srv.Users.Messages.Send("me", gmsg).Do()
		if err != nil {
			return &http.Response{}, err
		}
		return &http.Response{StatusCode: googleResponse.ServerResponse.HTTPStatusCode, Header: googleResponse.ServerResponse.Header}, err
	}
}
