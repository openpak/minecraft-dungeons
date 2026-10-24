# minecraft-dungeons

Research on what an OpenPak replacement for **Minecraft Dungeons** online play
on Nintendo Switch would have to be.

**Status (2026-10-24): traffic-first service, not deployed.** `cmd/dungeonsd`
serves the game's own online services while their shapes are learned from the
game's traffic:

- `:22690` api.minecraftservices.com stand-in: logs every request (method, path,
  query keys, header names, JSON body with credentials redacted) and answers
  from fixture files (`<method>_<path>.json`, re-read per request), else a
  default status.
- `:22691` the WebRTC signaling websocket (`wss://signaling.minecraftservices.com:5443/ws`):
  a generic relay (socket ids from the path or query, JSON frames with `To`
  forwarded with `From`), every frame logged with ICE passwords redacted, and an
  optional credential frame on connect filled from a TURN REST credential minted
  from the coturn secret (`VOICE_TURN_SECRET`).
- `:22692` `/health` and `/seen` (endpoints hit so far).

The Xbox identity, friends, presence, sessions and invites the game also needs
are served by [openpak/xbl](https://github.com/openpak/xbl). Test plan:
[docs/harbor-test-plan.md](docs/harbor-test-plan.md). Ports 22690-22699.

## The finding

Minecraft Dungeons has **no game-specific online service to replace**. Since
the Cross-Platform update (1.6.0.0, November 2020) its online layer is entirely
Microsoft's: a Microsoft account, an Xbox network gamertag, an in-game Friends
tab keyed on that gamertag, and invite-only sessions between those friends.
There is no matchmaking, and gameplay is peer-hosted. Nintendo Switch Online is
required on Switch, but only as the console's permission to use the network —
the Switch friend list is not what Dungeons uses to find a session.

So the work is an **Xbox-Live-shaped identity and session adapter** —
Microsoft account sign-in, the Xbox token chain, people and presence, and the
multiplayer session directory that invites ride on. That belongs beside
`nx-baas` as a shared adapter (OpenPak's roadmap already names it `xbox-live`),
not in a repository named after one game. Minecraft Bedrock on Switch needs the
same thing.

Read [`docs/design.md`](docs/design.md) for the reasoning, the public sources,
and what is genuinely unknown.

## What would make this repository real

1. A capture (see [`docs/evidence-needed.md`](docs/evidence-needed.md), item
   E1) showing that Dungeons contacts a service of its own, beyond Microsoft
   identity and Xbox services, that gates online play. Then this repository
   implements exactly that one service.
2. The Xbox adapter existing. Dungeons then becomes an integration exercise on
   top of it — most likely a title profile and a test, not a server.

Until one of those happens this repository stays as it is. No port block is
claimed in [`ports.md`](https://docs.openpak.org/contributors/ports/); the third-party range 22600–22699 has
room reserved for this title, and taking ten ports for a service that does not
exist would be a lie in a shared document.

## Ground rules

- Clean room. Read [clean-room policy](https://docs.openpak.org/contributors/clean-room/)
  before writing a line. Public protocol documentation, our own captures, and
  compatibly licensed code are the only inputs.
- AGPL-3.0-only, English only, Go 1.25, no secrets in the repository.
- Raw captures, tokens, and account data stay outside every repository.

This project is not affiliated with or endorsed by Nintendo, Mojang, Microsoft,
or any original service provider. It contains no proprietary game code, assets,
SDK material, binaries, certificates, keys, firmware, extracted files,
credentials, or raw packet captures.
