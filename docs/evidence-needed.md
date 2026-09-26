# Evidence needed

The design in `design.md` is built entirely on public documentation. No packet
from a Minecraft Dungeons client has been observed. This is the ordered list of
observations that would confirm it, narrow it, or kill it.

Rules, from `https://github.com/openpak/docs/blob/main/playbooks/online-compatibility-playbook.md` and the family
conventions:

- **One variable per run.** State the decision rule before the run, not after.
- **Raw captures, logs, tokens and account data stay outside this repository.**
  Keep them under a research directory on the pool, not in `/tmp`; commit only
  sanitized conclusions.
- **Emulator Launch Protocol** (`https://github.com/openpak/docs/blob/main/playbooks/emulator-launch-protocol.md`)
  before killing or restarting any emulator instance. Launch the fork
  directly; there are no per-title profiles and no persona launchers.
- Stop early. **E1 alone can end this investigation**, in either direction.

---

## E0 — Do we have the title at all?

**Question.** Is there a Minecraft Dungeons Switch dump on this machine, which
update version, and does it reach its main menu under citron and Ryujinx?

**Method.** Check the emulator game lists for the title; record the base title
ID and the installed update version. Boot it once per emulator with networking
enabled but no redirects, and note whether it reaches the home screen. Ryujinx
writes one log per launch (`portable/Logs/Ryujinx_<ver>_<date>.log`); citron
truncates `~/.local/share/citron/log/citron_log.txt` on every start of the same
profile, so run one instance at a time.

**Decision rule.** No dump, or it will not boot on either emulator ⇒ stop.
Every item below needs a running client, and nothing else in this repository is
worth doing without one.

**Cost.** Minutes. Do this before reading further.

**Record.** Base title ID and update version — `design.md` currently states
neither, because guessing an identifier is worse than admitting we do not have
it.

---

## E1 — What hostnames does it resolve?

**The decisive one.** Everything in `design.md` hinges on whether Dungeons
talks only to Microsoft identity and Xbox services, or has a service of its
own.

**Question.** From a cold boot to the point where the game offers the Microsoft
sign-in code, which hostnames does the title resolve, in what order?

**Method.** Boot with networking enabled and **no redirects at all** — this run
is passive observation of the game against production. Walk it to the home
screen, then start the Microsoft sign-in and stop at the code screen. Mine the
emulator's own DNS resolution log, which is metadata and the cheapest lawful
source:

- citron: `grep -i 'sfdnsres\|resolve' ~/.local/share/citron/log/citron_log.txt`
- Ryujinx: `grep 'GetAddrInfoRequestImpl: Trying to resolve' portable/Logs/Ryujinx_*.log`

Do both emulators if both boot; a hostname that appears on one only is an
emulator gap, not a protocol fact.

**Decision rule — this is the branch point for the whole repository.**

- Every hostname belongs to Microsoft identity (`login.live.com`,
  `*.live.com`), Xbox services (`*.xboxlive.com`), or a CDN/telemetry host
  with no session role ⇒ **the verdict in `design.md` stands.** There is no
  per-title server. Park this repository, and record the hostname inventory
  here as the input to the Xbox adapter's design.
- A hostname appears that is Dungeons- or Mojang-specific and is contacted
  *before or during* sign-in or session setup ⇒ **this repository has a job.**
  Go to E1a.
- Nothing resolves at all before the sign-in screen ⇒ the client is offline
  until the player opts in; re-run with the sign-in flow started, and note that
  as a finding (it means every gate below is behind a user action).

**Record.** A hostname inventory table: host, first observed at (boot / menu /
sign-in), and the guess at its role. Names only. No payloads.

### E1a — only if E1 found a title-specific host

**Question.** What does that host do — is it session-relevant, or is it
telemetry, DLC entitlement, or a content feed we can ignore?

**Method.** Redirect that one host to a local listener that logs metadata only
(connection, SNI, byte counts, timing — never values), leaving every Microsoft
host untouched. citron takes a per-title redirect hook in
`src/core/hle/service/sockets/sfdnsres.cpp` behind an env var, matching the
existing `OPENPAK_PHOTON_IP` / `OPENPAK_AMONGUS_IP` / `OPENPAK_CTR_AUTH_IP`
pattern; Ryujinx needs no rebuild — it writes the console's
`atmosphere/hosts/default.txt` itself.

**Decision rule.** The game blocks, retries, or refuses to progress when that
host is dark ⇒ it is a gate, and implementing it is this repository's scope.
The game shrugs and continues ⇒ telemetry. Write it down and go back to
parking the repository.

---

## E2 — Confirm the Xbox services surface

**Question.** Does the client actually speak to `*.xboxlive.com`, and which
service subdomains — token (`user.auth`, `xsts.auth`), people/presence, session
directory, real-time activity?

**Method.** Extend the E1 log mining through a complete sign-in on a throwaway
Microsoft account, then through opening the Friends tab and sending an invite.
Record the hostnames and the order. Cross-reference each against the public
XSAPI source and the Xbox services documentation to name the service — the
naming comes from the public docs, not from the capture.

**Decision rule.** The observed subdomains map onto documented XSAPI services
⇒ the Xbox adapter can be specified from public documentation, and this is the
input to that specification. They do not map ⇒ Dungeons uses a private
Microsoft surface and the adapter is far more expensive than
`design.md` assumes; say so loudly.

**Cost.** Needs a throwaway Microsoft account and a working NSO session on the
emulator. Do not use a personal account.

---

## E3 — Will the client accept our certificate?

**Question.** Can `login.live.com` and `*.xboxlive.com` be terminated with a
certificate from our CA, or does the title or the SDK pin?

**Method.** The known-good rig: per-SNI certificates from our CA, the CA trust
patch on the client, DNS redirect of one hostname at a time. Start with the
least important Xbox host found in E2, not with `login.live.com` — a failure
on a leaf host is a cheap answer to the same question. A plain TLS terminator
that completes the handshake and logs the handshake parameters is enough; it
does not have to answer correctly.

**Decision rule.** Handshake completes and the client sends a request ⇒ TLS is
not a blocker, and the Xbox adapter is buildable on the existing rig.
Handshake is refused, or completes and the client immediately drops ⇒ pinning
is suspected; the next step is a client-side patch investigation, and the whole
Dungeons line is blocked behind it. Record which.

**Note.** The terminator belongs in the Xbox adapter or in a scratch directory,
**not in this repository**. This repository has no code and should not acquire
some as a side effect of an experiment.

---

## E4 — What carries the gameplay?

Only worth running after E2 and E3, and only with two clients that can actually
reach each other.

**Question.** When two players are in one session, what protocol and ports
carry the game, and how does the joiner learn the host's address?

**Method.** Host and joiner personas, one session, socket-level logging from
both emulators (`bsd` connect/send/recv with sizes and peer addresses). Note
whether the peer address is a public address, a relay, or something handed
down from the session record.

**Decision rule.** Direct peer addressing ⇒ OpenPak needs a NAT/relay story
for Dungeons (the existing `nat-check` and the STUN/TURN block at 22900–22999),
but no game server. A Microsoft-side relay in the path ⇒ that relay has to be
replaced too, and the cost of the whole Dungeons line roughly doubles.

**Record.** Protocol, port range, and the addressing mechanism. No payloads.

---

## E5 — Does the client verify the tokens it is handed?

The largest single feasibility lever, and cheap once E3 passes.

**Question.** If our adapter mints an XSTS-shaped token with its own signature,
does the client accept it and proceed, or does it validate a signature it will
never be able to verify against us?

**Method.** With E3's terminator in place on the token endpoint, return a
well-formed but self-signed token with plausible claims (a gamertag, an
XUID-shaped identifier). Watch whether the client proceeds to the next service
call or errors out.

**Decision rule.** Client proceeds ⇒ the adapter is a normal server-side
reimplementation and the Dungeons line is genuinely feasible. Client rejects
⇒ find out what it checked before concluding anything; a claim shape mismatch
looks identical to a signature check from the outside, and only one of the two
is fatal.

---

## E6 — The cheap kill switch

**Question.** Does the Switch build have a genuinely offline mode, and is
single-player unaffected by all of the above?

**Method.** Boot with networking disabled entirely and play a mission.

**Decision rule.** Public reporting says the Switch version is "true offline —
you never have to touch the internet to play it". If that holds, then the
*only* thing OpenPak adds for this title is co-op with friends, and the
priority of everything above should be judged against that, not against
"the game does not work". Confirm it once so the scope conversation is honest.

---

## What is deliberately not on this list

- Any experiment that needs the Xbox adapter to exist. Those are the adapter's
  evidence list, not this title's.
- Anything about Minecraft Bedrock. Same identity provider, different
  repository (`servers/minecraft`); do not test both in one run and attribute
  the result to either.
- Payload decoding of Microsoft traffic. E1–E4 need hostnames, sizes and
  ordering only, and metadata is both cheaper and safer than payloads.
