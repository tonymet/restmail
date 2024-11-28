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
	)
	if oauthConfig, err = OpenConfig(provider); err != nil {
		panic(err)
	}
	if setUp {
		setUpToken(oauthConfig)
		return
	}
	switch provider {
	case "outlook":
		p, _ := NewProviderOutlook()
		p.sendMessage(os.Stdin)
	case "gmail":
		p, _ := NewProviderGoogle()
		p.sendMessage(os.Stdin)
	default:
		flag.PrintDefaults()
	}
}
