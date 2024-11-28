package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"

	"golang.org/x/oauth2"
)

type SavedToken struct {
	provider, id string
	reader       io.Reader
	token        *oauth2.Token
}

func (s SavedToken) Path() (string, error) {
	if home, err := os.UserHomeDir(); err != nil {
		panic(err)
	} else {
		return path.Join(home, fmt.Sprintf(".config/restmail/%s/%s.json", s.id, s.provider)), nil
	}
}

func setUpToken(oauthConfig *oauth2.Config) {
	aHandler := newCallbackHandler()
	if savedToken, err := aHandler.getCredentials(oauthConfig); err != nil {
		panic(err)
	} else {
		savedToken.provider = "gmail"
		savedToken.id = sender
		if err := savedToken.Save(); err != nil {
			panic(err)
		}
	}
}

func (s *SavedToken) Open() error {
	var tokenVar oauth2.Token
	if configPath, err := s.Path(); err != nil {
		panic(err)
	} else if s.reader, err = os.Open(configPath); err != nil {
		panic(err)
	} else if jsonMem, err := io.ReadAll(s.reader); err != nil {
		panic(err)
	} else if err := json.Unmarshal(jsonMem, &tokenVar); err != nil {
		panic(err)
	}
	s.token = &tokenVar
	return nil
}

func (s *SavedToken) Save() error {
	if tokenJson, err := json.Marshal(*(s.token)); err != nil {
		panic(err)
	} else if configPath, err := s.Path(); err != nil {
		panic(err)
	} else {
		return os.WriteFile(configPath, tokenJson, 0600)
	}
}

func (s *SavedToken) Token() (*oauth2.Token, error) {
	return s.token, nil
}
