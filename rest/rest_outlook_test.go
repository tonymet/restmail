package rest

import (
	"bytes"
	"encoding/base64"
	"flag"
	"io"
	"log"
	"os"
	"strings"
	"testing"
)

// benchmark comparison method
func encodeMessageString(in io.Reader) (string, error) {
	message, err := io.ReadAll(in)
	if err != nil {
		log.Fatalf("unable to read stdin")
	}
	header := parseArgs(flag.Args())
	var messageString strings.Builder
	io.Copy(&messageString, header.mimeHeader())
	messageString.WriteString(string(message))

	return base64.StdEncoding.EncodeToString([]byte(message)), nil
}

func BenchmarkEncodeMessage(b *testing.B) {
	if in, err := os.Open("/etc/mime.types"); err != nil {
		panic(err)
	} else {
		var buf bytes.Buffer
		io.Copy(&buf, in)
		r := bytes.NewReader(buf.Bytes())
		b.ResetTimer()
		for range b.N {
			r.Seek(0, 0)
			if rc, err := encodeMessage(r); err != nil {
				panic(err)
			} else {
				io.Copy(io.Discard, rc)
			}
		}
	}
}

func BenchmarkEncodeMessageString(b *testing.B) {
	if in, err := os.Open("/etc/mime.types"); err != nil {
		panic(err)
	} else {
		var buf bytes.Buffer
		io.Copy(&buf, in)
		r := bytes.NewReader(buf.Bytes())
		b.ResetTimer()
		for range b.N {
			r.Seek(0, 0)
			if s, err := encodeMessageString(r); err != nil {
				panic(err)
			} else {
				r := strings.NewReader(s)
				io.Copy(io.Discard, r)
			}
		}
	}
}
