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
		key, val, hasColon := strings.Cut(arg, ":")
		if hasColon {
			val, _, _ = strings.Cut(val, ":")
			switch key {
			case "bcc":
				mh.bcc = append(mh.bcc, val)
				continue
			case "cc":
				mh.cc = append(mh.cc, val)
				continue
			}
		}
		mh.to = append(mh.to, key)
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

func writeHeaderLine(b *strings.Builder, name string, values []string) {
	b.WriteString(name)
	b.WriteString(": ")
	for i, v := range values {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(v)
	}
	b.WriteString(MIME_LINE)
}

func (mh messageHeader) mimeHeader() io.Reader {
	var header strings.Builder
	header.WriteString("MIME-Version: 1.0")
	header.WriteString(MIME_LINE)
	header.WriteString("Content-Transfer-Encoding: quoted-printable")
	header.WriteString(MIME_LINE)
	writeHeaderLine(&header, "To", mh.to)
	if len(mh.bcc) > 0 {
		writeHeaderLine(&header, "Bcc", mh.bcc)
	}
	if len(mh.cc) > 0 {
		writeHeaderLine(&header, "Cc", mh.cc)
	}
	return strings.NewReader(header.String())
}

func encodeMessage(in io.Reader, args []string) (*bytes.Buffer, error) {
	return encodeMessageOpt(in, args, false)
}

func prependMissingMIMEHeaders(in io.Reader) (io.Reader, error) {
	raw, err := io.ReadAll(in)
	if err != nil {
		return nil, err
	}
	msg, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return bytes.NewReader(raw), nil
	}
	var injected strings.Builder
	if msg.Header.Get("Mime-Version") == "" {
		injected.WriteString("MIME-Version: 1.0")
		injected.WriteString(MIME_LINE)
	}
	if msg.Header.Get("Content-Transfer-Encoding") == "" {
		injected.WriteString("Content-Transfer-Encoding: quoted-printable")
		injected.WriteString(MIME_LINE)
	}
	if injected.Len() > 0 {
		return io.MultiReader(strings.NewReader(injected.String()), bytes.NewReader(raw)), nil
	}
	return bytes.NewReader(raw), nil
}

func encodeMessageOpt(in io.Reader, args []string, parseHeaders bool) (*bytes.Buffer, error) {
	var (
		finalPayload io.Reader
		err          error
	)

	if parseHeaders {
		if finalPayload, err = prependMissingMIMEHeaders(in); err != nil {
			return nil, err
		}
	} else {
		parsedArgs := parseArgs(args)
		finalPayload = io.MultiReader(parsedArgs.mimeHeader(), in)
	}

	encodedBuf := bytes.NewBuffer(make([]byte, 0, 2048))
	messageEncoder := base64.NewEncoder(base64.StdEncoding, encodedBuf)
	if _, err := io.Copy(messageEncoder, finalPayload); err != nil {
		_ = messageEncoder.Close()
		return nil, err
	}
	if err := messageEncoder.Close(); err != nil {
		return nil, err
	}
	return encodedBuf, nil
}
