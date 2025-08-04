package rest

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"

	"golang.org/x/oauth2/microsoft"
	"google.golang.org/api/gmail/v1"
)

type OutlookProvider struct {
	config           *oauth2.Config
	client           *http.Client
	provider, sender string
}

type GoogleProvider struct {
	config *oauth2.Config
	srv    *gmail.Service
}

type IProvider interface {
	//init(config *oauth2.Config) error
	SendMessage(io.Reader) error
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

func (p *OutlookProvider) SendMessage(messageReader io.Reader) error {
	return p.sendMessageRest(messageReader)
}

func NewProviderOutlook(conf *oauth2.Config, sender string) (IProvider, error) {
	var p = &OutlookProvider{
		provider: "outlook",
		config:   conf,
		sender:   sender,
	}
	_, err := p.getClient()
	return p, err
}

func (p *OutlookProvider) getClient() (*http.Client, error) {
	var (
		st  = SavedToken{provider: p.provider, id: p.sender}
		err error
	)
	if err := st.Open(); err != nil {
		return nil, err

	}
	ctx := context.Background()
	// pass through token source to refresh
	if st.token, err = p.config.TokenSource(ctx, st.token).Token(); err != nil {
		return nil, err
	} else if err := st.Save(); err != nil {
		return nil, err
	} else {
		p.client = p.config.Client(ctx, st.token)
		return p.client, nil
	}
}

// send from stdin
func (p *OutlookProvider) sendMessageRest(messageReader io.Reader) error {
	if encodedBuf, err := encodeMessage(messageReader); err != nil {
		return err
	} else if req, err := http.NewRequest("POST", MailSendEndpoint, encodedBuf); err != nil {
		return err
	} else {
		req.Header.Set("Content-type", "text/plain")
		if res, err := p.client.Do(req); err != nil {
			return err
		} else if res.StatusCode > 299 {
			return fmt.Errorf("error sending mail: statusCode = %d", res.StatusCode)
		}
	}
	return nil
}
