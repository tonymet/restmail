package rest

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// http server handling oauth2 callback 3-legged-flow
type callbackHandler struct {
	codeChan chan string
	state    string
}

// basic handler on / for receiving code
func (h *callbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if query.Has("error") {
		log.Printf("Oauth error: %s\n", query.Get("error"))
		return
	}
	if !(query.Has("code")) {
		if _, err := w.Write([]byte("expecting query param: code")); err != nil {
			panic(err)
		}
		return
	}
	code := query.Get("code")
	if _, err := w.Write([]byte(fmt.Sprintf("code: %s\n", code))); err != nil {
		panic(err)
	}
	h.codeChan <- code
	close(h.codeChan)
}

// prompt user on CLI for oauth consent screen, wait on channel for code from http
func (h *callbackHandler) authHandler(authCodeURL, sender string) (string, string, error) {
	fmt.Println()
	fmt.Println("1. Ensure that you are logged in as", sender, "in your browser.")
	fmt.Println()
	fmt.Println("2. Click this link to authorize:")
	fmt.Println(authCodeURL) // hack to obtain a refresh token
	fmt.Println()
	fmt.Println("3. Waiting for authorization code:")
	return <-h.codeChan, h.state, nil
}

// oauth signing / challenge / etc
func (h *callbackHandler) getCredentials(oauthConfig *oauth2.Config) (*SavedToken, error) {
	ctx := context.Background()
	verifier := oauth2.GenerateVerifier()

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	url := oauthConfig.AuthCodeURL(h.state, oauth2.AccessTypeOffline, oauth2.ApprovalForce, oauth2.S256ChallengeOption(verifier))
	code, state, err := h.authHandler(url, oauthConfig.ClientID)
	if err != nil {
		panic(err)
	}
	if state != h.state {
		panic(fmt.Errorf("state mismatch expect %s, received %s", h.state, state))
	}
	if tok, err := oauthConfig.Exchange(ctx, code, oauth2.VerifierOption(verifier)); err != nil {
		panic(err)
	} else {
		//todo link to provider
		var s SavedToken
		s.token = tok
		return &s, nil
	}
}

// set up the server and listen on localhost:5000
func newCallbackHandler() *callbackHandler {
	var h callbackHandler
	h.codeChan = make(chan string)
	h.state = uuid.NewString()
	go func() {
		http.Handle("/", &h)
		if err := http.ListenAndServe("127.0.0.1:5000", nil); err != nil {
			panic(err)
		}
	}()
	return &h
}
