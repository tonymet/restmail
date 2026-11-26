package rest

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/oauth2/google"
)

type ConfigStorageBucket struct {
	ConfigStorageBase
	Client *http.Client
	Bucket string
	Path   func(string, string, string) string
}

func (s *ConfigStorageBucket) Setup() (err error) {
	s.Client, err = google.DefaultClient(s.Context(), "https://www.googleapis.com/auth/devstorage.read_write")
	return
}

func (s ConfigStorageBucket) GetReader() (io.ReadCloser, error) {
	objectName := s.Path(s.Prefix, s.Provider, s.Id)
	endpoint := fmt.Sprintf("https://storage.googleapis.com/storage/v1/b/%s/o/%s?alt=media",
		url.PathEscape(s.Bucket), url.PathEscape(objectName))
	req, err := http.NewRequestWithContext(s.Context(), http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open GCS config object %s: %w", objectName, err)
	}
	res, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to open GCS config object %s: %w", objectName, err)
	}
	if res.StatusCode > 299 {
		res.Body.Close()
		return nil, fmt.Errorf("failed to open GCS config object %s: HTTP %d", objectName, res.StatusCode)
	}
	return res.Body, nil
}

type gcsPipeWriter struct {
	pw   *io.PipeWriter
	done chan error
}

func (w *gcsPipeWriter) Write(p []byte) (int, error) {
	return w.pw.Write(p)
}

func (w *gcsPipeWriter) Close() error {
	if err := w.pw.Close(); err != nil {
		return err
	}
	return <-w.done
}

func (s ConfigStorageBucket) GetWriter() (io.WriteCloser, error) {
	objectName := s.Path(s.Prefix, s.Provider, s.Id)
	endpoint := fmt.Sprintf("https://storage.googleapis.com/upload/storage/v1/b/%s/o?uploadType=media&name=%s",
		url.PathEscape(s.Bucket), url.QueryEscape(objectName))
	pr, pw := io.Pipe()
	done := make(chan error, 1)

	req, err := http.NewRequestWithContext(s.Context(), http.MethodPost, endpoint, pr)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	go func() {
		res, err := s.Client.Do(req)
		if err != nil {
			pr.CloseWithError(err)
			done <- err
			return
		}
		res.Body.Close()
		if res.StatusCode > 299 {
			done <- fmt.Errorf("failed to write GCS config object %s: HTTP %d", objectName, res.StatusCode)
			return
		}
		done <- nil
	}()

	return &gcsPipeWriter{pw: pw, done: done}, nil
}
