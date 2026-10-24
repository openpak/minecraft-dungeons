// Package signal relays the game's WebRTC signaling websocket
// (wss://signaling.minecraftservices.com:5443/ws). It is traffic-first:
// the envelope the game uses is learned from the log, so the relay is
// deliberately generic.
//
//   - Each socket gets an id: the path segment after /ws/, else an "id" or
//     "networkId" query parameter, else a random number. A JSON frame that
//     names its sender ("From", "from", "NetworkId", "networkId") adds that
//     name as an alias of the socket.
//   - A JSON object frame naming a recipient ("To" or "to") that is a live
//     socket is forwarded to it with "From" set to the sender's id.
//   - On connect, an optional template file is sent first (for the TURN
//     credential handoff once its shape is known): {{username}},
//     {{password}}, {{expires}}, {{ttl}} and {{urls}} (a JSON array) are
//     filled from a fresh TURN REST credential.
//   - Every frame is logged with ICE passwords, tokens and long runs
//     redacted.
package signal

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/openpak/minecraft-dungeons/internal/redact"
	"github.com/openpak/minecraft-dungeons/internal/turn"
	"github.com/openpak/minecraft-dungeons/internal/ws"
)

// Relay is the signaling server.
type Relay struct {
	Log        *slog.Logger
	HelloFile  string // optional template sent on connect
	TURNSecret string
	TURNURLs   []string
	TURNTTL    time.Duration

	mu    sync.Mutex
	socks map[string]*sock // id or alias -> socket
}

type sock struct {
	id    string
	c     *ws.Conn
	names []string
}

func randomID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return strconv.FormatUint(binary.BigEndian.Uint64(b)>>1, 10)
}

func (r *Relay) idFor(req *http.Request) string {
	p := strings.Trim(req.URL.Path, "/")
	if i := strings.Index(strings.ToLower(p), "ws/"); i >= 0 {
		if seg := strings.Trim(p[i+3:], "/"); seg != "" && !strings.Contains(seg, "/") {
			return seg
		}
	}
	for _, k := range []string{"id", "networkId", "NetworkId"} {
		if v := req.URL.Query().Get(k); v != "" {
			return v
		}
	}
	return randomID()
}

var icePwd = regexp.MustCompile(`(a=ice-pwd:)\S+`)

// Scrub redacts a signaling frame for the log.
func Scrub(b []byte) string {
	s := icePwd.ReplaceAllString(string(b), "${1}<redacted>")
	return redact.Body([]byte(s))
}

func (r *Relay) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	id := r.idFor(req)
	protos := strings.Split(req.Header.Get("Sec-WebSocket-Protocol"), ",")
	for i := range protos {
		protos[i] = strings.TrimSpace(protos[i])
	}
	c, err := ws.Accept(w, req, protos...)
	if err != nil {
		r.Log.Warn("signal: not a websocket", "path", redact.Text(req.URL.Path))
		return
	}
	s := &sock{id: id, c: c, names: []string{id}}
	r.mu.Lock()
	if r.socks == nil {
		r.socks = map[string]*sock{}
	}
	r.socks[id] = s
	r.mu.Unlock()
	r.Log.Info("signal connected", "id", id, "path", redact.Text(req.URL.Path), "query", redact.Text(req.URL.RawQuery),
		"headers", redact.HeaderNames(req.Header), "protocol", c.Protocol, "remote", c.RemoteAddr().String())
	stop := make(chan struct{})
	defer func() {
		close(stop)
		c.Close()
		r.mu.Lock()
		for _, n := range s.names {
			if r.socks[n] == s {
				delete(r.socks, n)
			}
		}
		r.mu.Unlock()
		r.Log.Info("signal disconnected", "id", id)
	}()
	go func() {
		t := time.NewTicker(30 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-stop:
				return
			case <-t.C:
				if c.Ping() != nil {
					return
				}
			}
		}
	}()
	if hello := r.hello(id); hello != nil {
		c.WriteText(hello)
		r.Log.Info("signal hello sent", "id", id, "frame", Scrub(hello))
	}
	for {
		op, msg, err := c.Read()
		if err != nil {
			return
		}
		r.handle(s, op, msg)
	}
}

// hello fills the template, or returns nil when there is none.
func (r *Relay) hello(id string) []byte {
	if r.HelloFile == "" {
		return nil
	}
	t, err := os.ReadFile(r.HelloFile)
	if err != nil {
		return nil
	}
	ttl := r.TURNTTL
	if ttl == 0 {
		ttl = 12 * time.Hour
	}
	cred := turn.Mint(r.TURNSecret, "dungeons-"+id, ttl, r.TURNURLs, time.Now())
	urls, _ := json.Marshal(r.TURNURLs)
	q := func(s string) string { b, _ := json.Marshal(s); return string(b[1 : len(b)-1]) }
	out := strings.NewReplacer(
		"{{username}}", q(cred.Username), "{{password}}", q(cred.Password),
		"{{expires}}", strconv.FormatInt(cred.Expires.Unix(), 10), "{{ttl}}", strconv.Itoa(int(ttl.Seconds())),
		"{{urls}}", string(urls), "{{id}}", q(id),
	).Replace(string(t))
	return []byte(strings.TrimRight(out, "\r\n"))
}

func field(m map[string]any, keys ...string) string {
	for _, k := range keys {
		switch v := m[k].(type) {
		case string:
			if v != "" {
				return v
			}
		case float64:
			return strconv.FormatFloat(v, 'f', -1, 64)
		}
	}
	return ""
}

func (r *Relay) handle(s *sock, op int, msg []byte) {
	var m map[string]any
	if op != ws.OpText || json.Unmarshal(msg, &m) != nil {
		r.Log.Info("signal frame", "from", s.id, "op", op, "json", false, "frame", Scrub(msg))
		return
	}
	if name := field(m, "From", "from", "NetworkId", "networkId"); name != "" && name != s.id {
		r.mu.Lock()
		if r.socks[name] == nil {
			r.socks[name] = s
			s.names = append(s.names, name)
		}
		r.mu.Unlock()
	}
	to := field(m, "To", "to")
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	r.mu.Lock()
	dst := r.socks[to]
	r.mu.Unlock()
	forwarded := false
	if to != "" && dst != nil && dst != s {
		m["From"] = s.id
		if _, ok := m["from"]; ok {
			m["from"] = s.id
		}
		b, _ := json.Marshal(m)
		forwarded = dst.c.WriteText(b) == nil
	}
	r.Log.Info("signal frame", "from", s.id, "to", to, "forwarded", forwarded, "keys", fmt.Sprint(keys), "frame", Scrub(msg))
}

// Live is the number of connected sockets.
func (r *Relay) Live() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	seen := map[*sock]bool{}
	for _, s := range r.socks {
		seen[s] = true
	}
	return len(seen)
}
