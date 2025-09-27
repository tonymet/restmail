package main

import (
	"flag"
	"os"
	"strings"

	"github.com/tonymet/restmail/rest"
	_ "golang.org/x/crypto/x509roots/fallback"
)

var (
	dummyI    bool
	appConfig rest.AppConfigContainer
)

func init() {
	flag.StringVar(&appConfig.Sender, "f", "", "Specifies the sender's email address.")
	flag.BoolVar(&appConfig.Authorize, "authorize", false, "Set up the OAuth2 authorization token before sending mail")
	flag.BoolVar(&dummyI, "i", true, "Dummy flag for compatibility with sendmail.")
	flag.StringVar(&appConfig.Provider, "provider", "gmail", "gmail|outlook -- which provider to use")
	flag.BoolVar(&appConfig.ConfigClient, "configClient", false, "start initial client config")
	flag.StringVar(&appConfig.Storage, "storage", "os", "os|gcs Where to save config")
	flag.StringVar(&appConfig.InitConfig.Web.ClientID, "clientId", "", "OAuth2 Client ID")
	flag.StringVar(&appConfig.InitConfig.Web.ClientSecret, "clientSecret", "", "OAuth2 Client Secret")
	flag.StringVar(&appConfig.Message, "m", "", "message content")
}

func main() {
	flag.Parse()
	if appConfig.Message != "" {
		appConfig.MessageReader = strings.NewReader(appConfig.Message)
	} else {
		appConfig.MessageReader = os.Stdin
	}

	rest.RunApp(&appConfig, flag.Args())
}
