package main

import "strings"

// routeByKeyword returns (backendName, matchedKeyword).
// Rules are evaluated in order (caller must pre-sort by Priority if needed).
// Returns ("", "") if no rule matches.
func routeByKeyword(prompt string, rules []RoutingRule) (string, string) {
	promptLower := strings.ToLower(prompt)
	for _, rule := range rules {
		for _, kw := range rule.Keywords {
			if strings.Contains(promptLower, strings.ToLower(kw)) {
				return rule.Backend, kw
			}
		}
	}
	return "", ""
}
