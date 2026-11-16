# Security review

An internal review, so an outside reviewer can check each claim instead of
trusting it. Each line names where to look. Last done 2026-11-14 against
main (the agent ledger, budgets, replay, the Sway, i3 and X11 collectors
and the crash logs were added to it then); report anything that does not
hold through [SECURITY.md](../SECURITY.md).

## What is worth protecting

Which apps, windows and sites you used, and when. It lives in one SQLite
file plus backups, the sync folder if you use one, short-lived logs, the
agent index and any crash logs.

## Checked

**Nothing leaves the machine unless you turned it on.**
- trackd imports no HTTP client; its only network-ish code is the desktop
  sockets (`internal/collect/hypr.go`, `i3ipc.go`, `x11.go`).
- `tock` talks to Claude (`internal/claude`) and Slack (`internal/slack`)
  only when configured. What goes to Claude passes `suggest.Filter`
  (`internal/suggest/filter.go`): private windows, `[llm]` deny apps,
  hosts and titles, and items under the minimum time are dropped.
- [privacy.md](privacy.md) lists what each feature sends.

**The dashboard on 127.0.0.1** (`internal/web/server.go`)
- Requests with any Host but `127.0.0.1:<port>`, `localhost:<port>` or
  `[::1]:<port>` are refused, which stops DNS rebinding.
- Every API call needs a 256-bit per-run token, compared in constant time.
  The page gets it from the URL fragment, which browsers never send.
- CSP `default-src 'self'`, `frame-ancestors 'none'`, `X-Frame-Options:
  DENY`, same-origin COOP/CORP: no framing, nothing loaded from outside.

**Files** — data, state and invoice folders are made 0700 and chmodded
if they existed; the database, its WAL and SHM, backups (VACUUM INTO then
chmod), sync files, tokens and keys are 0600 (`internal/store/store.go`,
`internal/snapshot/file.go`, `internal/syncdir`, `cmd/tock/cmd_slack.go`).

**Updates** (`internal/update`, `cmd/trackd/update.go`) — hourdial has no
self-update from GitHub; release downloads are checked by hand
([verify.md](verify.md)). The one updater is the opt-in Linux one, which
installs what the Mac publishes to the sync folder. The sums come from
that same folder, so the question was whether a swapped binary plus a
swapped sum list would be accepted. It would not: the manifest is signed
with ed25519, the private key stays on the Mac (`cmd/tockpub` refuses a
key inside the sync folder), and the public key is pinned on the Linux
machine only through `tock update trust <fingerprint>`, never taken from
the folder (`Pinned`, `Trust`). The bytes copied are hashed as they are
written and checked before any rename (`Apply`, `stage`); older builds
are refused by build time. Found and fixed 2026-11-14: the manifest did
not name its arch, so a signed arm64 build moved into the amd64 folder
would have been installed (and trackd would not start). The arch is now
signed and checked (`forArch`). Left as is: a binary built without a
build time (plain `go build`) accepts older signed builds, and anyone who
can write the sync folder can stop updates.

**Stopping trackd on Windows** (`internal/stopsignal/stopsignal_windows.go`)
— the event's DACL grants only the current user's SID.

**Bug reports** (`cmd/tock/cmd_bugreport.go`) — config values are
stripped, log lines that could hold a title or URL are dropped, rules
are left out unless asked for; nothing is sent. Home folders (any user's
`/Users/…`, `/home/…`), Claude and Slack tokens and the dashboard token
are replaced in every file (`scrubSecrets`). Crash logs keep only lines
shaped like a Go trace and the runtime's own messages (`redactCrash`);
`TestBugReportCrashLogs` plants a title, URL, tokens and a user name in
a panic and checks none comes out.

**Agent ledger** (`internal/agentlog`, `internal/app/agents.go`) — reads
Claude Code's logs under `~/.claude/projects` and nothing else; it never
writes there. Each line is decoded into a struct of type, time, working
directory, session id and the session's title, so the prompt and reply
text passes through memory while a line is parsed but is never kept, and
`agents.index` (in the state folder, written 0600 through a temp file)
holds only times, directories and titles. Nothing it reads is sent;
`tock agents` and the terminal labels are local. The titles are written by
Claude Code from your conversation, so they say what a session was about;
`agents.index` is as private as `tock.db`. Only `tock` reads these logs
(the terminal labels need `Paths.AgentIndex`, which trackd does not set).
Found and fixed 2026-11-14: a line was read whole whatever its size, and
a pasted image makes lines of tens of MB, so `tock web`, which looks
every 30 seconds for hours, held about twice such a line in memory
(measured: 40 MB for a 20 MB line). Lines over 1 MB are now stepped over
in a bounded buffer, 4.5 MB for the same line (`readLine`,
`TestLongLineIsSkipped`).

**Budgets** (`internal/app/budget.go`, `cmd/tockbar`) — computed from the
database; nothing new is read or sent. The notice goes through the
`notice` table to the menu bar app, which draws it with AppKit text, or
through `osascript`/`notify-send`, where title and body are argv items,
never part of a script (`cmd/trackd/notify.go`). Some Linux notification
servers render simple markup in the body; the text is your own project
names and numbers.

**Replay** (`web/src/lib/replay.ts`, `replayVideo.ts`) — plays the rows
`/api/day` already sent, behind the same token, in the page. The video is
drawn on a canvas and saved with a local download; nothing is uploaded.
The clip shows the app and window title under the playhead, so look
through it before you share it.

**Sway, i3 and X11** (`internal/collect/i3ipc.go`, `x11.go`)
- Sway and i3: the socket is `$SWAYSOCK`, `$I3SOCK` or one under
  `$XDG_RUNTIME_DIR`, which is yours alone. A reply over 4 MiB is
  refused before it is read. Real Sway and i3 trees, measured in the e2e
  job with 42 windows titled like browser tabs, take about 1.1 KB per
  window, so 4 MiB holds about 3,800 windows; reading and decoding a
  reply that size costs under 8 MiB, well inside trackd's 40 MB budget.
  When a tree is refused, trackd logs it once and keeps sampling the
  window its last focus event named.
- Blocking an app moves its windows to the scratchpad by app id; the id is
  regex-quoted and its `"` escaped before it goes into the command
  (`SwayHideCommands`). The id comes from the window, that is from a
  program already running as you, so this is not a boundary between users.
- X11: only local displays (`/tmp/.X11-unix/X<n>`); the MIT cookie from
  your Xauthority goes only to that socket. Found and fixed 2026-11-14: a
  reply's length was taken from the server and allocated as is (up to
  16 GB); replies over 1 MB are now refused (`xMaxReply`,
  `TestX11RefusesAHugeReply`).
- Neither adds network code: `cmd/trackd/deps_test.go` fails if trackd
  links `net/http` or `crypto/tls`.

**Crash logs** (`internal/crashlog`) — each program sends the Go
runtime's fatal output also to `<state>/<program>-crash.log`, created
0600 in the 0700 state folder, written only when it crashes, moved to
`.1` above 256 KB at the next start. No network. A panic message may hold
any value, a title included, so the file is as private as the database;
the bug report redacts it (below).

**Dependencies** — `govulncheck` runs in CI and before every release
(`.github/workflows/ci.yml`, `release.yml`); `npm audit --omit=dev` on
`web/`: 0 vulnerabilities. Dependabot is on.

**Releases** — tarballs and the .dmg get GitHub build provenance
(`gh attestation verify <file> -R confuserdotcom/hourdial`). The tarballs are
listed in `SHA256SUMS`; the .dmg has its own `.dmg.sha256`. The macOS app is ad-hoc signed, not notarized: there is no
Apple Developer ID, so the first open needs Privacy & Security → Open
Anyway.

## Open

Found in the 2026-11-14 pass and not changed:
- The X11 socket's owner is not checked. `/tmp/.X11-unix` is normally
  root's with the sticky bit, so only the X server can make the socket;
  where it does not exist, another local user could make it first and
  read the cookie.
- The Sway/i3 reply cap is 4 MiB, set from real trees measured in the
  e2e job and below trackd's memory budget. Only your own window manager
  can send one.
- An hourdial built without a build time accepts older signed Linux builds
  (see Updates).

## Not covered

- Another process running as you: it can read tock.db like any of your
  files. Out of scope by design. Encryption at rest would not change that
  while the key sits where the same process can read it; see
  [privacy.md](privacy.md#encryption-at-rest).
- The sync folder's own sharing and encryption.
- An independent reviewer. This page was written by the project.
