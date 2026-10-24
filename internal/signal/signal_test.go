package signal

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/openpak/minecraft-dungeons/internal/turn"
)

type client struct {
	c  net.Conn
	br *bufio.Reader
}

func dial(t *testing.T, srv *httptest.Server, path string) *client {
	t.Helper()
	c, err := net.Dial("tcp", strings.TrimPrefix(srv.URL, "http://"))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprintf(c, "GET %s HTTP/1.1\r\nHost: signaling.minecraftservices.com\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n"+
		"Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\nSec-WebSocket-Version: 13\r\nAuthorization: Bearer TOPSECRETTOKEN\r\n\r\n", path)
	br := bufio.NewReader(c)
	resp, err := http.ReadResponse(br, nil)
	if err != nil || resp.StatusCode != 101 {
		t.Fatalf("upgrade %v %v", resp, err)
	}
	return &client{c: c, br: br}
}

func (c *client) send(s string) {
	b := []byte(s)
	h := []byte{0x81}
	if len(b) < 126 {
		h = append(h, 0x80|byte(len(b)))
	} else {
		h = append(h, 0x80|126, byte(len(b)>>8), byte(len(b)))
	}
	m := []byte{9, 8, 7, 6}
	h = append(h, m...)
	for i := range b {
		b[i] ^= m[i%4]
	}
	c.c.Write(append(h, b...))
}

func (c *client) recv(t *testing.T) string {
	t.Helper()
	c.c.SetReadDeadline(time.Now().Add(3 * time.Second))
	for {
		var h [2]byte
		if _, err := io.ReadFull(c.br, h[:]); err != nil {
			t.Fatalf("read: %v", err)
		}
		n := int(h[1] & 0x7f)
		if n == 126 {
			var b [2]byte
			io.ReadFull(c.br, b[:])
			n = int(binary.BigEndian.Uint16(b[:]))
		}
		p := make([]byte, n)
		io.ReadFull(c.br, p)
		if h[0]&0x0f == 1 {
			return string(p)
		}
	}
}

type syncBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuf) Write(p []byte) (int, error) { s.mu.Lock(); defer s.mu.Unlock(); return s.b.Write(p) }
func (s *syncBuf) String() string              { s.mu.Lock(); defer s.mu.Unlock(); return s.b.String() }

func TestRelayHelloAndForward(t *testing.T) {
	dir := t.TempDir()
	hello := filepath.Join(dir, "hello.json")
	os.WriteFile(hello, []byte(`{"Type":2,"Message":{"Username":"{{username}}","Password":"{{password}}","Urls":{{urls}},"ExpirationInSeconds":{{ttl}}}}`), 0o600)
	logs := &syncBuf{}
	r := &Relay{Log: slog.New(slog.NewJSONHandler(logs, nil)), HelloFile: hello, TURNSecret: "coturn-secret",
		TURNURLs: []string{"turn:turn.minecraftservices.com:3478"}, TURNTTL: time.Hour}
	srv := httptest.NewServer(r)
	defer srv.Close()

	host := dial(t, srv, "/ws/1111")
	var h map[string]any
	if err := json.Unmarshal([]byte(host.recv(t)), &h); err != nil {
		t.Fatal(err)
	}
	msg := h["Message"].(map[string]any)
	if !turn.Valid("coturn-secret", msg["Username"].(string), msg["Password"].(string), time.Now()) || msg["ExpirationInSeconds"].(float64) != 3600 {
		t.Fatalf("hello credential %v", h)
	}
	join := dial(t, srv, "/ws?id=2222")
	join.recv(t) // its own hello

	offer := "CONNECTREQUEST 77 v=0\r\na=ice-ufrag:abcd\r\na=ice-pwd:VERYSECRETICEPASSWORD\r\n"
	b, _ := json.Marshal(map[string]any{"Type": 1, "To": 1111, "Message": offer})
	join.send(string(b))
	var got map[string]any
	json.Unmarshal([]byte(host.recv(t)), &got)
	if got["From"] != "2222" || got["Message"] != offer {
		t.Fatalf("forwarded %v", got)
	}
	// Reply addressed by the alias the joiner announced.
	b, _ = json.Marshal(map[string]any{"Type": 1, "To": "2222", "Message": "CONNECTRESPONSE 77 v=0"})
	host.send(string(b))
	json.Unmarshal([]byte(join.recv(t)), &got)
	if got["From"] != "1111" {
		t.Fatalf("reply %v", got)
	}
	host.send("not json")
	time.Sleep(100 * time.Millisecond)
	l := logs.String()
	for _, s := range []string{"VERYSECRETICEPASSWORD", "TOPSECRETTOKEN", msg["Password"].(string)} {
		if strings.Contains(l, s) {
			t.Fatalf("log leaks %q", s)
		}
	}
	if !strings.Contains(l, "CONNECTREQUEST 77") || !strings.Contains(l, `"forwarded":true`) {
		t.Fatalf("log misses frames: %s", l)
	}
	if r.Live() != 2 {
		t.Fatalf("live %d", r.Live())
	}
}
