package api

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFixtureDefaultAndRedaction(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, FixtureName("POST", "/dungeons/login/")),
		[]byte(`{"status":200,"headers":{"X-Test":"1"},"body":{"accessToken":"abc"}}`), 0o600)
	var logs bytes.Buffer
	h := &Handler{Log: slog.New(slog.NewJSONHandler(&logs, nil)), FixtureDir: dir}
	srv := httptest.NewServer(h)
	defer srv.Close()

	req, _ := http.NewRequest("POST", srv.URL+"/dungeons/login/", strings.NewReader(`{"identityToken":"XBL3.0 x=1;SECRETSECRET","platform":"SWITCH"}`))
	req.Header.Set("Authorization", "Bearer SUPERSECRETVALUE")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 || resp.Header.Get("X-Test") != "1" || !strings.Contains(string(b), "accessToken") {
		t.Fatalf("fixture answer %d %s", resp.StatusCode, b)
	}
	resp, _ = http.Get(srv.URL + "/dungeons/config/game?build=1")
	if resp.StatusCode != 404 {
		t.Fatalf("default %d", resp.StatusCode)
	}
	l := logs.String()
	for _, secret := range []string{"SECRETSECRET", "SUPERSECRETVALUE"} {
		if strings.Contains(l, secret) {
			t.Fatalf("log leaks %s: %s", secret, l)
		}
	}
	if !strings.Contains(l, `SWITCH`) || !strings.Contains(l, `"auth_scheme":"Bearer"`) || !strings.Contains(l, `"query":"build"`) {
		t.Fatalf("log misses shape: %s", l)
	}
	if s := h.Seen(); len(s) != 2 {
		t.Fatalf("seen %v", s)
	}
}

func TestFixtureName(t *testing.T) {
	if n := FixtureName("GET", "/dungeons/config/tower/mobgroups"); n != "get_dungeons_config_tower_mobgroups.json" {
		t.Fatal(n)
	}
}
