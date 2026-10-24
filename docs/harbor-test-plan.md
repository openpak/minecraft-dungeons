# Harbor test plan (traffic-first)

Shapes are learned from what the game sends to dungeonsd and xbl, not from the
game's code. Each run changes one thing; the decision rule is written before
the run. Two @openpak.org test accounts, Harbor `--netns` households.

| Run | Change | Watch | Decision |
| --- | --- | --- | --- |
| 1 | none (every API path 404) | xbl sign-in log through SISU/XSTS; `api request` lines; game screen | the ordered endpoint list, request bodies (redacted) and which call the game stops on |
| 2 | fixture for the call it stopped on, minimal shape from the request it made | same | repeat until the main menu shows the player signed in online |
| E1 | `/entitlements/nintendo`: no fixture (404) | does online play (host, invite, join) still work | works: entitlements are not an online gate |
| E2 | `/entitlements/nintendo`: 200 with an empty list, no signature | online play, DLC/cosmetics | online blocked only here = the signature is not checked or not fatal |
| E3 | 200 with a list and a wrong signature | online play, DLC/cosmetics | online blocked = signed entitlements gate online play: stop, report to the owner as a decision (no game patches without approval) |
| S1 | host + join, relay generic | `signal connected` path/query/protocol, `signal frame` keys and `forwarded` | the envelope (id scheme, To/From, message types) |
| S2 | `DUNGEONS_SIGNAL_HELLO` template if S1 shows the client waiting for credentials | ICE candidates of type relay in the frames, coturn log | TURN works |
| M | full mission, two consoles | both finish | done |

Fixture files are re-read per request: edit them on the box while the game
runs (`deploy/minecraft-dungeons/fixtures/<method>_<path>.json`).
