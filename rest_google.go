package main

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

func (p *GoogleProvider) sendMessage(messageReader io.Reader) error {
	if encodedMessage, err := encodeMessage(messageReader); err != nil {
		panic(err)
	} else if _, err := p.sendMessageRest(encodedMessage); err != nil {
		panic(err)
	}
	return nil
}

func NewProviderGoogle() (IProvider, error) {
	var (
		p *GoogleProvider = &GoogleProvider{config: googleOAuth2Config}
	)
	ctx := context.Background()
	var st = SavedToken{provider: provider, id: sender}
	if err := st.Open(); err != nil {
		return nil, err
	}
	// use refresh tokensource
	if token, err := st.Token(); err != nil {
		panic(err)
	} else {
		ts := p.config.TokenSource(ctx, token)
		if p.srv, err = gmail.NewService(ctx, option.WithTokenSource(ts)); err != nil {
			panic(err)
		} else if st.token, err = ts.Token(); err != nil {
			return nil, err
		} else if err := st.Save(); err != nil {
			panic(err)
		}
		return p, nil
	}
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
