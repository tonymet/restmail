package rest

import (
	"bytes"
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
	ConfigParams OAuthConfigJSON
	Provider     string
}

func OpenConfig(provider string) (r *oauth2.Config, err error) {
	var s SavedConfig
	if home, err := os.UserHomeDir(); err != nil {
		panic(err)
	} else if err := os.MkdirAll(path.Join(home, ".config/restmail"), 0755); err != nil {
		panic(err)
	} else if f, err := os.Open(path.Join(home, ".config/restmail/"+provider+".json")); err != nil {
		return nil, fmt.Errorf("provider config not found: %s", err)
	} else {
		defer f.Close()
		if buf, err := io.ReadAll(f); err != nil {
			panic(err)
		} else if err := json.Unmarshal(buf, &s.ConfigParams); err != nil {
			panic(err)
		}
	}
	switch provider {
	case "gmail":
		r = googleOAuth2Config
	case "outlook":
		r = outlookOAuth2Config
	default:
		panic(fmt.Errorf("provider does not exist: %s", provider))
	}
	r.ClientID = s.ConfigParams.Web.ClientID
	r.ClientSecret = s.ConfigParams.Web.ClientSecret
	return r, nil
}

func (s *SavedConfig) Save() error {
	if home, err := os.UserHomeDir(); err != nil {
		panic(err)
	} else if err := os.MkdirAll(path.Join(home, ".config/restmail"), 0755); err != nil {
		panic(err)
	} else if f, err := os.Create(path.Join(home, ".config/restmail/"+s.Provider+".json")); err != nil {
		return fmt.Errorf("provider config not found: %s", err)
	} else if jsonContents, err := json.Marshal(s.ConfigParams); err != nil {
		defer f.Close()
		return err
	} else {
		defer f.Close()
		buf := bytes.NewBuffer(jsonContents)
		_, err := io.Copy(f, buf)
		if err != nil {
			return err
		}
		return nil
	}

}
