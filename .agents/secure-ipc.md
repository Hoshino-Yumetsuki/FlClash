# FlClash Secure IPC Design

## Threat

PoC: https://github.com/Yohane-Mashiro/flclash-lpe

Privileged core (Windows LocalSystem via helper, or Unix setuid root) dials an
attacker-controlled IPC endpoint. Action layer has no auth, so `deleteFile` /
`getConfig` / config setup run as the privileged identity.

## Root cause

1. Helper / setuid path trusts caller-controlled IPC address.
2. Helper HTTP (`127.0.0.1:47890`) has no caller authentication.
3. Core actions have no session authentication.
4. File actions are not sandboxed to the app home directory.

## Design (defense in depth)

```
UI ──auth──▶ Helper (Windows only) ──spawn+env──▶ Core
UI ◀──length-prefixed JSON + session token──▶ Core
```

| Layer | Rule |
| --- | --- |
| Helper API | Every endpoint requires `X-FlClash-Auth` = install secret. Core path is fixed next to helper binary (client `path` ignored). `/start` only accepts `\\.\pipe\FlClashCore_*`, generates session token server-side, returns JSON `{token}`, injects as `FLCLASH_IPC_TOKEN`. |
| Core session | Desktop core requires `FLCLASH_IPC_TOKEN`. First frame must be `auth` with that token. All other methods rejected until authenticated. |
| Path sandbox | `deleteFile`, `getConfig`, `validateConfig` only allow paths under `constant.Path.HomeDir()` (canonical prefix check). |
| Unix setuid | If `euid==0 && ruid!=0`, parent realpath must equal installed UI binary next to core (no basename substring match). |
| Socket perms | Unix IPC socket `chmod 0600` after bind. |

## Non-goals (this pass)

- Full replacement of setuid with a Unix privileged helper daemon.
- Code signing enforcement of the UI binary on Windows (path+secret only).
- Android FFI path (in-process; out of LPE scope).

## Compatibility

Older helpers without auth will fail `pingHelper` / `startCoreByHelper` and fall
back to non-elevated `Process.start` where possible. Re-authorize (reinstall
helper) after upgrade.


## Residual after review

- Same local user who can read `helper.auth` can still call `/start` and receive the session token (no process-identity on HTTP). Full fix: named pipe + SID / invert dial model.
- Path sandbox uses `EvalSymlinks` when possible.
