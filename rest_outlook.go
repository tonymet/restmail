package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"golang.org/x/oauth2/microsoft"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type Provider struct {
	config   *oauth2.Config
	client   *http.Client
	provider string
}

type GoogleProvider struct {
	config *oauth2.Config
	srv    *gmail.Service
}

type IProvider interface {
	//init(config *oauth2.Config) error
	sendMessage(io.Reader) error
}

var (
	OpenIdEndpoint = "https://login.microsoftonline.com/consumers/v2.0/.well-known/openid-configuration"
	Endpoint       = oauth2.Endpoint{
		AuthURL:       "https://login.microsoftonline.com/consumers/oauth2/v2.0/authorize",
		TokenURL:      "https://login.microsoftonline.com/consumers/oauth2/v2.0/token",
		DeviceAuthURL: "https://login.microsoftonline.com/consumers/oauth2/v2.0/devicecode",
		AuthStyle:     oauth2.AuthStyleInParams,
	}
	MailSendEndpoint = "https://graph.microsoft.com/v1.0/me/sendMail"
)

var outlookOAuth2Config = &oauth2.Config{
	Scopes:   []string{"Mail.Send", "offline_access"},
	Endpoint: microsoft.AzureADEndpoint("consumers"),
}

func (p *Provider) sendMessage(messageReader io.Reader) error {
	return p.sendMessageRest(messageReader)
}

func NewProviderOutlook() (IProvider, error) {
	var p *Provider = &Provider{provider: provider}
	p.client = p.getClient(outlookOAuth2Config)
	p.config = outlookOAuth2Config
	return p, nil
}

func (p Provider) getClient(conf *oauth2.Config) *http.Client {
	var st = SavedToken{provider: p.provider, id: sender}
	if err := st.Open(); err != nil {
		panic(err)
	}
	ctx := context.Background()
	if tok, err := st.Token(); err != nil {
		panic(err)
	} else {
		return conf.Client(ctx, tok)
	}
}

type SavedToken struct {
	provider, id string
	reader       io.Reader
	token        *oauth2.Token
}

func (s SavedToken) Path() (string, error) {
	if home, err := os.UserHomeDir(); err != nil {
		panic(err)
	} else {
		return path.Join(home, fmt.Sprintf(".config/restmail/%s/%s.json", s.id, s.provider)), nil
	}
}

func (s *SavedToken) Open() error {
	var tokenVar oauth2.Token
	if configPath, err := s.Path(); err != nil {
		panic(err)
	} else if s.reader, err = os.Open(configPath); err != nil {
		panic(err)
	} else if jsonMem, err := io.ReadAll(s.reader); err != nil {
		panic(err)
	} else if err := json.Unmarshal(jsonMem, &tokenVar); err != nil {
		panic(err)
	}
	s.token = &tokenVar
	return nil
}

func (s *SavedToken) Save() error {
	if tokenJson, err := json.Marshal(*(s.token)); err != nil {
		panic(err)
	} else if configPath, err := s.Path(); err != nil {
		panic(err)
	} else {
		return os.WriteFile(configPath, tokenJson, 0600)
	}
}

func (s *SavedToken) Token() (*oauth2.Token, error) {
	return s.token, nil
}

func encodeMessage(in io.Reader) (io.ReadCloser, error) {
	message, err := io.ReadAll(in)
	if err != nil {
		log.Fatalf("unable to read stdin")
	}
	header := parseArgs(flag.Args())
	headerByte := []byte(header.mimeHeader())
	var messageBuf bytes.Buffer
	messageBuf.Write(headerByte)
	messageBuf.Write(message)

	var encodedBuf bytes.Buffer
	messageEncoder := base64.NewEncoder(base64.StdEncoding, &encodedBuf)
	if _, err := io.Copy(messageEncoder, &messageBuf); err != nil {
		panic(err)
	}
	return io.NopCloser(&encodedBuf), nil
}

// send from stdin
func (p *Provider) sendMessageRest(messageReader io.Reader) error {
	if encodedBuf, err := encodeMessage(messageReader); err != nil {
		return err
	} else if req, err := http.NewRequest("POST", MailSendEndpoint, encodedBuf); err != nil {
		return err
	} else {
		req.Header.Set("Content-type", "text/plain")
		if res, err := p.client.Do(req); err != nil {
			panic(err)
		} else if res.StatusCode > 299 {
			panic(fmt.Errorf("error sending mail: statusCode = %d", res.StatusCode))
		}
	}
	return nil
}

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
		panic(err)
	}
	// use refresh tokensource
	if token, err := st.Token(); err != nil {
		panic(err)
	} else {
		ts := p.config.TokenSource(ctx, token)
		p.srv, err = gmail.NewService(ctx, option.WithTokenSource(ts))
		if err != nil {
			panic(err)
		} else if st.token, err = ts.Token(); err != nil {
			panic(err)
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
