// Package redact strips credentials from what the services log: header
// values are never logged, JSON values under credential-looking keys are
// replaced, and long token-like runs anywhere are elided.
package redact

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"
)

var secretKey = regexp.MustCompile(`(?i)token|ticket|password|secret|signature|proof|refresh|credential|authorization|cookie|session|key|identity`)

var longRun = regexp.MustCompile(`[A-Za-z0-9_\-.=+/]{40,}`)

// Limit is the largest logged body.
const Limit = 8192

// Body returns b fit for a log line.
func Body(b []byte) string {
	var v any
	if json.Unmarshal(b, &v) != nil {
		return clip(longRun.ReplaceAllString(string(b), "<redacted>"))
	}
	out, _ := json.Marshal(value(v, ""))
	return clip(string(out))
}

// Text redacts a non-JSON string (an SDP line, a query string).
func Text(s string) string { return clip(longRun.ReplaceAllString(s, "<redacted>")) }

func clip(s string) string {
	if len(s) > Limit {
		return s[:Limit] + "..."
	}
	return s
}

func value(v any, key string) any {
	switch t := v.(type) {
	case map[string]any:
		for k, x := range t {
			t[k] = value(x, k)
		}
		return t
	case []any:
		for i, x := range t {
			t[i] = value(x, key)
		}
		return t
	case string:
		if key != "" && secretKey.MatchString(key) {
			return "<redacted>"
		}
		// A JSON document inside a string (signaling payloads) is redacted too.
		if strings.HasPrefix(strings.TrimSpace(t), "{") {
			var inner any
			if json.Unmarshal([]byte(t), &inner) == nil {
				b, _ := json.Marshal(value(inner, ""))
				return string(b)
			}
		}
		return longRun.ReplaceAllString(t, "<redacted>")
	}
	return v
}

// HeaderNames lists header names only, sorted.
func HeaderNames(h map[string][]string) string {
	n := make([]string, 0, len(h))
	for k := range h {
		n = append(n, k)
	}
	sort.Strings(n)
	return strings.Join(n, ",")
}
