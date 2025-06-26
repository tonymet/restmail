package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/tonymet/restmail/rest"
	"golang.org/x/oauth2"
)

var (
	provider          string
	sender            string
	authorize, dummyI bool
	initConfig        rest.OAuthConfigJSON
	configClient      bool
	cmdConfigClient   *flag.FlagSet
)

func init() {
	flag.StringVar(&sender, "f", "", "Specifies the sender's email address.")
	flag.BoolVar(&authorize, "authorize", false, "Set up the OAuth2 authorization token before sending mail")
	flag.BoolVar(&dummyI, "i", true, "Dummy flag for compatibility with sendmail.")
	flag.StringVar(&provider, "provider", "gmail", "gmail|outlook -- which provider to use")
	flag.BoolVar(&configClient, "configClient", false, "start initial client config")
	flag.StringVar(&initConfig.Web.ClientID, "clientId", "", "OAuth2 Client ID")
	flag.StringVar(&initConfig.Web.ClientSecret, "clientSecret", "", "OAuth2 Client Secret")
}

func main() {
	flag.Parse()
	if !strings.ContainsRune(sender, '@') {
		log.Fatalf("error: sender should be an email")
	}
	var (
		oauthConfig *oauth2.Config
		err         error
		p           rest.IProvider
	)
	// initial client config
	if configClient {
		var sConfig = rest.SavedConfig{Provider: provider}
		if initConfig.Web.ClientID == "" || initConfig.Web.ClientSecret == "" || provider == "" {
			log.Fatal("-clientID , -provider, and -clientSecret need to be set")
		}
		sConfig.ConfigParams = initConfig
		err := sConfig.Save()
		initConfig := oauth2.Config{ClientID: sConfig.ConfigParams.Web.ClientID}
		rest.CreateInitialToken(&initConfig, provider, sender)

		if err != nil {
			panic(err)
		}
		os.Exit(0)
	}

	if oauthConfig, err = rest.OpenConfig(provider); err != nil {
		panic(err)
	}
	switch provider {
	case "outlook":
		p, err = rest.NewProviderOutlook(oauthConfig)
		if err != nil {
			log.Printf("error accessing token: %s", err)
		}
	case "gmail":
		p, err = rest.NewProviderGoogle(provider, sender)
		if err != nil {
			log.Printf("error accessing token: %s", err)
		}
	default:
		flag.PrintDefaults()
	}
	if authorize {
		rest.SetUpToken(oauthConfig, provider, sender)
		return
	} else {
		p.SendMessage(os.Stdin)
	}
}
