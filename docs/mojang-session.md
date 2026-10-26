# Mojang session plane: signaling + TURN plan

How `dungeonsd` stands in for the Mojang-hosted half of Minecraft Dungeons
multiplayer on Switch (`01006C100EC08000`). Recon facts only — frame shapes
and API paths are UNKNOWN until the game's traffic is traced (see
[protocol](protocol.md)).

## What the client expects

- **Signaling**: a WebSocket at `signaling.minecraftservices.com:5443/ws`
  (raw strings in `main`, RFC 6455 accept GUID present). Opened by
  `OnlineSubsystemDungeons`/`DungeonsLobby`; the session coordination channel.
- **TURN**: `stun:turn.minecraftservices.com:3478` (raw string in `main`,
  next to `RelayCredentials`/`RelayServ` helpers). Peers connect over WebRTC
  ICE with STUN keepalives; the relay carries media/session traffic the P2P
  path cannot.
- **API**: `api.minecraftservices.com` over HTTPS (libcurl 7.55.1 +
  OpenSSL 1.1.1e-dev). Paths UNKNOWN.

## Plan

1. **Signaling relay** (`:23521`, implemented, generic). Accept the `/ws`
   upgrade, take the socket id from the path or query, forward JSON frames
   carrying `To` with `From` attached, log every frame with ICE passwords
   redacted. No Dungeons-specific frame knowledge yet — the relay learns it
   from traffic, and fixtures pin it down once known.
2. **Credential frame** (template-ready). If the trace shows the server greeting
   a new socket with TURN credentials, serve it from `DUNGEONS_SIGNAL_HELLO`
   with a TURN REST credential minted from the shared coturn secret
   (`VOICE_TURN_SECRET`, 12 h TTL by default). Until the frame is observed the
   template stays unset and connect sends nothing extra.
3. **TURN** (shared coturn, not this repo). Advertise
   `stun:turn.minecraftservices.com:3478` / `turn:turn.minecraftservices.com:3478`
   (override with `DUNGEONS_TURN_URLS`) backed by the deployment's coturn
   instance, so ICE connectivity checks and relayed traffic terminate inside
   OpenPak. `dungeonsd` only mints the credentials the relay hands out.
4. **API stand-in** (`:23520`, implemented). Log-and-fixture: every request is
   logged with credentials redacted and answered from
   `<method>_<path>.json` fixtures, so the trace doubles as the fixture set
   for the next run.

## What the first trace must capture

In order: the auth posture on `/ws` connect (headers, query, first frames),
the full auth→signaling→TURN sequence from the online menu, the credential
frame if any, and the `api.minecraftservices.com` paths hit before the game
considers itself online. Runs: [harbor-test-plan](harbor-test-plan.md).

## Open questions

- Does `/ws` require an Xbox token (header/query/first frame), or is socket
  identity enough?
- Is there a server-sent hello with TURN credentials, or does the client
  derive them?
- Which ICE candidates the client gathers (host/srflx/relay) and whether it
  falls back to relay-only behind symmetric NAT.
