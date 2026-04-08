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
