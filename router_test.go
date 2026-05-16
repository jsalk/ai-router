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
	// "summarize this pdf" — gemini rule keywords are [document, pdf, summarize, ...].
	// "pdf" appears before "summarize" in the keyword list, so "pdf" is the matched keyword.
	backend, kw := routeByKeyword("summarize this pdf", rules)
	if backend != "gemini" {
		t.Errorf("expected backend=gemini, got %q", backend)
	}
	if kw != "pdf" {
		t.Errorf("expected keyword=pdf (first match in gemini keyword list), got %q", kw)
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

	// Also test CODE — cc rule has priority 3; perplexity/cl rules don't contain "code"/"review".
	// "code" keyword in the cc rule matches "CODE review" case-insensitively.
	backend2, kw2 := routeByKeyword("CODE review", rules)
	if backend2 != "cc" {
		t.Errorf("expected backend=cc for 'CODE review', got %q", backend2)
	}
	if kw2 != "code" {
		t.Errorf("expected keyword=code for 'CODE review', got %q", kw2)
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
