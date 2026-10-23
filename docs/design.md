# Minecraft Dungeons on Switch — network stack, feasibility, scope

> **Current Switch client (2026-10-13): Harbor**, at
> `~/REPOS/Openpak/emulators/harbor`. Harbor is an independent
> OpenPak project based on Eden. Citron, Ryujinx (desktop and Android), and
> the former OpenPak Eden checkouts have been removed. References to those
> forks below record historical evidence only: do not run their commands,
> reinstall them, or select them for current work. Use Harbor's README and
> current handoff for builds and testing; old results do not establish Harbor parity.

Status: research only, 2026-09-09. No code exists and none is justified yet.
Everything here is from public sources or reasoned from them; nothing in this
document comes from a capture, because no capture has been taken. Claims are
labelled **PUBLIC** (cited), **INFERRED** (reasoned from a cited fact, not
observed) or **UNKNOWN** (needs the matching item in `evidence-needed.md`).

## Verdict

**Not feasible now, and not as a per-title server.** This repository stays
documentation-only.

Minecraft Dungeons has no game-specific online service to replace. Since the
Cross-Platform update its entire online layer — identity, friends, invites,
sessions — is Microsoft's: a Microsoft account, an Xbox network gamertag, and
the in-game Friends tab keyed on that gamertag. There is no matchmaking
service to stand up and no lobby server to imitate. Play is invite-only
between Microsoft friends, and the gameplay itself is peer-hosted.

So the work that would make Dungeons play online against OpenPak is:

1. an **Xbox-Live-shaped identity and session adapter** — Microsoft account
   sign-in, Xbox token chain, people/presence, and the session directory the
   invites ride on; and
2. **whatever Dungeons does on top of it**, which is currently UNKNOWN and
   cannot be known without a capture.

Item 1 is not a Minecraft Dungeons problem. It is the same problem for
Minecraft Bedrock on Switch (`servers/minecraft`), for every other
Microsoft-published cross-play title, and it is already on OpenPak's roadmap
as the `xbox-live` component. Building it inside this repository would be the
wrong place, and building a skeleton here before it exists would be building
against an interface nobody has seen.

Item 2 might turn out to be nothing at all — in which case this repository is
permanently a stub and Dungeons support is a line item in the Xbox adapter.

**Recommendation: park this repository until the Xbox adapter exists or until
E1–E3 in `evidence-needed.md` are taken. Do not claim a port block.**

## What the Switch build talks to

### Established from public sources

- **Online play requires a Microsoft account, on every platform including
  Switch.** Nintendo Switch players sign in with a Microsoft account from the
  Dungeons home screen; the game shows a verification code that the player
  enters on another device. (PUBLIC — [Windows Central cross-play
  guide](https://www.windowscentral.com/minecraft-dungeons-guide-cross-play),
  [ACHIVX walkthrough of the Switch
  link](https://achivx.com/how-to-link-your-microsoft-account-to-minecraft-dungeons-nintendo-switch/))
- **The social graph is Microsoft's, not Nintendo's.** "Crossplay is tied to
  your Microsoft account rather than your console's friends list… You add
  friends with their Microsoft Gamertag… Console usernames (PSN IDs, Switch
  friend codes) are not used for crossplay." (PUBLIC —
  [GHOSTCAP](https://www.ghostcap.com/is-minecraft-dungeons-cross-platform))
  This is the single most important fact in this document: the Switch friend
  list, and therefore everything `nx-baas` and OpenPak's own friends layer
  provide, is not what Dungeons uses to find a session.
- **There is no matchmaking.** Sessions are invite-only, up to four players,
  drop-in/drop-out, one player hosting. (PUBLIC —
  [game8](https://game8.co/games/Minecraft-Dungeons/archives/289091),
  [PC Gamer](https://www.pcgamer.com/minecraft-dungeons-crossplay-multiplayer-pc-coop/))
  There is no server-side player pool to reproduce.
- **Cross-play arrived in the free Cross-Platform update (1.6.0.0, November
  2020)** and is still live in 2026. (PUBLIC — GHOSTCAP, above)
- **NSO is required on Switch, in addition to the Microsoft account.**
  (PUBLIC — GHOSTCAP, above.) NSO is the console's permission to use the
  network at all; OpenPak already satisfies it through the Switch adapter's
  NSO membership handling. It is not the game's identity.
- **Microsoft's console sign-in is the OAuth 2.0 remote-connect (device code)
  flow** against `https://login.live.com/oauth20_remoteconnect.srf` — the
  console displays a code, the player signs in on a phone or PC, the console
  polls for the token. This is the documented mechanism behind the same
  code-entry screen Minecraft Bedrock shows on Switch. (PUBLIC —
  [Minecraft Wiki, Microsoft
  authentication](https://minecraft.wiki/w/Microsoft_authentication);
  [Xbox Support, Minecraft for Nintendo Switch
  FAQ](https://support.xbox.com/en-CA/help/games-apps/game-titles/minecraft-switch-edition-faq))
- **The Xbox client library used for cross-platform titles is public and MIT
  licensed.** `microsoft/xbox-live-api` (XSAPI) is the C/C++ library titles
  use for social, presence, multiplayer sessions, achievements and real-time
  activity, and Microsoft shipped it explicitly so an iOS, Android, **Nintendo
  Switch**, Xbox or PC player could play together. (PUBLIC —
  [microsoft/xbox-live-api](https://github.com/microsoft/xbox-live-api), MIT,
  still maintained as of 2026-08;
  [MicrosoftDocs/xbox-live-docs](https://github.com/MicrosoftDocs/xbox-live-docs))
  Its request shapes and hostnames are readable from that repository and from
  the Learn documentation without touching anything proprietary.
- **The Xbox token chain is documented publicly**: MSA token →
  `user.auth.xboxlive.com` (user token) → `xsts.auth.xboxlive.com` (XSTS
  token) → XSTS token plus user hash in the `Authorization` header of every
  Xbox services call. (PUBLIC — [Microsoft Learn, Xbox services
  authentication](https://learn.microsoft.com/en-us/gaming/gdk/docs/services/fundamentals/s2s-auth-calls/service-authentication/live-xbox-live-authentication),
  [wiki.vg Microsoft Authentication
  Scheme](https://wiki.vg/Microsoft_Authentication_Scheme))
- **Sessions are brokered through the Multiplayer Session Directory (MPSD)**,
  the Xbox service whose stated job is to "share the basic information needed
  to connect a group of users" — that is exactly the invite/join shape
  Dungeons exposes. (PUBLIC — [MPSD
  overview](https://learn.microsoft.com/en-us/gaming/gdk/_content/gc/live/features/multiplayer/mpsd/live-xbox-multiplayer-session-directory))

### Inferred, not observed

- **Dungeons on Switch uses XSAPI (or an equivalent in-house client) against
  the same `*.xboxlive.com` REST services as any other cross-play title.**
  (INFERRED from: Microsoft published Dungeons; cross-play is gamertag-keyed;
  XSAPI is the shipping cross-platform client and names Switch as a target.)
  Needs E2.
- **The invite/join path is MPSD sessions plus the real-time activity (RTA)
  websocket for change notifications**, with presence and people services
  behind the Friends tab. (INFERRED from the feature set matching MPSD's
  stated purpose.) Needs E2/E3.
- **Gameplay traffic does not go through Microsoft.** Peer-hosted co-op means
  the host's console carries the session; the Microsoft side hands out the
  connection information. (INFERRED from the peer-hosted design.) Needs E4.

### Unknown — needs a capture

- **Which hostnames the Switch build actually resolves.** UNKNOWN, needs E1.
  Nothing below should be treated as an endpoint list until E1 exists.
- **Whether Dungeons has any service of its own** beyond the Xbox services —
  entitlement/DLC checks, telemetry, a content or config feed. UNKNOWN, needs
  E1. This is the question that decides whether this repository ever holds
  code.
- **Whether the game or the SDK pins certificates**, and whether an
  Atmosphere/citron CA substitution is enough to terminate TLS for
  `login.live.com` and `*.xboxlive.com`. UNKNOWN, needs E3.
- **The gameplay transport**: protocol, port range, and how the peer address
  reaches the joiner. UNKNOWN, needs E4. Public port-forwarding pages are
  worthless here — they list the generic Xbox network set (88/500/3074/3544/
  4500, i.e. Teredo and IPsec) for the Xbox rows and a boilerplate GameSpy-era
  set for the Switch row, which no 2020 Switch title uses. Do not cite them.
- **Whether the client validates anything in the XSTS token it is handed.**
  UNKNOWN, needs E5. If it does not — the usual arrangement, because the
  *services* are the ones that verify — then a replacement adapter can mint
  its own tokens freely, and this is the single biggest feasibility lever.

## Why there is no per-title server to write

Compare with the two working non-Nintendo-network servers in this tree:

| | what the title needs from a server | OpenPak answer |
|---|---|---|
| Outbound | Photon Realtime auth, rooms, and a Fusion Shared-mode simulation | `servers/photon` — a real server, because the title needs an authoritative room host |
| Stardew Valley | NPLN identity and session RPCs | `servers/stardew-valley` — a real server, same reason |
| **Minecraft Dungeons** | a Microsoft identity, a Microsoft friend graph, and a session record to hang an invite on | **none of it is title-specific** |

Every line item in the Dungeons column is generic Xbox services. Writing
them in a repository named after one game would guarantee a rewrite the moment
`servers/minecraft` needs the same thing — which it will, because Bedrock on
Switch shows the same code-entry sign-in screen against the same identity
provider.

## Where the identity half belongs

**In a shared Xbox-Live-shaped adapter, alongside `nx-baas` — not here, and
not inside `account`.**

The shape follows what already works for the Switch:

- `account` stays the one identity store. An OpenPak account gains an
  Xbox-shaped projection (an XUID-equivalent, a gamertag) the same way it
  already gains a Nintendo-shaped one.
- A new adapter — call it `xbl` for now; OpenPak's roadmap already reserves
  the name `xbox-live` — owns the Microsoft-side wire protocol and nothing
  else. It terminates the console-facing hostnames and translates to
  `account`'s internal API, exactly as `nx-baas` does on `:20070` behind
  `X-Internal-Key`. It must cover, in this order:
  1. **MSA sign-in**, the device-code/remote-connect flow, so the player can
     link a console to an OpenPak account by entering a code — the same
     interaction OpenPak already implements for console linking, with a
     different front door.
  2. **The Xbox token chain**: user token, device token, title token, XSTS.
     Ours to mint; we are the issuer and the verifier.
  3. **People and presence**, backing the in-game Friends tab from
     `account`'s existing friend graph.
  4. **MPSD sessions plus the RTA websocket**, which is what an invite
     actually is.
  Items 1–4 are shared by every Microsoft cross-play title on every console.
  None of them are Dungeons.
- The console-side plumbing is the existing rig: DNS redirection of
  `login.live.com` and `*.xboxlive.com` (the console and Harbor use the
  declarative route table), a per-SNI certificate from our CA, and the CA
  trust patch. This is a known-good path — it is how the Switch adapter,
  Stardew and Outbound all got their traffic.

Note what this adapter is *not*: it is not the planned `x360-live`. The Xbox
360 generation speaks a different, older protocol, and the existing open prior
art (Xenia-WebServices, MIT; LibreLive, closed) targets that generation, not
the modern XSAPI REST services. For modern XSAPI there appears to be **no open
prior art at all** — searching turned up only 360-era and original-Xbox
projects. That is a cost, not a blocker: the client library and the service
documentation are both public and permissively licensed, which is a better
starting position than NPLN ever offered.

## What would change the verdict

In order of how much they move it:

1. **E1 shows a Dungeons-specific hostname** that is neither Microsoft
   identity nor Xbox services. Then this repository has a real, narrow job and
   should implement exactly that one service. This is the only finding that
   makes code here correct before the adapter exists.
2. **The Xbox adapter gets built** for `servers/minecraft` or for its own
   sake. Dungeons then becomes an integration exercise — E4 and E5, plus
   whatever session document fields the title expects — and this repository
   becomes a thin title profile, likely a config file and a test, not a
   server.
3. **E3 shows the client refuses a substituted CA and pins.** Then everything
   above is blocked until a client-side patch exists, and this repository
   stays parked regardless.

## Deliberate non-goals

- No skeleton, no `cmd/`, no handler stubs. There is nothing to stub against.
- No port block claimed in `ports.md`. The third-party block 22600–22699
  reserves room for this title; taking ten ports for a service that does not
  exist would be a lie in a shared document.
- No Dungeons-specific Xbox client work. If the adapter needs an XSAPI
  behaviour, it belongs to the adapter.

## Sources

- [Windows Central — Minecraft Dungeons cross-play guide](https://www.windowscentral.com/minecraft-dungeons-guide-cross-play)
- [GHOSTCAP — Is Minecraft Dungeons cross-platform? (2026)](https://www.ghostcap.com/is-minecraft-dungeons-cross-platform)
- [game8 — How to play online co-op](https://game8.co/games/Minecraft-Dungeons/archives/289091)
- [PC Gamer — How to set up online co-op in Minecraft Dungeons](https://www.pcgamer.com/minecraft-dungeons-crossplay-multiplayer-pc-coop/)
- [ACHIVX — Linking a Microsoft account on Switch](https://achivx.com/how-to-link-your-microsoft-account-to-minecraft-dungeons-nintendo-switch/)
- [Minecraft Wiki — Microsoft authentication](https://minecraft.wiki/w/Microsoft_authentication)
- [Xbox Support — Minecraft for Nintendo Switch FAQ](https://support.xbox.com/en-CA/help/games-apps/game-titles/minecraft-switch-edition-faq)
- [microsoft/xbox-live-api (XSAPI, MIT)](https://github.com/microsoft/xbox-live-api)
- [MicrosoftDocs/xbox-live-docs](https://github.com/MicrosoftDocs/xbox-live-docs)
- [Microsoft Learn — Xbox services authentication](https://learn.microsoft.com/en-us/gaming/gdk/docs/services/fundamentals/s2s-auth-calls/service-authentication/live-xbox-live-authentication)
- [Microsoft Learn — Multiplayer Session Directory](https://learn.microsoft.com/en-us/gaming/gdk/_content/gc/live/features/multiplayer/mpsd/live-xbox-multiplayer-session-directory)
- [wiki.vg — Microsoft Authentication Scheme](https://wiki.vg/Microsoft_Authentication_Scheme)
- [VGC — Minecraft Dungeons support ends; servers stay up](https://www.videogameschronicle.com/news/minecraft-dungeons-support-ends-as-the-game-hits-25-million-players/)
- [AdrianCassar/Xenia-WebServices (MIT, Xbox 360 era)](https://github.com/AdrianCassar/Xenia-WebServices)
