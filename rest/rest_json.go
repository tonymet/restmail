package rest

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
	ConfigParams OAuthConfigJSON
	Provider     string
}

func GetConfigReader(provider string) (io.Reader, error) {
	if home, err := os.UserHomeDir(); err != nil {
		panic(err)
	} else if err := os.MkdirAll(path.Join(home, ".config/restmail"), 0755); err != nil {
		panic(err)
	} else if f, err := os.Open(path.Join(home, ".config/restmail/"+provider+".json")); err != nil {
		return nil, fmt.Errorf("provider config not found: %s", err)
	} else {
		return f, nil
	}
}

func OpenConfig(provider string) (r *oauth2.Config, err error) {
	var (
		f io.Reader
		s SavedConfig
	)
	if f, err = GetConfigReader(provider); err != nil {
		return &oauth2.Config{}, err
	} else {
		decoder := json.NewDecoder(f)
		err := decoder.Decode(&s.ConfigParams)
		if err != nil {
			return &oauth2.Config{}, err
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
	} else {
		defer f.Close()
		enc := json.NewEncoder(f)
		return enc.Encode(&s.ConfigParams)
	}
}
