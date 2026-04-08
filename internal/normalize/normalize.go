package normalize

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/hzexe/saz-tool/internal/archive"
	"github.com/hzexe/saz-tool/internal/decode"
	"github.com/hzexe/saz-tool/internal/model"
)

func Normalize(inputPath, outDir string) error {
	sessions, err := archive.ReadSessions(inputPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(outDir, "sessions"), 0o755); err != nil {
		return err
	}

	manifest := model.Manifest{
		Format:            "fiddler-saz-normalized-bundle",
		FormatVersion:     "0.1.0",
		SourceSaz:         inputPath,
		SourceSazBaseName: filepath.Base(inputPath),
		SessionCount:      len(sessions),
		NormalizationPolicy: model.Policy{
			PreserveRawFiles:         true,
			ReplaceOriginalFiles:     false,
			NormalizeTransferCoding:  true,
			NormalizeContentEncoding: true,
			DecodeCharset:            true,
			PrettyPrintJSON:          false,
			SearchLayerIncluded:      false,
			Notes: []string{
				"normalized body keeps transport-decoded text only",
				"json is not pretty-printed in canonical normalized output",
			},
		},
		Notes: []string{
			"This directory is a derived normalized export from a Fiddler SAZ archive.",
			"Fiddler session ids and ascending order are preserved explicitly.",
		},
	}

	for idx, session := range sessions {
		ordinal := idx + 1
		manifest.FiddlerSessionOrder = append(manifest.FiddlerSessionOrder, session.SessionID)

		sessionDir := filepath.Join(outDir, "sessions", fmt.Sprintf("%06d", session.SessionID))
		if err := os.MkdirAll(sessionDir, 0o755); err != nil {
			return err
		}

		requestPath := filepath.Join(sessionDir, "request.raw.txt")
		responsePath := filepath.Join(sessionDir, "response.raw.txt")
		metaPath := filepath.Join(sessionDir, "meta.json")

		meta := model.SessionMeta{
			SessionID:          session.SessionID,
			Ordinal:            ordinal,
			SourceSaz:          inputPath,
			RequestPath:        relative(outDir, requestPath),
			ResponsePath:       relative(outDir, responsePath),
			SourceRequestPath:  entryName(session.Request),
			SourceResponsePath: entryName(session.Response),
			Transforms:         []string{},
			BodyExactAfterDecode: true,
			JSONPrettyPrinted:  false,
			StatusMarkers:      []string{},
			NormalizationNotes: []string{
				"Ordinal is the ascending order by Fiddler session id found in raw/*.txt",
				"Canonical normalized body only removes transport encodings and decodes charset",
			},
		}

		summary := model.SessionSummary{
			SessionID:    session.SessionID,
			Ordinal:      ordinal,
			RequestPath:  meta.RequestPath,
			ResponsePath: meta.ResponsePath,
			MetaPath:     relative(outDir, metaPath),
		}

		if session.Request != nil {
			if err := os.WriteFile(requestPath, session.Request.Data, 0o644); err != nil {
				return err
			}
			if parsedReq, err := decode.ParseRequest(session.Request.Data); err == nil {
				meta.Method = parsedReq.Method
				meta.URL = parsedReq.URL
				summary.Method = parsedReq.Method
				summary.URL = parsedReq.URL
			}
		}

		if session.Response == nil {
			meta.BodyMissing = true
			meta.StatusMarkers = append(meta.StatusMarkers, "body-missing")
			summary.StatusMarkers = append(summary.StatusMarkers, "body-missing")
		} else {
			if err := os.WriteFile(responsePath, session.Response.Data, 0o644); err != nil {
				return err
			}
			parsedResp, err := decode.ParseResponse(session.Response.Data)
			if err != nil {
				meta.DecodeFailed = true
				meta.DecodeFailureReason = err.Error()
				meta.StatusMarkers = append(meta.StatusMarkers, "decode-failed")
				summary.StatusMarkers = append(summary.StatusMarkers, "decode-failed")
			} else {
				if parsedResp.DecodeFailed {
					meta.DecodeFailed = true
					meta.DecodeFailureReason = parsedResp.DecodeFailure
					meta.StatusMarkers = append(meta.StatusMarkers, "decode-failed")
					summary.StatusMarkers = append(summary.StatusMarkers, "decode-failed")
				}
				meta.StatusCode = parsedResp.StatusCode
				meta.ContentType = parsedResp.ContentType
				meta.TransferEncoding = parsedResp.TransferEncoding
				meta.ContentEncoding = parsedResp.ContentEncoding
				meta.Charset = parsedResp.Charset
				meta.Transforms = append(meta.Transforms, parsedResp.Transforms...)
				meta.DecodedBodyText = parsedResp.BodyIsText
				summary.StatusCode = parsedResp.StatusCode
				summary.ResponseBodyIsText = parsedResp.BodyIsText
				if len(parsedResp.BodyRaw) == 0 {
					meta.BodyMissing = true
					meta.StatusMarkers = append(meta.StatusMarkers, "body-missing")
					summary.StatusMarkers = append(summary.StatusMarkers, "body-missing")
				}
				contentLength := inferContentLength(session.Response.Data)
				if contentLength > 0 && len(parsedResp.BodyRaw) > 0 && len(parsedResp.BodyRaw) < contentLength {
					meta.BodyTruncated = true
					meta.StatusMarkers = append(meta.StatusMarkers, "body-truncated")
					summary.StatusMarkers = append(summary.StatusMarkers, "body-truncated")
				}
				if parsedResp.BodyIsText {
					decodedPath := filepath.Join(sessionDir, "response.body.decoded.txt")
					if err := os.WriteFile(decodedPath, []byte(parsedResp.BodyText), 0o644); err != nil {
						return err
					}
					meta.DecodedBodyPath = relative(outDir, decodedPath)
					summary.DecodedBodyPath = meta.DecodedBodyPath
				} else {
					meta.BinaryBodySkipped = true
					meta.StatusMarkers = append(meta.StatusMarkers, "binary-body-skipped")
					summary.StatusMarkers = append(summary.StatusMarkers, "binary-body-skipped")
				}
			}
		}

		meta.StatusMarkers = dedupeStrings(meta.StatusMarkers)
		summary.StatusMarkers = dedupeStrings(summary.StatusMarkers)

		if err := writeJSON(metaPath, meta); err != nil {
			return err
		}
		manifest.Sessions = append(manifest.Sessions, summary)
	}

	if err := os.WriteFile(filepath.Join(outDir, "README.md"), []byte(readmeText(manifest)), 0o644); err != nil {
		return err
	}
	return writeJSON(filepath.Join(outDir, "manifest.json"), manifest)
}

func relative(root, target string) string {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return target
	}
	return filepath.ToSlash(rel)
}

func entryName(e *archive.Entry) string {
	if e == nil {
		return ""
	}
	return e.Name
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func inferContentLength(raw []byte) int {
	sep := []byte("\r\n\r\n")
	idx := bytes.Index(raw, sep)
	if idx < 0 {
		sep = []byte("\n\n")
		idx = bytes.Index(raw, sep)
	}
	if idx < 0 {
		return 0
	}
	headerPart := raw[:idx]
	reader := bufio.NewReader(bytes.NewReader(append(headerPart, []byte("\r\n\r\n")...)))
	resp, err := http.ReadResponse(reader, nil)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	if resp.ContentLength <= 0 {
		return 0
	}
	return int(resp.ContentLength)
}

func dedupeStrings(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, v := range values {
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}

func readmeText(manifest model.Manifest) string {
	return fmt.Sprintf(`# %s

This directory is a normalized export derived from a Fiddler SAZ archive.

Source SAZ: %s
Session count: %d

## Invariants
- Original SAZ/raw files are preserved separately and are not replaced.
- Fiddler session ids are preserved.
- Session ordinal follows ascending Fiddler session id order recorded in manifest.json.
- response.body.decoded.txt contains only transport-normalized text:
  - dechunked when applicable
  - decompressed when applicable
  - charset-decoded when applicable
- JSON bodies are NOT pretty-printed in canonical normalized output.
- meta.json records provenance, transform steps, and status markers for each session.

## Important files
- manifest.json: bundle-level metadata, including fiddlerSessionOrder
- sessions/<id>/meta.json: per-session provenance and normalization details
- sessions/<id>/request.raw.txt: original request text from raw/*_c.txt
- sessions/<id>/response.raw.txt: original response text from raw/*_s.txt
- sessions/<id>/response.body.decoded.txt: canonical decoded text when body is textual
`, manifest.SourceSazBaseName, manifest.SourceSaz, manifest.SessionCount)
}
