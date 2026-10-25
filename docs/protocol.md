# Protocol: minecraft-dungeons

## Endpoints

- `:23520` api.minecraftservices.com stand-in: logs every request (method, path,
  query keys, header names, JSON body with credentials redacted) and answers
  from fixture files (`<method>_<path>.json`, re-read per request), else a
  default status.
- `:23521` the WebRTC signaling websocket (`wss://signaling.minecraftservices.com:5443/ws`):
  a generic relay (socket ids from the path or query, JSON frames with `To`
  forwarded with `From`), every frame logged with ICE passwords redacted, and an
  optional credential frame on connect filled from a TURN REST credential minted
  from the coturn secret (`VOICE_TURN_SECRET`).
- `:23522` `/health` and `/seen` (endpoints hit so far).

## Xbox services

The Xbox identity, friends, presence, sessions and invites the game also needs
are served by [openpak/xbl](https://github.com/openpak/xbl). Test plan:
[harbor-test-plan.md](harbor-test-plan.md). Ports 23520-23529.
