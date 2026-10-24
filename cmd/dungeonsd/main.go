// Command dungeonsd serves Minecraft Dungeons' own online services for
// OpenPak, traffic-first: api.minecraftservices.com (logged, answered from
// fixtures) and the WebRTC signaling websocket with TURN credentials from
// the shared coturn secret.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/openpak/minecraft-dungeons/internal/api"
	sig "github.com/openpak/minecraft-dungeons/internal/signal"
)

var version = "dev"

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		addr := env("DUNGEONS_HEALTH_ADDR", ":23522")
		if strings.HasPrefix(addr, ":") {
			addr = "127.0.0.1" + addr
		}
		c := http.Client{Timeout: 3 * time.Second}
		r, err := c.Get("http://" + addr + "/health")
		if err != nil || r.StatusCode != 200 {
			os.Exit(1)
		}
		return
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(log); err != nil {
		log.Error("dungeonsd", "err", err)
		os.Exit(1)
	}
}

// Handlers builds the two service handlers from the environment.
func Handlers(log *slog.Logger) (*api.Handler, *sig.Relay) {
	status, _ := strconv.Atoi(os.Getenv("DUNGEONS_DEFAULT_STATUS"))
	a := &api.Handler{Log: log, FixtureDir: env("DUNGEONS_FIXTURES", "/app/fixtures"), DefaultStatus: status}
	var urls []string
	for _, u := range strings.Split(env("DUNGEONS_TURN_URLS", "stun:turn.minecraftservices.com:3478,turn:turn.minecraftservices.com:3478"), ",") {
		if u = strings.TrimSpace(u); u != "" {
			urls = append(urls, u)
		}
	}
	ttl, _ := time.ParseDuration(env("DUNGEONS_TURN_TTL", "12h"))
	r := &sig.Relay{Log: log, HelloFile: os.Getenv("DUNGEONS_SIGNAL_HELLO"), TURNSecret: env("DUNGEONS_TURN_SECRET", os.Getenv("VOICE_TURN_SECRET")),
		TURNURLs: urls, TURNTTL: ttl}
	return a, r
}

func run(log *slog.Logger) error {
	a, r := Handlers(log)
	if r.TURNSecret == "" {
		log.Warn("no TURN secret: credentials in the hello template will not be accepted by coturn")
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	errc := make(chan error, 3)
	var servers []*http.Server
	listen := func(name, addr string, h http.Handler) error {
		l, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}
		s := &http.Server{Handler: h, ReadHeaderTimeout: 15 * time.Second}
		servers = append(servers, s)
		log.Info("listening", "listener", name, "addr", addr, "version", version)
		go func() {
			if err := s.Serve(l); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errc <- err
			}
		}()
		return nil
	}
	if err := listen("api", env("DUNGEONS_API_ADDR", ":23520"), a); err != nil {
		return err
	}
	if err := listen("signal", env("DUNGEONS_SIGNAL_ADDR", ":23521"), r); err != nil {
		return err
	}
	hm := http.NewServeMux()
	hm.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok\n")) })
	hm.HandleFunc("/seen", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"api": a.Seen(), "signal_live": r.Live()})
	})
	if err := listen("health", env("DUNGEONS_HEALTH_ADDR", ":23522"), hm); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
	case err := <-errc:
		return err
	}
	sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, s := range servers {
		s.Shutdown(sctx)
	}
	return nil
}
