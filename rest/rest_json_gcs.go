package rest

import (
	"fmt"
	"io"

	"cloud.google.com/go/storage"
)

type ConfigStorageBucket struct {
	ConfigStorageBase
	Client *storage.Client
	Bucket string
	Path   func(string, string, string) string
}

func (s *ConfigStorageBucket) Setup() (err error) {
	s.Client, err = storage.NewClient(s.Context())
	return
}

func (s ConfigStorageBucket) GetReader() (io.ReadCloser, error) {
	objectName := s.Path(s.Prefix, s.Provider, s.Id)
	rc, err := s.Client.Bucket(s.Bucket).Object(objectName).NewReader(s.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to open GCS config object %s: %w", objectName, err)
	}
	return rc, nil
}

func (s ConfigStorageBucket) GetWriter() (io.WriteCloser, error) {
	objectName := s.Path(s.Prefix, s.Provider, s.Id)
	wc := s.Client.Bucket(s.Bucket).Object(objectName).NewWriter(s.Context())
	wc.ContentType = "application/json"
	return wc, nil
}
