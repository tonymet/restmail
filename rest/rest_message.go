package rest

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/mail"
	"strings"
)

const MIME_LINE = "\r\n"

type messageHeader struct {
	from        string
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

func parseHeaderAddresses(header mail.Header, key string) []string {
	res := make([]string, 0)
	for _, v := range header[key] {
		addrs, err := mail.ParseAddressList(v)
		if err == nil {
			for _, a := range addrs {
				res = append(res, a.Address)
			}
		} else {
			for _, part := range strings.Split(v, ",") {
				trimmed := strings.TrimSpace(part)
				if trimmed != "" {
					res = append(res, trimmed)
				}
			}
		}
	}
	return res
}

type MessageHeader struct {
	From        string
	To, Cc, Bcc []string
}

func ParseMessageHeaders(in io.Reader) (MessageHeader, io.Reader, error) {
	raw, err := io.ReadAll(in)
	if err != nil {
		return MessageHeader{}, nil, err
	}

	msg, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return MessageHeader{}, bytes.NewReader(raw), nil
	}

	var fromAddr string
	if fromAddrs := parseHeaderAddresses(msg.Header, "From"); len(fromAddrs) > 0 {
		fromAddr = fromAddrs[0]
	}

	mh := MessageHeader{
		From: fromAddr,
		To:   parseHeaderAddresses(msg.Header, "To"),
		Cc:   parseHeaderAddresses(msg.Header, "Cc"),
		Bcc:  parseHeaderAddresses(msg.Header, "Bcc"),
	}

	return mh, bytes.NewReader(raw), nil
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
	return encodeMessageOpt(in, args, false)
}

func encodeMessageOpt(in io.Reader, args []string, parseHeaders bool) (io.Reader, error) {
	var (
		header       MessageHeader
		finalPayload io.Reader
	)

	if parseHeaders {
		parsedHdr, reader, err := ParseMessageHeaders(in)
		if err != nil {
			return nil, err
		}
		header = parsedHdr
		// The message from stdin already contains mime headers; send reader as is.
		finalPayload = reader
	} else {
		parsedArgs := parseArgs(args)
		header = MessageHeader{
			To:  parsedArgs.to,
			Cc:  parsedArgs.cc,
			Bcc: parsedArgs.bcc,
		}
		finalPayload = io.MultiReader(parsedArgs.mimeHeader(), in)
	}
	_ = header // available if recipient list is ever needed separately

	var encodedBuf = bytes.NewBuffer(make([]byte, 0, 2048))
	messageEncoder := base64.NewEncoder(base64.StdEncoding, encodedBuf)
	defer messageEncoder.Close() //nolint: errcheck
	if _, err := io.Copy(messageEncoder, finalPayload); err != nil {
		return nil, err
	}
	return encodedBuf, nil
}
