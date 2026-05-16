package main

import (
	"net/url"
	"os"
	"strings"
)

// removeEnvVar returns env with all entries starting with key= removed.
func removeEnvVar(env []string, key string) []string {
	result := []string{}
	prefix := key + "="
	for _, e := range env {
		if !strings.HasPrefix(e, prefix) {
			result = append(result, e)
		}
	}
	return result
}

// setEnvVar returns env with key=val set (replacing existing entry if present).
func setEnvVar(env []string, key, val string) []string {
	result := removeEnvVar(env, key)
	return append(result, key+"="+val)
}

// setupEnv returns a modified copy of os.Environ() with backend's env changes applied.
// NEVER calls os.Setenv — only the returned slice is modified.
func setupEnv(backend Backend) []string {
	env := os.Environ()
	for _, key := range backend.UnsetEnv {
		env = removeEnvVar(env, key)
	}
	for key, val := range backend.SetEnv {
		env = setEnvVar(env, key, val)
	}
	return env
}

// buildCommand returns the binary path and args slice for exec.Command.
// For backends with URLTemplate, substitutes {prompt} with the URL-encoded prompt
// and returns ("xdg-open", [url]). For standard backends, returns
// (backend.Command, backend.Args + [prompt]).
func buildCommand(backend Backend, prompt string) (string, []string) {
	if backend.URLTemplate != "" {
		encoded := url.QueryEscape(prompt)
		u := strings.ReplaceAll(backend.URLTemplate, "{prompt}", encoded)
		return "xdg-open", []string{u}
	}
	args := append(append([]string{}, backend.Args...), prompt)
	return backend.Command, args
}
