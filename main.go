package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"golang.org/x/oauth2"
)

var (
	provider          string
	sender            string
	authorize, dummyI bool
	initConfig        OAuthConfigJSON
	configClient      bool
	cmdConfigClient   *flag.FlagSet
)

const MIME_LINE = "\r\n"

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
		p           IProvider
	)
	// initial client config
	if configClient {
		var sConfig SavedConfig
		if initConfig.Web.ClientID == "" || initConfig.Web.ClientSecret == "" || provider == "" {
			log.Fatal("-clientID , -provider, and -clientSecret need to be set")
		}
		sConfig.configParams = initConfig
		err := sConfig.Save()
		if err != nil {
			panic(err)
		}
		os.Exit(0)
	}

	if oauthConfig, err = OpenConfig(provider); err != nil {
		panic(err)
	}
	switch provider {
	case "outlook":
		p, err = NewProviderOutlook(oauthConfig)
		if err != nil {
			log.Printf("error accessing token: %s", err)
		}
	case "gmail":
		p, err = NewProviderGoogle()
		if err != nil {
			log.Printf("error accessing token: %s", err)
		}
	default:
		flag.PrintDefaults()
	}
	if authorize {
		setUpToken(oauthConfig)
		return
	} else {
		p.sendMessage(os.Stdin)
	}
}
