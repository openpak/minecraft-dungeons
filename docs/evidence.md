# Evidence: minecraft-dungeons

## Static recon (2026-10-26)

Switch v0 dump, fully decrypted (titlekey known): Minecraft Dungeons
`01006C100EC08000`, display version 1.0.0.6.0, build 2020-04-20, SDK 7.7.0.0.
`main` is a 55 MB NSO; NPDM name `Dungeons`.

Observed in binary strings (`main` + `subsdk0` + `sdk`):

- UE4 Mojang branch (`D:/depot/Jungle/Wolf/Engine/`), pak v8; custom
  `OnlineSubsystemDungeons` + `DungeonsLobby`.
- XAL + XSAPI `kinoko2_2020-04-08`; `XblMultiplayerActivity*` + MPSD fragment.
- `api.minecraftservices.com`, `signaling.minecraftservices.com` + `:5443/ws`,
  `stun:turn.minecraftservices.com:3478`, WebRTC ICE + `STUN_KEEPALIVE`.
- Xbox hosts: `auth`/`rta`/`notify`/`stats`/`vortex-events` `.xboxlive.com`.
- Zero NEX/PRUDP/Pia, Photon, EOS, or Demonware hits; 3 PlayFab remnant
  strings, no SDK. `ldn:u` declared in NPDM, no `nn::ldn` symbols (unused).

No runtime capture exists yet. Raw dumps, strings and notes stay outside this
repository.
