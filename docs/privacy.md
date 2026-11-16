# Privacy

hourdial records which app and window are in front, all day. That is private by
nature, so this page says exactly what is kept, where, and what can leave
your machine. Everything here is checked against the code; if you find a
difference, it is a bug (see [SECURITY.md](../SECURITY.md)).

## What is recorded

trackd writes spans to `~/.local/share/tock/tock.db`:

- start and end time;
- the app (bundle id and name);
- the window title, and the URL of the front browser tab;
- whether you were active, idle, asleep or locked.

It does not record keystrokes, screenshots, clipboard, or what is inside a
window. Optional extras, each off until you turn it on:

- **Calendar** (`[calendar]`): event titles and times from the calendars
  you pick, read through macOS's own calendar access.
- **Screen text (OCR)** (`[ocr]`, macOS): a few keywords read from the
  screen of windows the rules cannot label, and of the apps in `[ocr]
  apps`, to help label time. The screenshot is never written to disk; the
  keywords stay in tock.db and are never sent anywhere.

## What is never stored

- **Private browser windows:** recorded as `<app> (private)`, with no
  title and no URL. On macOS hourdial reads the URL from Chrome, Arc,
  Brave, Edge, Vivaldi, Chromium and Safari (beta and nightly builds
  included), and only titles from Firefox and its forks (Zen, LibreWolf,
  Waterfox, Floorp). Browsers whose private windows it cannot recognise
  (Opera, Dia, Orion, Tor Browser, Mullvad Browser, Safari Technology
  Preview, Comet, Helium, SigmaOS) are recorded as the app only, never
  with a title or URL.
- **Hidden apps and sites** (`[privacy] hide_apps`, `hide_sites`): kept as
  time only, with no title or URL.
- **While `config.toml` does not load**, trackd records nothing, so a typo
  can never switch your hide lists off.

`[privacy] keep_titles = N` deletes titles and URL paths older than N days.

`tock commits` reads your repos' history when you run it and stores
nothing from it; it prints each commit's short hash and subject. hourdial
(not trackd or the menu bar app) reads Claude Code's session logs in
`~/.claude/projects` for each entry's type, time and working directory
and each session's title, never the text of prompts, replies or tool
output (it counts the prompts you typed, by entry type). Session titles
are written by Claude Code from your first prompts, so they can hint at
what a session was about. It
keeps those times, directories and titles in `agents.index` in its state
folder. `[shell] agents = false` stops it using them for labels.

## Files

The data dir (`~/.local/share/tock`) and the state dir (`~/.local/state/tock`)
are readable by you only (0700 folders, 0600 files); trackd closes either
folder to other users at start if it finds it open. That covers the
database, its snapshots in `backups/`, the logs, and the invoice previews
and week reports hourdial writes for you (`drafts/`, `reports/` in the state
dir). Snapshots and `tock
backup` archives hold your titles and URLs too, so keep them out of git and
shared folders.

Issued invoices sit in the database and in `invoices/` in the data dir.
They hold your and your client's names, addresses and tax numbers, and the
time billed per day, but no window titles or URLs, so `tock prune` and
`keep_titles` leave them alone. They are never synced.

## Encryption at rest

`tock.db` is not encrypted; titles and URLs are plain text in a file only
you can read. We looked at encrypting it (2026-11-14) and decided not to,
for now. What we weighed:

**CPU is not the obstacle.** Measured on an Apple M2 with Go's AES-256-GCM:
sealing one 79-byte title takes 0.35 µs and opening it 0.1 µs; a 4 KiB
database page takes about 0.9 µs either way. Building a day view (about
180 spans) already takes 2.8 ms and a week 12 ms (`BenchmarkDay`,
`BenchmarkWeek` in `internal/app`), so decrypting every title in them
would add well under 1 %. trackd writes a span every few seconds, far
inside its 1 % idle CPU budget either way.

**Whole-file (SQLCipher-style) does not fit the build.** SQLCipher is C,
and trackd and tock are pure Go with no cgo (the Linux builds use
`CGO_ENABLED=0`). The SQLite driver hourdial uses (`modernc.org/sqlite`) has
no encrypting storage layer. The pure-Go route would be a different
driver with an encrypting VFS, which means replacing the database layer
and re-measuring trackd's memory, for the whole data set.

**Per-column (titles and URLs only) fits the build but covers less than it
seems.** The same titles also sit in snapshots, `tock backup` archives,
the sync folder's day files, screen text (`ocr`), other machines' views,
the agent index and crash logs; each would need the same treatment, and
URL hosts must stay readable for the rules.

**The key is the real problem.** Encryption protects against whoever
cannot read the key. Kept in a file next to the database, it protects a
copied `tock.db` that was copied without the key, nothing else. Another
program running as you, the main risk on a personal machine, can read
the key as easily as the database. A key in the macOS Keychain or a Linux
secret service would help, but that needs cgo or a helper program on
macOS and a running keyring on Linux (headless machines have none).

**Decision.** No encryption at rest in 1.x for now. What protects the data
today: full-disk encryption (FileVault, LUKS), which covers a lost or
stolen machine better than anything hourdial could add; 0600 files in 0700
folders; `keep_titles` and `tock prune`, which delete titles after N
days; and `[privacy] hide_apps` and `hide_sites`, which keep titles and
URLs from being stored at all.
If people ask for it, the candidate is per-column AES-GCM with the key in
the OS keychain, applied to every copy listed above, and it would come as
a minor release with its own migration.

## What leaves the machine

trackd and the menu bar app have no network code at all; tests check that
they don't link any. Only `tock` itself talks to the network, and only for
these, each off until you turn it on:

| Feature | Goes to | What is sent | Turn on / off |
|---|---|---|---|
| Claude suggestions | Anthropic, through the `claude` CLI or `ANTHROPIC_API_KEY` | app names, sites and title snippets of unlabelled time, with minutes | `tock choice claude on` / `off` |
| `tock summary` | Anthropic | the day's work blocks | same |
| `tock ask` | Anthropic | your question and your label names | same |
| Slack status | Slack | a status text and do-not-disturb during a focus session; no activity | `tock slack login` / `logout` |

Never sent to Claude: private windows, hidden apps and sites, the `[llm]`
deny lists (password managers and mail by default), and screen text.
`--dry-run` on `tock suggest`, `tock summary` and `tock ask` prints what
would be sent without sending it.

The dashboard (`tock web`) listens on 127.0.0.1 only, needs a per-run
token, and loads nothing from the internet.

## The sync folder

With `[sync] dir` set, each machine writes its spans (titles and URLs
included), its `machine.json` and, unless `rules = false`, its
`rules.toml` to that folder. hourdial does not upload it anywhere: whatever
syncs the folder (Syncthing, iCloud Drive, Dropbox) decides where it goes.
Pick one you trust with your activity.

On Linux, the sync folder can also carry new hourdial builds from the Mac, if
you turn that on (`[sync] update = true`). A build is installed only when
it is signed by the key you trusted by fingerprint with `tock update
trust`; the folder alone can never make a machine trust a key. See
"Updates" in the [guide](guide.md).

## Who can read your record

tock.db is an ordinary SQLite file: titles and URLs are stored as plain
text, protected by file permissions (0600) and nothing else (see
[Encryption at rest](#encryption-at-rest) for why). This is what that
does and does not cover.

| Who | Can they read it? |
|---|---|
| Another user account on the same machine | No, unless they are an administrator. |
| An administrator or root | Yes. |
| Any program running as you (including malware) | Yes. hourdial cannot defend against your own account. |
| Someone with your unlocked machine | Yes. |
| Someone with your powered-off or stolen machine | Only if the disk is not encrypted. Turn on FileVault (macOS) or LUKS (Linux). |
| Your backup tool (Time Machine, a cloud backup of your home folder) | Yes, the data dir is backed up like any other folder unless you exclude it. |
| Whatever syncs your sync folder | Yes, if you set `[sync] dir`: spans go there with titles and URLs. |
| Anthropic or Slack | Only what the table above lists, and only after you turn the feature on. |
| Other sites or programs, through the dashboard | No. It listens on 127.0.0.1, needs a per-run token and checks the Host header. |
| The author of hourdial | No. Nothing is sent, and there is no account or server. |

To keep less: `hide_apps` and `hide_sites` keep only the time,
`keep_titles = N` deletes titles and URL paths after N days, `tock prune`
does that now (the spans and their labels stay), and `tock uninstall
--purge` removes everything.

## What hourdial is for

hourdial is a tool for looking at your own time. It is not built for watching
anyone else, and it is not to be used that way: do not install it on a
machine another person uses without them knowing and agreeing. It has no
remote view, no hidden mode and no central server, on purpose. An employer
who records staff activity takes on legal duties (GDPR in the EU, and
similar laws elsewhere) that hourdial does nothing to meet.

## No telemetry

hourdial sends no usage data, crash reports or analytics. `tock bug-report`
writes a file for you to look through and attach to an issue yourself:
`tock doctor --json`, `config.toml` with every text value removed, the
end of the logs without lines that could hold a title or URL, the crash
logs, and when trackd ran in the last two weeks. Your home and sync
folders are replaced by `~` and `<sync>`, and Claude, Slack and dashboard
tokens by `<token>`. Your rules go in only with `--include-rules`.

Crashes are written only to `~/.local/state/tock/<program>-crash.log`
(mode 600). A panic's message can be any value, a window title included,
so the bug report keeps only the trace (function names, files, line
numbers, addresses) and the Go runtime's own errors (`runtime error:
index out of range`, `concurrent map writes`); every other line becomes
`<removed>`.
