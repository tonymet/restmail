package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"golang.org/x/oauth2"
)

var (
	provider, dummyF string
	sender           string
	setUp, dummyI    bool
)

const MIME_LINE = "\r\n"

func init() {
	flag.StringVar(&sender, "sender", "", "Specifies the sender's email address.")
	flag.BoolVar(&setUp, "setup", false, "Set up the OAuth2 authorization token before sending mail")
	flag.StringVar(&dummyF, "f", "", "Dummy flag for compatibility with sendmail.")
	flag.BoolVar(&dummyI, "i", true, "Dummy flag for compatibility with sendmail.")
	flag.StringVar(&provider, "provider", "gmail", "gmail|outlook -- which provider to use")
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
	if oauthConfig, err = OpenConfig(provider); err != nil {
		panic(err)
	}
	switch provider {
	case "outlook":
		p, _ = NewProviderOutlook(oauthConfig)
	case "gmail":
		p, _ = NewProviderGoogle()
	default:
		flag.PrintDefaults()
	}
	if setUp {
		setUpToken(oauthConfig)
		return
	} else {
		p.sendMessage(os.Stdin)
	}
}
