package decode

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"testing"

	"github.com/andybalholm/brotli"
)

func TestLooksLikeChunkedBody(t *testing.T) {
	if !looksLikeChunkedBody([]byte("4\r\nWiki\r\n0\r\n\r\n")) {
		t.Fatal("expected chunked body detection to succeed")
	}
	if looksLikeChunkedBody([]byte("plain text body")) {
		t.Fatal("did not expect plain text to be treated as chunked")
	}
}

func TestParseResponseChunkedBrotli(t *testing.T) {
	compressed := brotliEncode([]byte("const hello = 'world';\n"))
	raw := buildChunkedResponse(compressed, "application/javascript", "br", true)
	parsed, err := ParseResponse(raw)
	if err != nil {
		t.Fatalf("ParseResponse failed: %v", err)
	}
	if parsed.BodyText != "const hello = 'world';\n" {
		t.Fatalf("unexpected decoded body: %q", parsed.BodyText)
	}
	assertHasTransform(t, parsed.Transforms, "dechunk")
	assertHasTransform(t, parsed.Transforms, "brotli")
}

func TestParseResponseDetectedChunkedWithoutHeader(t *testing.T) {
	compressed := brotliEncode([]byte("console.log('detected');\n"))
	raw := buildChunkedResponse(compressed, "application/javascript", "br", false)
	parsed, err := ParseResponse(raw)
	if err != nil {
		t.Fatalf("ParseResponse failed: %v", err)
	}
	if parsed.BodyText != "console.log('detected');\n" {
		t.Fatalf("unexpected decoded body: %q", parsed.BodyText)
	}
	if parsed.TransferEncoding == "" {
		t.Fatal("expected detected transfer encoding")
	}
	assertHasTransform(t, parsed.Transforms, "dechunk")
	assertHasTransform(t, parsed.Transforms, "brotli")
}

func TestParseResponseGzip(t *testing.T) {
	compressed := gzipEncode([]byte("{\"ok\":true}\n"))
	raw := buildIdentityResponse(compressed, "application/json", "gzip")
	parsed, err := ParseResponse(raw)
	if err != nil {
		t.Fatalf("ParseResponse failed: %v", err)
	}
	if parsed.BodyText != "{\"ok\":true}\n" {
		t.Fatalf("unexpected decoded body: %q", parsed.BodyText)
	}
	assertHasTransform(t, parsed.Transforms, "gunzip")
}

func TestParseResponseDeflate(t *testing.T) {
	compressed := zlibEncode([]byte("<xml>ok</xml>\n"))
	raw := buildIdentityResponse(compressed, "application/xml; charset=utf-8", "deflate")
	parsed, err := ParseResponse(raw)
	if err != nil {
		t.Fatalf("ParseResponse failed: %v", err)
	}
	if parsed.BodyText != "<xml>ok</xml>\n" {
		t.Fatalf("unexpected decoded body: %q", parsed.BodyText)
	}
	assertHasTransform(t, parsed.Transforms, "inflate")
}

func buildChunkedResponse(body []byte, contentType string, contentEncoding string, includeTEHeader bool) []byte {
	headers := "HTTP/1.1 200 OK\r\nContent-Type: " + contentType + "\r\n"
	if includeTEHeader {
		headers += "Transfer-Encoding: chunked\r\n"
	}
	if contentEncoding != "" {
		headers += "Content-Encoding: " + contentEncoding + "\r\n"
	}
	headers += "\r\n"
	return append([]byte(headers), encodeChunked(body)...)
}

func buildIdentityResponse(body []byte, contentType string, contentEncoding string) []byte {
	headers := "HTTP/1.1 200 OK\r\nContent-Type: " + contentType + "\r\n"
	if contentEncoding != "" {
		headers += "Content-Encoding: " + contentEncoding + "\r\n"
	}
	headers += "\r\n"
	return append([]byte(headers), body...)
}

func encodeChunked(body []byte) []byte {
	var buf bytes.Buffer
	buf.WriteString("16\r\n")
	if len(body) == 0 {
		buf.Reset()
	}
	if len(body) > 0 {
		buf.Reset()
		buf.WriteString(stringsToHexLen(len(body)))
		buf.WriteString("\r\n")
		buf.Write(body)
		buf.WriteString("\r\n0\r\n\r\n")
	}
	return buf.Bytes()
}

func stringsToHexLen(n int) string {
	const hex = "0123456789abcdef"
	if n == 0 {
		return "0"
	}
	out := make([]byte, 0)
	for n > 0 {
		out = append([]byte{hex[n%16]}, out...)
		n /= 16
	}
	return string(out)
}

func gzipEncode(body []byte) []byte {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, _ = zw.Write(body)
	_ = zw.Close()
	return buf.Bytes()
}

func zlibEncode(body []byte) []byte {
	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	_, _ = zw.Write(body)
	_ = zw.Close()
	return buf.Bytes()
}

func brotliEncode(body []byte) []byte {
	var buf bytes.Buffer
	zw := brotli.NewWriter(&buf)
	_, _ = zw.Write(body)
	_ = zw.Close()
	return buf.Bytes()
}

func assertHasTransform(t *testing.T, transforms []string, expected string) {
	t.Helper()
	for _, v := range transforms {
		if v == expected {
			return
		}
	}
	t.Fatalf("expected transform %q in %#v", expected, transforms)
}
