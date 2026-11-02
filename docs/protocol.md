# Protocol: minecraft-dungeons

Read from the Switch executable (Minecraft Dungeons `01006C100EC08000`,
display version 1.0.0.6.0, v0 launch build 2020-04-20, SDK 7.7.0.0); recon below.
Everything below is static recon only: no capture of
the client talking to these endpoints exists yet, so frame shapes, API paths
and auth details are UNKNOWN until the game's traffic is traced. Keep this
file in step with `cmd/dungeonsd`.

## Stack

- Engine: Unreal Engine 4, Mojang custom branch (`D:/depot/Jungle/Wolf/Engine/`,
  pak v8, ~4.22-era). Custom `OnlineSubsystemDungeons` + `DungeonsLobby`;
  stock UNetDriver symbols stripped.
- Auth: XAL (`XalFindUserByLocalId`) + XSAPI (`sdk.xal.kinoko2_2020-04-08`,
  API 2020.04.0.0).
- Multiplayer activity/invites: `XblMultiplayerActivity*` plus MPSD
  (`sessiondirectory.xboxlive.com`).
- Session plane (the actual multiplayer backend): Mojang services —
  `api.minecraftservices.com` (REST), WebSocket signaling at
  `wss://signaling.minecraftservices.com:5443/ws` (RFC 6455), and the TURN
  relay at `stun:turn.minecraftservices.com:3478`.
- Peer transport: WebRTC ICE (`webrtc::Ice`, `STUN_KEEPALIVE`, relay
  credential helpers).
- TLS: libcurl 7.55.1 + OpenSSL 1.1.1e-dev over `nn::ssl`/`nsd`/`nifm`.
  No certificate-pinning code confirmed.

## What it is not

- No NEX/NPLN (zero PRUDP/Pia banner hits across all three binaries), so no
  NEX server profile applies. No Photon, no EOS, no Demonware.
- PlayFab is 3 remnant strings (`playfabapi.com`, `playfab_key`), no linked
  SDK — config leftovers, not a backend.
- Local play is same-screen co-op only. `ldn:u` is declared in the NPDM but
  no `nn::ldn`/`nn::pia` symbols exist in any binary: declared-but-unused.

## Endpoints

| Host | What | Served by |
| --- | --- | --- |
| `api.minecraftservices.com` | Service API stand-in: every request logged (method, path, query keys, header names, JSON body with credentials redacted) and answered from fixture files (`<method>_<path>.json`, re-read per request), else a default status | `dungeonsd` `:23520` |
| `signaling.minecraftservices.com:5443/ws` | WebRTC signaling websocket: a generic relay (socket ids from the path or query, JSON frames with `To` forwarded with `From`), every frame logged with ICE passwords redacted, and an optional credential frame on connect filled from a TURN REST credential minted from the coturn secret (`VOICE_TURN_SECRET`) | `dungeonsd` `:23521` |
| `turn.minecraftservices.com:3478` | STUN/TURN relay the ICE peers use | shared coturn (see [mojang-session](mojang-session.md)) |
| `*.xboxlive.com`, `login.live.com` | Xbox identity, friends, presence, sessions, invites, RTA | [`xbl`](https://github.com/openpak/xbl) |

`:23522` carries `/health` and `/seen` (endpoints hit so far). Ports 23520-23529.

## Unknown — needs a runtime trace

- The signaling frame schema (socket id placement, message types, auth).
- The TURN credential frame, if the server sends one on connect.
- Which `api.minecraftservices.com` paths the client calls, and what answers
  unblock the online menu.
- The XToken shape the client demands (see `xbl`).

Next work: runtime DNS/TLS trace from the online menu (Harbor + mitmproxy),
then UE4 pak config extraction (`DefaultEngine.ini` online-subsystem section)
and disassembly around `DungeonsLobby`/signaling connect.

## What the game uses

Minecraft Dungeons has no game-specific online service to replace. Since the Cross-Platform update
(1.6.0.0, November 2020) its online layer is entirely Microsoft's: a Microsoft account, an Xbox
network gamertag, an in-game Friends tab keyed on that gamertag, and invite-only sessions between
those friends. There is no matchmaking and gameplay is peer-hosted. Nintendo Switch Online is
required on Switch only as the console's permission to use the network; the Switch friend list is
not used to find a session. The work is therefore an Xbox-Live-shaped identity and session adapter
([`xbl`](https://github.com/openpak/xbl)), shared with Minecraft Bedrock on Switch; this repo
becomes a real service only if a capture ([evidence-needed.md](evidence-needed.md), E1) shows a
game service beyond Microsoft identity and Xbox. No port block is claimed in the
[port plan](https://docs.openpak.org/contributors/ports/) beyond 23520-23529. Reasoning:
[design.md](design.md).

## Static recon (2026-10-26)

Switch v0 dump: Minecraft Dungeons `01006C100EC08000`, display version 1.0.0.6.0, build
2020-04-20, SDK 7.7.0.0. `main` is a 55 MB NSO; NPDM name `Dungeons`. Observed in binary strings
(`main` + `subsdk0` + `sdk`):

- UE4 Mojang branch, pak v8; custom `OnlineSubsystemDungeons` + `DungeonsLobby`.
- XAL + XSAPI `kinoko2_2020-04-08`; `XblMultiplayerActivity*` + MPSD fragment.
- `api.minecraftservices.com`, `signaling.minecraftservices.com` + `:5443/ws`,
  `stun:turn.minecraftservices.com:3478`, WebRTC ICE + `STUN_KEEPALIVE`.
- Xbox hosts: `auth`/`rta`/`notify`/`stats`/`vortex-events` `.xboxlive.com`.
- Zero NEX/PRUDP/Pia, Photon, EOS or Demonware hits; 3 PlayFab remnant strings, no SDK.
  `ldn:u` is declared in the NPDM with no `nn::ldn` symbols (unused).

## Recovered values (2026-10-27, static)

- Xbox TitleId `0x67ACC3CD` = 1739551181 (built by movz/movk at `0x2f967e0`/`0x2f967e4` into the
  Xbox init struct consumed by `XblInitialize` at `0x2f96be4`; cross-checked at `0x2f89c48` in the
  Friends path). Sandbox `RETAIL`.
- SCID `00000000-0000-0000-0000-000067acc3cd` (rodata VA `0x5727685`; consumed by `XblInitialize`
  args at `0x2f96be4`, `XblMultiplayerSetActivityAsync` at `0x2f7c008`, `[FindLobbies]` at
  `0x2f781c4`). Confirm the exact wire value on an XSTS capture.
- PlayFab verified absent (no SDK linked; pak hits are allowlist and stale uplugin text only, no
  TitleId).
