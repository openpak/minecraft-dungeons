# Minecraft Dungeons

OpenPak's Minecraft Dungeons Server.

| | |
| --- | --- |
| Game | [Minecraft Dungeons](https://openpak.org/titles#t-22) |
| Networks | Nintendo Switch |
| Title ID | `01006C100EC08000` |
| Licence | AGPL-3.0-only |

## Networks

| Network | Reached through | Also needs |
| --- | --- | --- |
| Nintendo Switch | Traefik (TLS for the two hosts below) | [`xbl`](https://github.com/openpak/xbl) |

## What it serves

| Host | What | Served by |
| --- | --- | --- |
| `api.minecraftservices.com` | Service API stand-in: logged, answered from fixtures | `dungeonsd` |
| `signaling.minecraftservices.com` | WebRTC signaling websocket with TURN credentials | `dungeonsd` |

## Run

```sh
go run ./cmd/dungeonsd
```

| Variable | Default | Meaning |
| --- | --- | --- |
| `DUNGEONS_API_ADDR` | `:23520` | API listener |
| `DUNGEONS_SIGNAL_ADDR` | `:23521` | Signaling websocket listener |
| `DUNGEONS_HEALTH_ADDR` | `:23522` | `/health` and `/seen` |
| `VOICE_TURN_SECRET` | unset | coturn secret for TURN REST credentials |

## Docs

- [Protocol](docs/protocol.md): endpoints, hosts, the Xbox services split
- [Mojang session](docs/mojang-session.md): signaling + TURN plan
- [Evidence](docs/evidence.md): captures, traces and what is confirmed
- [Deploy](docs/deploy.md): build, ports, environment, quadlet and Traefik
- [Notes](docs/notes.md): scope, status, design rationale, ground rules
- [Design](docs/design.md): network stack, feasibility and scope study
- [Evidence needed](docs/evidence-needed.md): captures still required
- [Harbor test plan](docs/harbor-test-plan.md): test runs against the game's traffic

Live compatibility status: [openpak.org/titles#t-22](https://openpak.org/titles#t-22).

## Licence

AGPL-3.0-only. See [LICENSE](LICENSE).
