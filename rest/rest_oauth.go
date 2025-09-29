package rest

import (
	"encoding/json"
	"fmt"
	"path"

	"golang.org/x/oauth2"
)

type SavedToken struct {
	provider, id string
	token        *oauth2.Token
	config       *oauth2.Config
	Storage      ConfigStorage
}

func PathToken(prefix, provider, id string) string {
	return path.Join(prefix, fmt.Sprintf("token-%s-%s.json", provider, id))
}

func CreateInitialToken(oauthConfig *oauth2.Config, provider, sender string, storage ConfigStorage) error {
	var savedToken = SavedToken{token: &oauth2.Token{}, config: oauthConfig, provider: provider, id: sender, Storage: storage}
	return savedToken.Save()
}

func OAuthFlowToken(oauthConfig *oauth2.Config, provider, sender string, storage ConfigStorage) {
	aHandler := newCallbackHandler()
	if oauthToken, err := aHandler.getCredentials(oauthConfig); err != nil {
		panic(err)
	} else {
		savedToken := &SavedToken{token: oauthToken, provider: provider, id: sender, Storage: storage}
		if err := savedToken.Save(); err != nil {
			panic(err)
		}
	}
}

func (s *SavedToken) Open() error {
	// update to operate on reader
	var tokenVar oauth2.Token
	reader, err := s.Storage.GetReader()
	if err != nil {
		return err
	}
	defer reader.Close() //nolint: errCheck
	decoder := json.NewDecoder(reader)
	err = decoder.Decode(&tokenVar)
	if err != nil {
		return err
	}
	s.token = &tokenVar
	return nil
}

func (s *SavedToken) Save() error {
	if f, err := s.Storage.GetWriter(); err != nil {
		return err
	} else {
		defer f.Close() //nolint: errCheck
		enc := json.NewEncoder(f)
		return enc.Encode(*(s.token))
	}
}

// wrap oauth2.Conf.TokenSource to handle saving the token when it changes
func (s *SavedToken) Token() (*oauth2.Token, error) {
	oldToken := s.token
	ts := s.config.TokenSource(s.Storage.Context(), oldToken) // this refreshes the token if needed
	newToken, err := ts.Token()
	if err != nil {
		return newToken, err
	}
	if newToken.AccessToken != oldToken.AccessToken {
		s.token = newToken
		err := s.Save()
		if err != nil {
			return newToken, err
		}
	}
	return newToken, nil
}
