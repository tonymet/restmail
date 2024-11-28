package main

import (
	"flag"
	"io"
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
	flag.BoolVar(&setUp, "setup", false, "If true, sendgmail sets up the sender's OAuth2 token and then exits.")
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

func setUpToken(oauthConfig *oauth2.Config) {
	aHandler := newCallbackHandler()
	if savedToken, err := aHandler.getCredentials(oauthConfig); err != nil {
		panic(err)
	} else {
		savedToken.provider = provider
		savedToken.id = sender
		if err := savedToken.Save(); err != nil {
			panic(err)
		}
	}
}

type messageHeader struct {
	to, cc, bcc []string
}

func parseArgs(args []string) (mh messageHeader) {
	mh.to, mh.cc, mh.bcc = make([]string, 0), make([]string, 0), make([]string, 0)
	for _, arg := range args {
		parts := strings.Split(arg, ":")
		switch parts[0] {
		case "bcc":
			mh.bcc = append(mh.bcc, parts[1])
		case "cc":
			mh.cc = append(mh.cc, parts[1])
		default:
			mh.to = append(mh.to, parts[0])
		}
	}
	return
}

func (mh messageHeader) mimeHeader() io.Reader {
	var header strings.Builder
	header.WriteString("To: " + strings.Join(mh.to, ",") + MIME_LINE)
	if len(mh.bcc) > 0 {
		header.WriteString("Bcc: " + strings.Join(mh.bcc, ",") + MIME_LINE)
	}
	if len(mh.cc) > 0 {
		header.WriteString("Cc: " + strings.Join(mh.cc, ",") + MIME_LINE)
	}
	return strings.NewReader(header.String())
}
