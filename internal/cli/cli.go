package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/hzexe/saz-tool/internal/model"
	"github.com/hzexe/saz-tool/internal/normalize"
)

func Run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "normalize":
		return runNormalize(args[1:])
	case "show":
		return runShow(args[1:])
	case "search":
		return runSearch(args[1:])
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func runNormalize(args []string) error {
	input, outDir, err := parseNormalizeInputs(args)
	if err != nil {
		return err
	}
	if outDir == "" {
		outDir = input + ".norm"
	}
	return normalize.Normalize(input, filepath.Clean(outDir))
}

func runShow(args []string) error {
	fs := flag.NewFlagSet("show", flag.ContinueOnError)
	bodyPreview := fs.Int("body-preview", 600, "max preview chars of decoded body")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 2 {
		return errors.New("usage: saztool show <bundle_dir> <session-id> [-body-preview N]")
	}

	bundleDir := filepath.Clean(fs.Arg(0))
	sessionID, err := strconv.Atoi(fs.Arg(1))
	if err != nil {
		return fmt.Errorf("invalid session-id: %w", err)
	}

	manifest, err := readManifest(bundleDir)
	if err != nil {
		return err
	}
	var summary *model.SessionSummary
	for i := range manifest.Sessions {
		if manifest.Sessions[i].SessionID == sessionID {
			summary = &manifest.Sessions[i]
			break
		}
	}
	if summary == nil {
		return fmt.Errorf("session %d not found", sessionID)
	}

	metaPath := filepath.Join(bundleDir, filepath.FromSlash(summary.MetaPath))
	meta, err := readMeta(metaPath)
	if err != nil {
		return err
	}

	fmt.Printf("sessionId: %d\n", meta.SessionID)
	fmt.Printf("ordinal: %d\n", meta.Ordinal)
	fmt.Printf("method: %s\n", fallback(meta.Method, "-"))
	fmt.Printf("url: %s\n", fallback(meta.URL, "-"))
	fmt.Printf("statusCode: %d\n", meta.StatusCode)
	fmt.Printf("contentType: %s\n", fallback(meta.ContentType, "-"))
	fmt.Printf("transferEncoding: %s\n", fallback(meta.TransferEncoding, "-"))
	fmt.Printf("contentEncoding: %s\n", fallback(meta.ContentEncoding, "-"))
	fmt.Printf("charset: %s\n", fallback(meta.Charset, "-"))
	fmt.Printf("decodedBodyText: %t\n", meta.DecodedBodyText)
	fmt.Printf("bodyExactAfterDecode: %t\n", meta.BodyExactAfterDecode)
	fmt.Printf("jsonPrettyPrinted: %t\n", meta.JSONPrettyPrinted)
	fmt.Printf("sourceRequestPath: %s\n", fallback(meta.SourceRequestPath, "-"))
	fmt.Printf("sourceResponsePath: %s\n", fallback(meta.SourceResponsePath, "-"))
	fmt.Printf("transforms: %s\n", joinOrDash(meta.Transforms))

	if meta.DecodedBodyPath != "" {
		decodedPath := filepath.Join(bundleDir, filepath.FromSlash(meta.DecodedBodyPath))
		body, err := os.ReadFile(decodedPath)
		if err == nil {
			fmt.Println("\n--- decoded body preview ---")
			fmt.Println(truncate(string(body), *bodyPreview))
		}
	}
	return nil
}

func runSearch(args []string) error {
	bundleDir, query, bodyPreview, beforeID, afterID, err := parseSearchInputs(args)
	if err != nil {
		return err
	}

	manifest, err := readManifest(bundleDir)
	if err != nil {
		return err
	}

	type result struct {
		Summary    model.SessionSummary
		Score      int
		Where      []string
		Preview    string
		BodyLines  []int
		MetaLines  []int
	}

	results := make([]result, 0)
	for _, s := range manifest.Sessions {
		if beforeID > 0 && s.SessionID >= beforeID {
			continue
		}
		if afterID > 0 && s.SessionID <= afterID {
			continue
		}

		score := 0
		where := make([]string, 0)
		metaLineHits := make([]int, 0)
		if strings.Contains(strings.ToLower(s.URL), query) {
			score += 4
			where = append(where, "url")
		}
		if strings.Contains(strings.ToLower(s.Method), query) {
			score += 2
			where = append(where, "method")
		}
		if strings.Contains(strings.ToLower(strconv.Itoa(s.StatusCode)), query) {
			score += 1
			where = append(where, "status")
		}
		preview := ""
		bodyLineHits := make([]int, 0)
		if s.DecodedBodyPath != "" {
			decodedPath := filepath.Join(bundleDir, filepath.FromSlash(s.DecodedBodyPath))
			body, err := os.ReadFile(decodedPath)
			if err == nil {
				bodyText := string(body)
				bodyLineHits = findLineMatches(bodyText, query)
				if len(bodyLineHits) > 0 {
					score += 8
					where = append(where, "body")
					idx := strings.Index(strings.ToLower(bodyText), query)
					preview = excerpt(bodyText, idx, bodyPreview)
				}
			}
		}

		metaPath := filepath.Join(bundleDir, filepath.FromSlash(s.MetaPath))
		if metaBytes, err := os.ReadFile(metaPath); err == nil {
			metaLineHits = findLineMatches(string(metaBytes), query)
			if len(metaLineHits) > 0 {
				score += 3
				where = append(where, "meta")
			}
		}

		if score > 0 {
			results = append(results, result{
				Summary:   s,
				Score:     score,
				Where:     dedupe(where),
				Preview:   preview,
				BodyLines: bodyLineHits,
				MetaLines: metaLineHits,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].Summary.SessionID < results[j].Summary.SessionID
		}
		return results[i].Score > results[j].Score
	})

	if len(results) == 0 {
		fmt.Println("no matches")
		return nil
	}

	for _, r := range results {
		fmt.Printf("session=%d ordinal=%d score=%d method=%s status=%d\n", r.Summary.SessionID, r.Summary.Ordinal, r.Score, fallback(r.Summary.Method, "-"), r.Summary.StatusCode)
		fmt.Printf("url=%s\n", fallback(r.Summary.URL, "-"))
		fmt.Printf("where=%s\n", joinOrDash(r.Where))
		if len(r.BodyLines) > 0 {
			fmt.Printf("body_lines=%s\n", intsToCSV(r.BodyLines))
		}
		if len(r.MetaLines) > 0 {
			fmt.Printf("meta_lines=%s\n", intsToCSV(r.MetaLines))
		}
		if r.Preview != "" {
			fmt.Printf("preview=%s\n", r.Preview)
		}
		fmt.Println()
	}
	return nil
}

func printUsage() {
	fmt.Println(`saztool - Fiddler SAZ normalize/search helper

Commands:
  normalize <file.saz> [-out dir]        Export normalized bundle
  show <bundle> <session-id>             Show one session summary
  search <bundle> <query>                Search normalized text and metadata
`)
}

func parseNormalizeInputs(args []string) (input string, outDir string, err error) {
	remaining := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-out" || arg == "--out":
			if i+1 >= len(args) {
				err = errors.New("-out requires a value")
				return
			}
			outDir = args[i+1]
			i++
		case strings.HasPrefix(arg, "-out="):
			outDir = strings.TrimPrefix(arg, "-out=")
		case strings.HasPrefix(arg, "--out="):
			outDir = strings.TrimPrefix(arg, "--out=")
		default:
			remaining = append(remaining, arg)
		}
	}
	if len(remaining) != 1 {
		err = errors.New("usage: saztool normalize <file.saz> [-out output_dir]")
		return
	}
	input = remaining[0]
	return
}

func parseSearchInputs(args []string) (bundleDir string, query string, bodyPreview int, beforeID int, afterID int, err error) {
	bodyPreview = 160
	remaining := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--body-preview":
			if i+1 >= len(args) {
				err = errors.New("--body-preview requires a value")
				return
			}
			bodyPreview, err = strconv.Atoi(args[i+1])
			if err != nil {
				err = fmt.Errorf("invalid --body-preview value: %w", err)
				return
			}
			i++
		case strings.HasPrefix(arg, "--body-preview="):
			bodyPreview, err = strconv.Atoi(strings.TrimPrefix(arg, "--body-preview="))
			if err != nil {
				err = fmt.Errorf("invalid --body-preview value: %w", err)
				return
			}
		case arg == "--before-id":
			if i+1 >= len(args) {
				err = errors.New("--before-id requires a value")
				return
			}
			beforeID, err = strconv.Atoi(args[i+1])
			if err != nil {
				err = fmt.Errorf("invalid --before-id value: %w", err)
				return
			}
			i++
		case strings.HasPrefix(arg, "--before-id="):
			beforeID, err = strconv.Atoi(strings.TrimPrefix(arg, "--before-id="))
			if err != nil {
				err = fmt.Errorf("invalid --before-id value: %w", err)
				return
			}
		case arg == "--after-id":
			if i+1 >= len(args) {
				err = errors.New("--after-id requires a value")
				return
			}
			afterID, err = strconv.Atoi(args[i+1])
			if err != nil {
				err = fmt.Errorf("invalid --after-id value: %w", err)
				return
			}
			i++
		case strings.HasPrefix(arg, "--after-id="):
			afterID, err = strconv.Atoi(strings.TrimPrefix(arg, "--after-id="))
			if err != nil {
				err = fmt.Errorf("invalid --after-id value: %w", err)
				return
			}
		default:
			remaining = append(remaining, arg)
		}
	}

	if len(remaining) < 2 {
		err = errors.New("usage: saztool search <bundle_dir> <query> [--body-preview N] [--before-id N] [--after-id N]")
		return
	}

	bundleDir = filepath.Clean(remaining[0])
	query = strings.TrimSpace(strings.Join(remaining[1:], " "))
	if query == "" {
		err = errors.New("query cannot be empty")
		return
	}
	if beforeID > 0 && afterID > 0 && afterID >= beforeID {
		err = errors.New("after-id must be less than before-id when both are set")
		return
	}
	return
}

func readManifest(bundleDir string) (*model.Manifest, error) {
	path := filepath.Join(bundleDir, "manifest.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	var manifest model.Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	return &manifest, nil
}

func readMeta(path string) (*model.SessionMeta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read meta: %w", err)
	}
	var meta model.SessionMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("parse meta: %w", err)
	}
	return &meta, nil
}

func fallback(v, d string) string {
	if strings.TrimSpace(v) == "" {
		return d
	}
	return v
}

func joinOrDash(v []string) string {
	if len(v) == 0 {
		return "-"
	}
	return strings.Join(v, ", ")
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if n <= 0 || len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func excerpt(s string, idx, n int) string {
	if n <= 0 {
		return ""
	}
	start := idx - n/3
	if start < 0 {
		start = 0
	}
	end := start + n
	if end > len(s) {
		end = len(s)
	}
	piece := strings.TrimSpace(s[start:end])
	if start > 0 {
		piece = "..." + piece
	}
	if end < len(s) {
		piece = piece + "..."
	}
	return strings.ReplaceAll(piece, "\n", " ")
}

func dedupe(v []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(v))
	for _, s := range v {
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

func findLineMatches(s, query string) []int {
	query = strings.ToLower(query)
	if query == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	out := make([]int, 0)
	for i, line := range lines {
		if strings.Contains(strings.ToLower(line), query) {
			out = append(out, i+1)
		}
	}
	return out
}

func intsToCSV(v []int) string {
	parts := make([]string, 0, len(v))
	for _, n := range v {
		parts = append(parts, strconv.Itoa(n))
	}
	return strings.Join(parts, ",")
}
