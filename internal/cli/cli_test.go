package cli

import "testing"

func TestParseSearchScopesDefault(t *testing.T) {
	scopes, err := parseSearchScopes("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !scopes["body"] || !scopes["meta"] {
		t.Fatalf("default scopes should include body and meta: %#v", scopes)
	}
	if scopes["request"] || scopes["response"] {
		t.Fatalf("default scopes should not include request/response: %#v", scopes)
	}
}

func TestParseSearchScopesAll(t *testing.T) {
	scopes, err := parseSearchScopes("all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, key := range []string{"body", "meta", "request", "response"} {
		if !scopes[key] {
			t.Fatalf("scope %s should be enabled in all: %#v", key, scopes)
		}
	}
}

func TestParseSearchScopesSubset(t *testing.T) {
	scopes, err := parseSearchScopes("request,response")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !scopes["request"] || !scopes["response"] {
		t.Fatalf("request/response should be enabled: %#v", scopes)
	}
	if scopes["body"] || scopes["meta"] {
		t.Fatalf("body/meta should be disabled: %#v", scopes)
	}
}

func TestParseSearchScopesInvalid(t *testing.T) {
	if _, err := parseSearchScopes("nope"); err == nil {
		t.Fatal("expected invalid scope error")
	}
}

func TestParseNormalizeInputsBothOrders(t *testing.T) {
	input, outDir, err := parseNormalizeInputs([]string{"file.saz", "-out", "outdir"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if input != "file.saz" || outDir != "outdir" {
		t.Fatalf("unexpected normalize parse result: input=%q out=%q", input, outDir)
	}

	input, outDir, err = parseNormalizeInputs([]string{"-out", "outdir", "file.saz"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if input != "file.saz" || outDir != "outdir" {
		t.Fatalf("unexpected normalize parse result: input=%q out=%q", input, outDir)
	}
}

func TestParseSearchInputsOptionOrder(t *testing.T) {
	bundle, query, _, beforeID, afterID, scopes, contextLines, err := parseSearchInputs([]string{"bundle.norm", "token", "--before-id", "20", "--after-id", "10", "--in", "request", "-C", "2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bundle != "bundle.norm" || query != "token" || beforeID != 20 || afterID != 10 || contextLines != 2 {
		t.Fatalf("unexpected parse result: bundle=%q query=%q before=%d after=%d context=%d", bundle, query, beforeID, afterID, contextLines)
	}
	if !scopes["request"] || scopes["body"] {
		t.Fatalf("unexpected scopes: %#v", scopes)
	}

	bundle, query, _, beforeID, afterID, scopes, contextLines, err = parseSearchInputs([]string{"--in", "response", "--after-id", "10", "bundle.norm", "token", "--before-id", "20"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bundle != "bundle.norm" || query != "token" || beforeID != 20 || afterID != 10 || contextLines != 0 {
		t.Fatalf("unexpected parse result: bundle=%q query=%q before=%d after=%d context=%d", bundle, query, beforeID, afterID, contextLines)
	}
	if !scopes["response"] || scopes["meta"] {
		t.Fatalf("unexpected scopes: %#v", scopes)
	}
}
