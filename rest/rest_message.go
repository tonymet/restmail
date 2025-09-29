package rest

import (
	"bytes"
	"encoding/base64"
	"io"
	"strings"
)

const MIME_LINE = "\r\n"

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

func encodeMessage(in io.Reader, args []string) (io.Reader, error) {
	header := parseArgs(args)
	var encodedBuf = bytes.NewBuffer(make([]byte, 0, 2048))
	messageEncoder := base64.NewEncoder(base64.StdEncoding, encodedBuf)
	defer messageEncoder.Close() //nolint: errcheck
	if _, err := io.Copy(messageEncoder, header.mimeHeader()); err != nil {
		return nil, err
	}
	if _, err := io.Copy(messageEncoder, in); err != nil {
		return nil, err
	}
	return encodedBuf, nil
}
