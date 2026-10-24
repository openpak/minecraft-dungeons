// Package api is the traffic-first stand-in for api.minecraftservices.com.
// Every request is logged with its shape (method, path, query keys, header
// names, the JSON body with credentials redacted). The answer comes from a
// fixture file when one exists, else a configurable default, so the next
// response shape can be tried without a rebuild or restart.
//
// Fixture: <dir>/<method>_<path with "/" as "_">.json, e.g.
// post_dungeons_login.json for POST /dungeons/login/, holding
// {"status": 200, "headers": {"X": "y"}, "body": <any JSON>}.
package api

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/openpak/minecraft-dungeons/internal/redact"
)

// Handler serves the API host.
type Handler struct {
	Log           *slog.Logger
	FixtureDir    string
	DefaultStatus int // when no fixture matches; 0 means 404

	mu   sync.Mutex
	seen map[string]*Seen
}

// Seen counts one endpoint.
type Seen struct {
	Endpoint string    `json:"endpoint"`
	Count    int       `json:"count"`
	Last     time.Time `json:"last"`
	Status   int       `json:"status"`
}

// Fixture is a canned response.
type Fixture struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    json.RawMessage   `json:"body"`
}

// FixtureName maps a request to its fixture file name.
func FixtureName(method, path string) string {
	p := strings.Trim(strings.ToLower(path), "/")
	p = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
			return r
		}
		return '_'
	}, p)
	return strings.ToLower(method) + "_" + p + ".json"
}

func (h *Handler) fixture(method, path string) (*Fixture, error) {
	if h.FixtureDir == "" {
		return nil, os.ErrNotExist
	}
	b, err := os.ReadFile(filepath.Join(h.FixtureDir, FixtureName(method, path)))
	if err != nil {
		return nil, err
	}
	var f Fixture
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, err
	}
	if f.Status == 0 {
		f.Status = http.StatusOK
	}
	return &f, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(io.LimitReader(r.Body, 256<<10))
	f, err := h.fixture(r.Method, r.URL.Path)
	status := h.DefaultStatus
	if status == 0 {
		status = http.StatusNotFound
	}
	src := "default"
	switch {
	case err == nil:
		src = "fixture"
		status = f.Status
		for k, v := range f.Headers {
			w.Header().Set(k, v)
		}
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "application/json")
		}
		w.WriteHeader(status)
		if len(f.Body) > 0 && string(f.Body) != "null" {
			w.Write(f.Body)
		}
	case errors.Is(err, os.ErrNotExist):
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(map[string]any{"path": r.URL.Path, "errorType": "NOT_FOUND", "error": "NOT_FOUND",
			"errorMessage": "not served by OpenPak yet"})
	default:
		h.Log.Error("bad fixture", "file", FixtureName(r.Method, r.URL.Path), "err", err)
		status = http.StatusInternalServerError
		w.WriteHeader(status)
	}
	q := make([]string, 0)
	for k := range r.URL.Query() {
		q = append(q, k)
	}
	sort.Strings(q)
	attrs := []any{"method", r.Method, "host", r.Host, "path", r.URL.Path, "status", status, "answer", src,
		"query", strings.Join(q, ","), "headers", redact.HeaderNames(r.Header), "ua", r.Header.Get("User-Agent"),
		"auth_scheme", scheme(r.Header.Get("Authorization")), "fixture", FixtureName(r.Method, r.URL.Path)}
	if len(body) > 0 {
		attrs = append(attrs, "req", redact.Body(body))
	}
	h.Log.Info("api request", attrs...)
	h.mu.Lock()
	if h.seen == nil {
		h.seen = map[string]*Seen{}
	}
	k := r.Method + " " + r.URL.Path
	s := h.seen[k]
	if s == nil {
		s = &Seen{Endpoint: k}
		h.seen[k] = s
	}
	s.Count++
	s.Last = time.Now().UTC()
	s.Status = status
	h.mu.Unlock()
}

// scheme is the first word of an Authorization header, never its value.
func scheme(a string) string {
	s, _, _ := strings.Cut(strings.TrimSpace(a), " ")
	if len(s) > 16 {
		return "?"
	}
	return s
}

// Seen lists the endpoints hit so far.
func (h *Handler) Seen() []Seen {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]Seen, 0, len(h.seen))
	for _, s := range h.seen {
		out = append(out, *s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Endpoint < out[j].Endpoint })
	return out
}
