package main

import (
	"testing"
)

// testRules returns a standard set of routing rules for use in router tests.
// Rules are ordered by Priority (ascending) to simulate pre-sorted config.
func testRules() []RoutingRule {
	return []RoutingRule{
		{Keywords: []string{"search", "news", "today", "current", "latest", "weather", "price", "stock", "who won"}, Backend: "perplexity", Priority: 1},
		{Keywords: []string{"private", "local", "offline", "sensitive", "secret", "confidential", "internal"}, Backend: "cl", Priority: 2},
		{Keywords: []string{"code", "debug", "implement", "refactor", "test", "architecture", "review", "explain"}, Backend: "cc", Priority: 3},
		{Keywords: []string{"document", "pdf", "summarize", "transcript", "paper", "report"}, Backend: "gemini", Priority: 4},
	}
}

func TestRouteByKeywordCC(t *testing.T) {
	rules := testRules()
	backend, kw := routeByKeyword("implement a function", rules)
	if backend != "cc" {
		t.Errorf("expected backend=cc, got %q", backend)
	}
	if kw != "implement" {
		t.Errorf("expected keyword=implement, got %q", kw)
	}
}

func TestRouteByKeywordPerplexity(t *testing.T) {
	rules := testRules()
	backend, kw := routeByKeyword("search for latest news", rules)
	if backend != "perplexity" {
		t.Errorf("expected backend=perplexity, got %q", backend)
	}
	if kw != "search" {
		t.Errorf("expected keyword=search, got %q", kw)
	}
}

func TestRouteByKeywordPrivacy(t *testing.T) {
	rules := testRules()
	backend, kw := routeByKeyword("private notes", rules)
	if backend != "cl" {
		t.Errorf("expected backend=cl, got %q", backend)
	}
	if kw != "private" {
		t.Errorf("expected keyword=private, got %q", kw)
	}
}

func TestRouteByKeywordGemini(t *testing.T) {
	rules := testRules()
	backend, kw := routeByKeyword("summarize this pdf", rules)
	if backend != "gemini" {
		t.Errorf("expected backend=gemini, got %q", backend)
	}
	if kw != "summarize" {
		t.Errorf("expected keyword=summarize, got %q", kw)
	}
}

func TestRouteByKeywordCaseInsensitive(t *testing.T) {
	rules := testRules()
	backend, kw := routeByKeyword("IMPLEMENT this", rules)
	if backend != "cc" {
		t.Errorf("expected backend=cc for IMPLEMENT, got %q", backend)
	}
	if kw != "implement" {
		t.Errorf("expected keyword=implement (lowercase), got %q", kw)
	}

	// Also test CODE
	backend2, kw2 := routeByKeyword("CODE review", rules)
	if backend2 != "perplexity" {
		// "review" is a cc keyword but "CODE" is also cc; "code" matches cc at priority 3
		// However "CODE review" — "review" is cc too; first match in keywords list for cc is "code"
		// but wait: perplexity has priority 1 and doesn't have "code"/"review"
		// cl has priority 2 — doesn't match
		// cc has priority 3 — "code" keyword matches "CODE review" case-insensitively
		t.Logf("backend2=%q kw2=%q", backend2, kw2)
	}
	_ = kw2
	backend3, kw3 := routeByKeyword("CODE review", rules)
	if backend3 != "cc" {
		t.Errorf("expected backend=cc for 'CODE review', got %q", backend3)
	}
	if kw3 != "code" {
		t.Errorf("expected keyword=code for 'CODE review', got %q", kw3)
	}
}

func TestRouteByKeywordNoMatch(t *testing.T) {
	rules := testRules()
	backend, kw := routeByKeyword("hello world", rules)
	if backend != "" {
		t.Errorf("expected backend=\"\" for no-match prompt, got %q", backend)
	}
	if kw != "" {
		t.Errorf("expected keyword=\"\" for no-match prompt, got %q", kw)
	}
}

func TestRouteByKeywordPriorityWins(t *testing.T) {
	// "search code" — perplexity has priority 1 (search), cc has priority 3 (code)
	// First rule evaluated is perplexity (priority 1), so perplexity wins.
	rules := testRules()
	backend, kw := routeByKeyword("search code", rules)
	if backend != "perplexity" {
		t.Errorf("expected backend=perplexity (priority 1 wins over cc priority 3), got %q", backend)
	}
	if kw != "search" {
		t.Errorf("expected keyword=search, got %q", kw)
	}
}

func TestRouteByKeywordEmptyRules(t *testing.T) {
	backend, kw := routeByKeyword("code", []RoutingRule{})
	if backend != "" {
		t.Errorf("expected backend=\"\" for empty rules, got %q", backend)
	}
	if kw != "" {
		t.Errorf("expected keyword=\"\" for empty rules, got %q", kw)
	}
}
