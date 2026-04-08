package normalize

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/hzexe/saz-tool/internal/model"
)

func TestDedupeStrings(t *testing.T) {
	got := dedupeStrings([]string{"a", "b", "a", "", "b", "c"})
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("unexpected length: got=%v want=%v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected dedupe result: got=%v want=%v", got, want)
		}
	}
}

func TestNormalizeStatusMarkersOnRealSample(t *testing.T) {
	input := filepath.Clean("/home/hzexe/.openclaw/workspace/tmp/public-saz/multipleSessions.saz")
	outDir := filepath.Join(t.TempDir(), "sample.norm")
	if err := Normalize(input, outDir); err != nil {
		t.Fatalf("Normalize failed: %v", err)
	}

	metaPath := filepath.Join(outDir, "sessions", "000003", "meta.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("read meta failed: %v", err)
	}
	var meta model.SessionMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatalf("unmarshal meta failed: %v", err)
	}
	if meta.DecodeFailed {
		t.Fatalf("did not expect decode failure: %#v", meta)
	}
	if meta.BinaryBodySkipped {
		t.Fatalf("did not expect binary skipped for text body: %#v", meta)
	}
}

func TestNormalizeTimelineOnRealCapturedSample(t *testing.T) {
	input := filepath.Clean("/home/hzexe/.openclaw/media/inbound/codex---3b73c252-d1bd-4e73-9767-c6bcdcfcebc4.zip")
	outDir := filepath.Join(t.TempDir(), "captured.norm")
	if err := Normalize(input, outDir); err != nil {
		t.Fatalf("Normalize failed: %v", err)
	}

	metaPath := filepath.Join(outDir, "sessions", "000027", "meta.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("read meta failed: %v", err)
	}
	var meta model.SessionMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatalf("unmarshal meta failed: %v", err)
	}
	if meta.TimelineOrdinal == 0 || meta.TimelineBegin == "" {
		t.Fatalf("expected timeline metadata to be populated: %#v", meta)
	}

	manifestPath := filepath.Join(outDir, "manifest.json")
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest failed: %v", err)
	}
	var manifest model.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("unmarshal manifest failed: %v", err)
	}
	if len(manifest.TimelineSessionOrder) == 0 {
		t.Fatalf("expected timeline session order in manifest: %#v", manifest)
	}
}
