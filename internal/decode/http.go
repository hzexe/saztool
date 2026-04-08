package decode

import (
	"bufio"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/andybalholm/brotli"
	"golang.org/x/net/html/charset"
)

// ParsedHTTP is a lightweight split of raw HTTP text into header/body plus decoded body.
type ParsedHTTP struct {
	StartLine        string
	Header           http.Header
	BodyRaw          []byte
	BodyDecoded      []byte
	BodyText         string
	BodyIsText       bool
	TransferEncoding string
	ContentEncoding  string
	ContentType      string
	Charset          string
	Transforms       []string
	Method           string
	URL              string
	StatusCode       int
	DecodeFailed     bool
	DecodeFailure    string
}

func ParseResponse(raw []byte) (*ParsedHTTP, error) {
	headerPart, bodyPart, err := splitHTTP(raw)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(bytes.NewReader(append(headerPart, []byte("\r\n\r\n")...)))
	resp, err := http.ReadResponse(reader, nil)
	if err != nil {
		return nil, fmt.Errorf("parse response header: %w", err)
	}
	defer resp.Body.Close()
	parsed, err := decodeHTTP(resp.Status, resp.Header, bodyPart, resp.TransferEncoding)
	if err != nil {
		return nil, err
	}
	parsed.StatusCode = resp.StatusCode
	return parsed, nil
}

func ParseRequest(raw []byte) (*ParsedHTTP, error) {
	headerPart, bodyPart, err := splitHTTP(raw)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(bytes.NewReader(append(headerPart, []byte("\r\n\r\n")...)))
	req, err := http.ReadRequest(reader)
	if err != nil {
		return nil, fmt.Errorf("parse request header: %w", err)
	}
	defer req.Body.Close()
	parsed, err := decodeHTTP(req.Method+" "+req.URL.String()+" "+req.Proto, req.Header, bodyPart, req.TransferEncoding)
	if err != nil {
		return nil, err
	}
	parsed.Method = req.Method
	parsed.URL = req.URL.String()
	if req.Host != "" && strings.HasPrefix(parsed.URL, "/") {
		scheme := "http"
		if req.TLS != nil {
			scheme = "https"
		}
		parsed.URL = scheme + "://" + req.Host + parsed.URL
	}
	return parsed, nil
}

func decodeHTTP(startLine string, header http.Header, body []byte, parsedTransferEncoding []string) (*ParsedHTTP, error) {
	transferEncodingValues := header.Values("Transfer-Encoding")
	if len(transferEncodingValues) == 0 && len(parsedTransferEncoding) > 0 {
		transferEncodingValues = append([]string(nil), parsedTransferEncoding...)
	}

	p := &ParsedHTTP{
		StartLine:        startLine,
		Header:           header,
		BodyRaw:          body,
		BodyDecoded:      body,
		TransferEncoding: strings.Join(transferEncodingValues, ", "),
		ContentEncoding:  strings.Join(header.Values("Content-Encoding"), ", "),
		ContentType:      header.Get("Content-Type"),
	}

	if hasChunkedEncoding(transferEncodingValues) || looksLikeChunkedBody(p.BodyDecoded) {
		decoded, err := decodeChunked(p.BodyDecoded)
		if err == nil {
			p.BodyDecoded = decoded
			p.Transforms = append(p.Transforms, "dechunk")
			if p.TransferEncoding == "" {
				p.TransferEncoding = "chunked (detected)"
			}
		}
	}

	encodings := splitEncodings(p.ContentEncoding)
	for i := len(encodings) - 1; i >= 0; i-- {
		enc := encodings[i]
		decoded, transform, err := decompress(p.BodyDecoded, enc)
		if err != nil {
			p.DecodeFailed = true
			p.DecodeFailure = err.Error()
			break
		}
		p.BodyDecoded = decoded
		if transform != "" {
			p.Transforms = append(p.Transforms, transform)
		}
	}

	mediaType, params, _ := mime.ParseMediaType(p.ContentType)
	charsetName := strings.TrimSpace(params["charset"])
	p.Charset = charsetName
	if isTextual(mediaType, p.BodyDecoded) {
		text, transform, err := decodeCharset(p.BodyDecoded, p.ContentType)
		if err == nil {
			p.BodyText = text
			p.BodyIsText = true
			if transform != "" {
				p.Transforms = append(p.Transforms, transform)
			}
		}
	}

	return p, nil
}

func splitHTTP(raw []byte) ([]byte, []byte, error) {
	sep := []byte("\r\n\r\n")
	idx := bytes.Index(raw, sep)
	if idx < 0 {
		sep = []byte("\n\n")
		idx = bytes.Index(raw, sep)
	}
	if idx < 0 {
		return nil, nil, fmt.Errorf("could not split headers and body")
	}
	return raw[:idx], raw[idx+len(sep):], nil
}

func decodeChunked(body []byte) ([]byte, error) {
	return io.ReadAll(httputil.NewChunkedReader(bytes.NewReader(body)))
}

func decompress(body []byte, enc string) ([]byte, string, error) {
	switch strings.ToLower(strings.TrimSpace(enc)) {
	case "gzip", "x-gzip":
		gr, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			return nil, "", err
		}
		defer gr.Close()
		data, err := io.ReadAll(gr)
		return data, "gunzip", err
	case "deflate":
		zr, err := zlib.NewReader(bytes.NewReader(body))
		if err == nil {
			defer zr.Close()
			data, err := io.ReadAll(zr)
			return data, "inflate", err
		}
		fr := flate.NewReader(bytes.NewReader(body))
		defer fr.Close()
		data, err := io.ReadAll(fr)
		return data, "inflate-raw", err
	case "br":
		br := brotli.NewReader(bytes.NewReader(body))
		data, err := io.ReadAll(br)
		return data, "brotli", err
	case "", "identity":
		return body, "", nil
	default:
		return nil, "", fmt.Errorf("unsupported content-encoding: %s", enc)
	}
}

func splitEncodings(v string) []string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func hasChunkedEncoding(values []string) bool {
	for _, v := range values {
		for _, enc := range splitEncodings(v) {
			if strings.EqualFold(enc, "chunked") {
				return true
			}
		}
	}
	return false
}

func looksLikeChunkedBody(body []byte) bool {
	body = bytes.TrimLeft(body, "\r\n \t")
	if len(body) == 0 {
		return false
	}
	lineEnd := bytes.Index(body, []byte("\r\n"))
	if lineEnd <= 0 || lineEnd > 16 {
		return false
	}
	line := string(body[:lineEnd])
	line = strings.TrimSpace(strings.SplitN(line, ";", 2)[0])
	if line == "" {
		return false
	}
	for _, r := range line {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}

func isTextual(mediaType string, body []byte) bool {
	mediaType = strings.ToLower(mediaType)
	if strings.HasPrefix(mediaType, "text/") {
		return true
	}
	switch mediaType {
	case "application/json", "application/javascript", "application/x-javascript", "application/xml", "application/xhtml+xml", "application/x-www-form-urlencoded", "image/svg+xml":
		return true
	}
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return true
	}
	switch trimmed[0] {
	case '{', '[', '<':
		return true
	}
	return false
}

func decodeCharset(body []byte, contentType string) (string, string, error) {
	reader, err := charset.NewReader(bytes.NewReader(body), contentType)
	if err != nil {
		return string(body), "", nil
	}
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return "", "", err
	}
	return string(decoded), "decode-charset", nil
}
