package rest

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

type AppConfigContainer struct {
	Provider, Sender, Storage string
	InitConfig                OAuthConfigJSON
	ConfigClient, Authorize   bool
	Message                   string
	MessageReader             io.Reader
}

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

type ConfigStorage interface {
	GetReader() (io.ReadCloser, error)
	GetWriter() (io.WriteCloser, error)
	Context() context.Context
}

type ConfigStorageBase struct {
	Prefix, Provider, Id string
	Ctx                  context.Context
}

type ConfigStorageOS struct {
	ConfigStorageBase
	Path func(string, string, string) string
}

func PathConfig(prefix, provider, id string) string {
	return path.Join(prefix, "config-"+provider+".json")
}

type SavedConfig struct {
	ConfigParams OAuthConfigJSON
	Provider     string
	Storage      ConfigStorage
}

func (s ConfigStorageBase) Context() context.Context {
	return s.Ctx
}

func (s ConfigStorageOS) GetReader() (io.ReadCloser, error) {
	if home, err := os.UserHomeDir(); err != nil {
		return nil, err
	} else if err := os.MkdirAll(path.Join(home, s.Prefix), 0755); err != nil {
		return nil, err
	} else if f, err := os.Open(path.Join(home, s.Path(s.Prefix, s.Provider, s.Id))); err != nil {
		return nil, fmt.Errorf("provider config not found: %s", err)
	} else {
		return f, nil
	}
}

func (s ConfigStorageOS) GetWriter() (io.WriteCloser, error) {
	if home, err := os.UserHomeDir(); err != nil {
		return nil, err
	} else if err := os.MkdirAll(path.Join(home, s.Prefix), 0755); err != nil {
		return nil, err
	} else if f, err := os.Create(path.Join(home, s.Path(s.Prefix, s.Provider, s.Id))); err != nil {
		return nil, fmt.Errorf("provider config not found: %s", err)
	} else {
		return f, nil
	}
}

func OpenConfig(s ConfigStorage, provider string) (r *oauth2.Config, err error) {
	if f, err := s.GetReader(); err != nil {
		return &oauth2.Config{}, err
	} else {
		defer f.Close() //nolint: errcheck
		return loadConfig(f, provider)
	}
}

func loadConfig(f io.Reader, provider string) (r *oauth2.Config, err error) {
	var s SavedConfig
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&s.ConfigParams)
	if err != nil {
		return &oauth2.Config{}, err
	}
	switch provider {
	case "gmail":
		r = googleOAuth2Config
	case "outlook":
		r = outlookOAuth2Config
	default:
		log.Fatal(fmt.Errorf("provider does not exist: %s", provider))
	}
	r.ClientID = s.ConfigParams.Web.ClientID
	r.ClientSecret = s.ConfigParams.Web.ClientSecret
	return r, nil
}

func (s *SavedConfig) Save() error {
	if f, err := s.Storage.GetWriter(); err != nil {
		return err
	} else {
		defer f.Close() //nolint: errcheck
		enc := json.NewEncoder(f)
		return enc.Encode(&s.ConfigParams)
	}
}

func setupStorage(ctx context.Context, c AppConfigContainer) (storageConfig ConfigStorage, storageToken ConfigStorage) {
	storageOSBase := ConfigStorageBase{Provider: c.Provider, Prefix: ".config/restmail", Ctx: ctx}
	storageConfig = ConfigStorageOS{ConfigStorageBase: storageOSBase, Path: PathConfig}
	storageToken = ConfigStorageOS{ConfigStorageBase: storageOSBase, Path: PathToken}
	switch c.Storage {
	case "gcs":
		storageTokenBase := ConfigStorageBase{Prefix: os.Getenv("GCS_PREFIX"),
			Provider: c.Provider, Id: c.Sender, Ctx: ctx}
		storageConfig = &ConfigStorageBucket{ConfigStorageBase: storageTokenBase, Bucket: os.Getenv("GCS_BUCKET"), Path: PathConfig}
		err := storageConfig.(*ConfigStorageBucket).Setup()
		if err != nil {
			log.Fatal(err)
		}
		storageToken = &ConfigStorageBucket{ConfigStorageBase: storageTokenBase, Bucket: os.Getenv("GCS_BUCKET"), Path: PathToken}
		err = storageToken.(*ConfigStorageBucket).Setup()
		if err != nil {
			log.Fatal(err)
		}
	}
	return
}

func RunApp(c *AppConfigContainer, args []string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if !strings.ContainsRune(c.Sender, '@') {
		log.Fatalf("error: sender should be an email")
	}
	var (
		oauthConfig *oauth2.Config
		err         error
		p           IProvider
	)
	// initial client config
	storageConfig, storageToken := setupStorage(ctx, *c)
	if c.ConfigClient {
		var sConfig = SavedConfig{Provider: c.Provider, Storage: storageConfig}
		if c.InitConfig.Web.ClientID == "" || c.InitConfig.Web.ClientSecret == "" || c.Provider == "" {
			log.Fatal("-clientID , -provider, and -clientSecret need to be set")
		}
		sConfig.ConfigParams = c.InitConfig
		err := sConfig.Save()
		if err != nil {
			log.Fatal(err)
		}
		initConfig := oauth2.Config{ClientID: sConfig.ConfigParams.Web.ClientID}
		err = CreateInitialToken(&initConfig, c.Provider, c.Sender, storageToken)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	if oauthConfig, err = OpenConfig(storageConfig, c.Provider); err != nil {
		log.Fatal(err)
	}
	switch c.Provider {
	case "outlook":
		p, err = NewProviderOutlook(oauthConfig, c.Sender, storageToken)
		if err != nil {
			log.Printf("error accessing token: %s", err)
		}
	case "gmail":
		p, err = NewProviderGoogle(c.Provider, c.Sender, storageToken)
		if err != nil {
			log.Printf("error accessing token: %s", err)
		}
	default:
		flag.PrintDefaults()
	}
	if c.Authorize {
		OAuthFlowToken(oauthConfig, c.Provider, c.Sender, storageToken)
		return
	} else if err := p.SendMessage(c.MessageReader, args); err != nil {
		log.Fatal(err)
	}
}
