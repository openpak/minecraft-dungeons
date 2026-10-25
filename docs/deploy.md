# Deploy: minecraft-dungeons

## Build

The Dockerfile builds `cmd/dungeonsd` statically (`CGO_ENABLED=0`, `-trimpath`) and
copies it into `gcr.io/distroless/static`, running as `nonroot:nonroot`. It exposes
23520, 23521 and 23522, and its healthcheck is `/dungeonsd healthcheck`, which
queries `DUNGEONS_HEALTH_ADDR` on 127.0.0.1.

Build from the repository root, and run the tests with a standalone workspace:

```sh
GOWORK=off go test -race ./...
go build -o dungeonsd ./cmd/dungeonsd
```

## Ports and environment

| Variable | Default | Meaning |
| --- | --- | --- |
| `DUNGEONS_API_ADDR` | `:23520` | API stand-in listener |
| `DUNGEONS_SIGNAL_ADDR` | `:23521` | WebRTC signaling websocket listener |
| `DUNGEONS_HEALTH_ADDR` | `:23522` | `/health` and `/seen` |
| `DUNGEONS_FIXTURES` | `/app/fixtures` | Response fixtures (`<method>_<path>.json`), re-read per request |
| `DUNGEONS_DEFAULT_STATUS` | unset (404) | Status when no fixture matches; 0 means 404 |
| `DUNGEONS_SIGNAL_HELLO` | unset | Optional template sent on connect |
| `DUNGEONS_TURN_URLS` | `stun:turn.minecraftservices.com:3478,turn:turn.minecraftservices.com:3478` | TURN URLs in the credential frame |
| `DUNGEONS_TURN_TTL` | `12h` | Lifetime of the TURN REST credential |
| `DUNGEONS_TURN_SECRET` | falls back to `VOICE_TURN_SECRET` | coturn shared secret; without it the server logs a warning and coturn will not accept the credentials |

## Quadlet and Traefik

`deploy/openpak-minecraft-dungeons.container` is an example quadlet. It runs the image
on `openpak.network`, mounts the fixtures directory read-only, reads
`%h/openpak/deploy/.env`, sets the three listener ports above, and uses the
`dungeonsd healthcheck` command for health.

Traefik terminates TLS for `api.minecraftservices.com` (console entrypoint, forwarded to
`:23520`) and for `signaling.minecraftservices.com` on a 5443 entrypoint (forwarded to
`:23521`, websocket). The service itself speaks plain HTTP inside the pod network.
