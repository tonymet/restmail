package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"

	"golang.org/x/oauth2"
)

type OAuthConfigJSON struct {
	Web struct {
		ClientID                string   `json:"client_id"`
		ProjectID               string   `json:"project_id"`
		AuthURI                 string   `json:"auth_uri"`
		TokenURI                string   `json:"token_uri"`
		AuthProviderX509CertURL string   `json:"auth_provider_x509_cert_url"`
		ClientSecret            string   `json:"client_secret"`
		RedirectUris            []string `json:"redirect_uris"`
		JavascriptOrigins       []string `json:"javascript_origins"`
		Scopes                  []string `json:"scopes"`
	} `json:"web"`
}

type SavedConfig struct {
	configParams OAuthConfigJSON
}

func OpenConfig(provider string) (*oauth2.Config, error) {
	var s SavedConfig
	if home, err := os.UserHomeDir(); err != nil {
		panic(err)
	} else if f, err := os.Open(path.Join(home, ".config/restmail/"+provider+".json")); err != nil {
		return nil, fmt.Errorf("Provider config not found: %s", err)
	} else if buf, err := io.ReadAll(f); err != nil {
		panic(err)
	} else if err := json.Unmarshal(buf, &s.configParams); err != nil {
		panic(err)
	}
	switch provider {
	case "gmail":
		var r = googleOAuth2Config
		r.ClientID = s.configParams.Web.ClientID
		r.ClientSecret = s.configParams.Web.ClientSecret
		return r, nil
	case "outlook":
		var r = outlookOAuth2Config
		r.ClientID = s.configParams.Web.ClientID
		r.ClientSecret = s.configParams.Web.ClientSecret
		return r, nil
	default:
		panic(fmt.Errorf("provider does not exist: %s", provider))
	}

}
