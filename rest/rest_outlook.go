package rest

import (
	"encoding/json"
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
	SendMessage(io.Reader, []string) error
	SendMessageOpt(io.Reader, []string, bool) error
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

func (p *OutlookProvider) SendMessage(messageReader io.Reader, args []string) error {
	return p.SendMessageOpt(messageReader, args, false)
}

func (p *OutlookProvider) SendMessageOpt(messageReader io.Reader, args []string, parseHeaders bool) error {
	return p.sendMessageRest(messageReader, args, parseHeaders)
}

func NewProviderOutlook(conf *oauth2.Config, sender string, storage ConfigStorage) (IProvider, error) {
	var p = &OutlookProvider{
		provider: "outlook",
		config:   conf,
		sender:   sender,
	}
	st := SavedToken{provider: p.provider, id: p.sender, Storage: storage}
	if err := st.Open(); err != nil {
		return nil, err

	}
	// pass through token source to refresh
	p.client = p.config.Client(storage.Context(), st.token)
	return p, nil
}

// GraphError represents an error returned by the Microsoft Graph API.
type GraphError struct {
	HTTPStatusCode int
	Code           string `json:"code"`
	Message        string `json:"message"`
}

func (e *GraphError) StatusCode() int {
	return e.HTTPStatusCode
}

func (e *GraphError) ExitCode() int {
	if e.HTTPStatusCode == http.StatusTooManyRequests || e.HTTPStatusCode >= 500 {
		return ExitTempFail
	}
	if e.HTTPStatusCode >= 400 && e.HTTPStatusCode < 500 {
		return ExitUnavailable
	}
	return ExitTempFail
}

func (e *GraphError) Error() string {
	if e.Code != "" || e.Message != "" {
		return fmt.Sprintf("Graph API error (HTTP %d, %s): %s", e.HTTPStatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("Graph API error: statusCode = %d", e.HTTPStatusCode)
}

func parseGraphError(res *http.Response) *GraphError {
	graphErr := &GraphError{HTTPStatusCode: res.StatusCode}
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil || len(bodyBytes) == 0 {
		return graphErr
	}

	var parsed struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(bodyBytes, &parsed); err == nil && (parsed.Error.Code != "" || parsed.Error.Message != "") {
		graphErr.Code = parsed.Error.Code
		graphErr.Message = parsed.Error.Message
	}
	return graphErr
}

// send from stdin
func (p *OutlookProvider) sendMessageRest(messageReader io.Reader, args []string, parseHeaders bool) error {
	if encodedBuf, err := encodeMessageOpt(messageReader, args, parseHeaders); err != nil {
		return err
	} else if req, err := http.NewRequest("POST", MailSendEndpoint, encodedBuf); err != nil {
		return err
	} else {
		req.Header.Set("Content-type", "text/plain")
		res, err := p.client.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.StatusCode > 299 {
			return parseGraphError(res)
		}
	}
	return nil
}
