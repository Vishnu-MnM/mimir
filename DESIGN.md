# mimir: TUI Todo App Design Doc (pre-code)

## Context
mimir is a learning project: a feature-rich, robust terminal todo app built test-first (TDD). The goal is to learn how a language, its libraries, and the architecture fit together. This doc settles features, architecture, edge cases and language **before** any code. Development happens on two machines (macOS and Linux) from the same GitHub repo (`github.com/Vishnu-MnM/mimir`, public). The repo is the only shared state between them, so decisions and progress are recorded in §7 of this doc.

---

## 1. Feature set

### 1.1 Core todo (required)
- **Task fields:** title, body (markdown/plain), status (todo / in-progress / done / cancelled), priority (none, low, med, high, urgent), due (date-only *or* datetime with timezone), start/scheduled date, tags, project, created/updated/completed timestamps, stable ID (UUIDv7: sortable, unique across sync).
- **Operations:** add, edit inline, edit body in `$EDITOR`, complete/uncomplete, delete → **trash** (restore or purge), archive, reorder (manual sort), bulk-select actions.
- **Quick-add syntax:** `Buy milk tomorrow 5pm #errands +home !high`, parsed into fields, with a preview of the parse before saving.
- **Views:** Inbox, Today, Upcoming, Overdue, by Project, by Tag, Done; saved custom filters.
- **Undo/redo** for every mutation (command pattern, persisted per session).

### 1.2 Calendar
- **.ics:** import/export VEVENT + VTODO, and subscribe to ICS URLs (read-only, with refresh).
- **Google** (Calendar + optionally Tasks API) and **Outlook** (Microsoft Graph: calendar + To Do), via OAuth 2.0 with PKCE.
- **Calendar pane toggle:** a keybind (default `c`, configurable) plus a mouse click on a tab or header. Shows agenda/week/month, with tasks that have due dates overlaid on events.
- **Sync model (decided):** two-way and local-first, using an outbox queue. Works fully offline and syncs when connected.
- **Remote form (decided):** todos ↔ **Google Tasks** / **Microsoft To Do** (native task systems). Calendar events are pulled from Google Calendar / Graph calendar and shown in the pane. There's an optional "time-block" action that creates a calendar event linked to a task. Caveat to design around: Google Tasks only stores a due *date* (it drops the time), so the local due time is kept and only the date goes remote.
- **Configurable keybinds + config file (TOML, XDG paths):** core, because the calendar toggle must be rebindable.

**Language (decided): Go.** Stack: Bubble Tea, Lip Gloss, Bubbles, bubblezone (mouse zones), Glamour (markdown), termenv, `modernc.org/sqlite`, `golang.org/x/oauth2`, `google.golang.org/api/{tasks,calendar}/v3`, Graph via plain REST, `arran4/golang-ical`, `teambition/rrule-go`, `zalando/go-keyring`, `rivo/uniseg`, `BurntSushi/toml`, `pgregory.net/rapid`, `teatest`.

### 1.3 Theming (respect the terminal)
- **Default theme uses only ANSI palette indices 0–15 and the terminal's default fg/bg** (no hex colors). Catppuccin or any terminal scheme then applies automatically, because the terminal maps those indices to its own colors. This is how most TUIs "respect" your theme.
- Detect capability: `NO_COLOR` → monochrome (bold, underline and reverse only); `TERM=dumb` → plain; also detect light vs dark background (OSC 11 query) to pick contrast-safe index pairs.
- Optional user themes (TOML) may use truecolor, but never by default.
- Express emphasis through style (bold, dim, italic, strikethrough for done tasks) so the UI stays legible in monochrome.

### 1.4 Text handling
- **Store the raw source text as-is** (UTF-8 `TEXT` column) plus a `format` flag (`plain` | `markdown`). Never store rendered output; render on display and cache in memory keyed by `(id, updated_at, width)`.
- **Efficiency reality check:** a typical note is under 1 KB, and SQLite handles millions of rows. Compression (zstd) only pays off past roughly 4 KB per note, so add it later behind a threshold if profiling ever shows a need. That makes it a good "measure first" lesson.
- **Search:** an SQLite FTS5 virtual table indexed on a *markdown-stripped* plain-text copy, kept in sync by triggers.
- **Normalization on input:** CRLF→LF, Unicode NFC, strip C0 control chars and **ANSI escape sequences**. This is a security edge case: a malicious `.ics` DESCRIPTION could inject terminal escapes.
- **Display width:** use grapheme-cluster width (CJK, emoji, ZWJ sequences, combining marks) for truncation and wrapping, never byte or rune length.
- **Calendar interop:** Google and Graph return HTML bodies, while `.ics` DESCRIPTION is plain text (X-ALT-DESC may hold HTML). We need HTML↔markdown conversion at the sync boundary.

### 1.5 Extras: all selected (Tiers A, B and C)
Notes on the extras: subtasks sync as one level of nesting (the Google Tasks / To Do limit), and deeper levels stay local-only. Reminders run through `mimir notify` under launchd (macOS) or a cron/systemd timer. Plugins are external executables that get JSON on stdin, run with a timeout, and never block the UI.
| Tier | Feature | What you learn |
|---|---|---|
| A | Natural-language dates + RRULE recurrence (incl. "every last Friday", completion-based repeats) | Parsing, RFC 5545, property testing |
| A | Subtasks (parent/child, progress, completion rules) | Tree data in SQL (recursive CTEs) |
| A | Fuzzy search + command palette (`:` or `ctrl+p`) | Algorithms, UI composition |
| A | Configurable keybinds + config file (TOML, XDG paths) | Config layering, validation |
| A | Non-interactive CLI (`mimir add "..."`, `mimir ls --json`) | Shared core with two front-ends |
| B | Reminders/notifications (background daemon or `mimir notify` via launchd/cron) | Processes, OS integration |
| B | Import/export: todo.txt, JSON, Markdown checklists | Serialization, round-trip tests |
| B | Time tracking / Pomodoro per task | Timers and messages in an event loop |
| B | Multi-instance safety + live reload when the DB changes externally | File locking, WAL, file watching |
| C | Task dependencies ("blocked by") | Graph cycle detection |
| C | Stats view (completion streaks, burndown) | Aggregation queries, terminal charts |
| C | Plugin/hook scripts on events | Process boundaries, IPC |

### 1.6 Layout & interaction (vertical tabs, full mouse support)
Navigation uses a browser-style **vertical tab sidebar**. Every element is clickable, and every mouse action also has a keyboard equivalent.

```
┌─ mimir ─────────┬────────────────────────────────────┬─────────────────────┐
│ ▌Inbox        12│ ☐ !! Buy milk             tmrw 5pm │ Buy milk            │
│  Today         4│ ☐ !  Write design doc          Fri │ due  Tue 28 Oct 5pm │
│  Upcoming      9│ ▾ ☐  Plan trip                 3/5 │ tags #errands       │
│  Overdue       1│     ☐ Pack bag                     │ ─────────────────── │
│  Done           │     ☑ Book flights                 │ ## Notes            │
│ ─ Projects ──── │                                    │ - semi-skimmed      │
│ ▾ home         3│                                    │                     │
│   work         6│                                    │                     │
│ ─ Tags ──────── │                                    │                     │
│   #errands     2│                                    │                     │
│ ─────────────── │                                    │                     │
│ ◫ Calendar    c │                                    │                     │
│ ⚙ Settings      │                                    │                     │
│ ⟳ Synced 2m ago │                                    │                     │
└─────────────────┴────────────────────────────────────┴─────────────────────┘
 a add  e edit  space done  / search  ^p palette  c calendar  ? help
```
Calendar pane (toggled with `c` or by clicking ◫). It takes the detail column's place, or appears alongside it at ≥160 cols:
```
┌─────────┬─ Calendar ───────┐
│ ▌Agenda │ Tue 28 Oct       │
│  Week   │ 09:00 Standup  G │
│  Month  │ 17:00 ☐ Buy milk │
│ ─────── │ Wed 29 Oct       │
│ ☑ Google│ all-day Holiday O│
│ ☑ Outlk │                  │
│ ☐ ics   │                  │
└─────────┴──────────────────┘
```

**Regions**
- **Sidebar (vertical tabs):** smart views (Inbox, Today, Upcoming, Overdue, Done, saved filters), collapsible Projects tree, Tags. Pinned at the bottom are the toggles: Calendar, Settings, Sync status (click to sync now). It collapses to a ~4-col icon rail (`ctrl+b` or click). Count badges update live.
- **List pane:** tasks with clickable checkbox, priority, title, due date and subtask progress, plus collapsible subtask trees. Virtualized for 10k+ tasks.
- **Detail pane (toggle):** fields as clickable chips (click due → date picker, click tag → filter) and the rendered markdown body.
- **Calendar pane (toggle):** its own vertical sub-tabs (Agenda / Week / Month) plus per-source visibility checkboxes (Google, Outlook, each ICS subscription), with source badges on events.
- **Settings tab:** vertical sub-tabs: General, Keybinds, Theme, Accounts, Sync, Notifications, Plugins. Editing a setting in the UI writes back to the TOML config.
- **Footer hint bar:** context-sensitive, and each hint is clickable (runs the action).
- **Overlays:** quick-add, command palette, help, confirm dialogs, right-click context menu, date picker. Clicking outside or pressing `esc` closes them.

**Mouse map**
| Action | Mouse | Keyboard |
|---|---|---|
| Switch tab / view | left click | `j/k` in sidebar, `1–9`, `gt/gT` |
| Select task | left click | `j/k` |
| Toggle done | click checkbox | `space` |
| Edit | double-click | `e` / `enter` |
| Expand/collapse | click ▸/▾ | `za` / `→ ←` |
| Context menu | right click | `m` |
| Scroll | wheel (scrolls the pane *under the cursor*, not the focused one) | `ctrl+d/u` |
| Resize panes | drag the border (widths persist) | `<` / `>` |
| Reorder | drag task within list | `J/K` |
| Move/tag/schedule | drag task onto a sidebar project/tag or a calendar day | `M` / `t` / `d` |
| Focus pane | click inside it | `tab` / `shift+tab`, `h/l` |
| Hover highlight | pointer motion (optional) | n/a |

**Responsive breakpoints** (computed in a pure `layout.Compute(w, h, state) → rects` function)
- ≥160 cols: sidebar + list + detail + calendar
- 120–159: sidebar + list + one of detail/calendar
- 80–119: sidebar + list; detail/calendar open as an overlay
- <80: sidebar becomes the icon rail; one pane at a time
- <40×10: list only, no chrome

**Implementation notes (Go)**
- Mouse mode: `tea.WithMouseCellMotion` (click, wheel, drag). Hover needs all-motion reporting, so it's opt-in in config, because some terminals and tmux flood events.
- Hit-testing in two levels: pane routing from our own `layout.Compute` rects (pure and unit-tested), and clickable items inside panes via `lrstanley/bubblezone` zone markers.
- Bubble Tea has no double-click event, so we detect it ourselves (same cell within ≤400 ms) using the injected Clock, which makes it TDD-friendly.
- Glyph sets: `unicode` (default), `nerd` (opt-in Nerd Font icons) and `ascii` fallback. Focus and selection are shown with reverse/bold, not color alone.

**Edge cases**
- Mouse reporting unavailable (tmux `mouse off`, some terminals): the keyboard path still covers everything.
- Mouse capture blocks native text selection, so document shift+drag and add a key that toggles capture.
- Coordinates need recomputing after a resize.
- A drag released outside the window cancels the drag.
- Wide terminals (>223 cols) need SGR extended mouse mode (Bubble Tea's default).
- Clicks landing on a stale zone during rerender.
- A drop target that disappears mid-drag.

**Tests**
- Table tests for `layout.Compute` at each breakpoint.
- Synthetic mouse-message sequences → expected model state (click, drag-drop, wheel routing).
- Double-click timing with a FakeClock.
- Golden snapshots for each breakpoint, in color and NO_COLOR modes.

---

## 2. Architecture (language-agnostic, ports & adapters)

Go module `github.com/Vishnu-MnM/mimir`; everything lives under `internal/` except `cmd/`.
```
cmd/mimir/      single binary: no args → TUI, subcommands → CLI (add, ls, done, sync, notify, import, export)
internal/
domain/         Task, Recurrence, Filter, QuickAdd parser, dependency graph. PURE, no I/O
app/            use cases + Command objects (undo/redo), event bus, plugin dispatch
ports/          interfaces: TaskRepo, Clock, IDGen, TaskProvider, CalendarProvider, TokenStore, Notifier
adapters/
  formats/      todo.txt, JSON, Markdown checklist import/export
  notify/       macOS (osascript / terminal-notifier), Linux (notify-send)
  sqlite/       repo, migrations, FTS
  ics/          parse/serialize, URL subscription
  google/       Calendar (+Tasks) client
  msgraph/      Graph calendar (+To Do) client
  auth/         OAuth PKCE loopback flow, device-code flow, keychain token store
sync/           engine: outbox, change tracking, conflict resolution, backoff
ui/             TUI: app model, layout engine + hit-testing, components (sidebar tabs, list, detail, editor, calendar pane, settings, palette, context menu, date picker), mouse/drag state, theme, keymap
config/         TOML load, defaults, validation, XDG paths
```
- **Rule:** dependencies point inward. `domain` imports nothing, and adapters implement ports. This is what makes TDD practical, because every test can swap in fakes (FakeClock, InMemoryRepo, FakeProvider).
- **Sync engine:** each local change writes an outbox row in the *same transaction*. The engine pushes the outbox, then pulls remote changes via incremental tokens (Google `syncToken`, Graph delta links). Per-item mapping table: `(local_id, provider, remote_id, etag, last_synced_hash)`. Conflict = both sides changed since `last_synced_hash`. The policy is field-level merge, falling back to last-writer-wins, with the losing version kept as a conflict copy.
- **Auth:** loopback redirect (`http://127.0.0.1:<random port>`) + PKCE + `state` check. Graph also supports the device-code flow (good over SSH). Tokens go in the OS keychain, with an encrypted-file fallback. Refresh happens proactively before expiry.
- **Time model:** store instants as UTC + original TZID. All-day items are **dates, not datetimes**. Floating times are allowed. Recurrence is expanded in the event's own TZ.

## 3. Edge-case catalogue (each becomes tests)
- **Time:** DST gaps/overlaps (2:30am doesn't exist / happens twice), Feb 29 and month-end recurrences, EXDATE/RECURRENCE-ID overrides, editing "this vs all future" occurrences, user changes system TZ, all-day events spanning TZs.
- **Sync:** Google `410 Gone` → full resync; Graph expired delta → resync; 429/5xx → exponential backoff + jitter; token revoked → re-auth prompt without data loss; remote delete vs local edit; duplicate pushes after a crash (idempotency keys); clock skew.
- **OAuth:** port in use, browser not available, user closes the tab, state mismatch, **Google "Testing" apps' refresh tokens expire after 7 days** (needs a publishing decision), the client "secret" for desktop apps isn't secret.
- **Storage:** crash mid-write (transactions), two mimir instances (WAL + busy_timeout + lock), forward-only migrations with tests from every past schema, corrupted DB → backup on startup.
- **UI:** terminal resized to tiny dimensions, no mouse support, no color, very long titles, 10k tasks (virtualized list), pasting 1 MB of text, wide/emoji chars.
- **ICS:** line folding at 75 octets, escaped commas/semicolons, VTIMEZONE vs IANA names, Windows TZ names from Outlook, malformed files (be lenient on read, strict on write).

## 4. TDD strategy
- Red → green → refactor, one behavior at a time, starting at `domain` (fast, pure).
- **Unit:** table-driven tests for the quick-add parser, filters, recurrence, and normalization.
- **Property-based:** ICS round-trip (`parse(serialize(x)) == x`), recurrence invariants, sync convergence (random op sequences on both sides converge).
- **Golden/snapshot:** ICS fixtures (RFC 5545 examples plus real Google/Outlook exports), and TUI rendering at fixed sizes, in both color and NO_COLOR modes.
- **Contract tests:** one suite that runs against `FakeProvider` *and* (opt-in, with credentials) real Google/Graph.
- **HTTP:** record/replay fixtures for provider clients, with tokens scrubbed.
- **Integration:** SQLite on a temp file, migrations from each prior version.
- CI: lint + tests + race detector/sanitizers on every push.

## 5. Language comparison for this feature set
| Concern | Go | Rust | Python |
|---|---|---|---|
| TUI | Bubble Tea (Elm arch) + Lip Gloss; ANSI colors are first-class | Ratatui + crossterm; `Color::Reset` / indexed colors | Textual (CSS-like); its own themes by default, with an ANSI theme available |
| Theme detection | termenv/lipgloss: profile + bg detection | crossterm + manual OSC query | built into Textual, less control |
| OAuth | `golang.org/x/oauth2` (mature, PKCE) | `oauth2` crate (good, async plumbing) | `google-auth-oauthlib`, `msal` (official) |
| Google/Graph SDKs | official Google Go client; Graph SDK heavy, use REST | REST by hand (reqwest) | official for both |
| ICS / RRULE | `arran4/golang-ical`, `teambition/rrule-go` | `icalendar`, `rrule` crates | `icalendar`, `dateutil.rrule` (best-in-class) |
| SQLite + FTS5 | `modernc.org/sqlite` (pure Go) | `rusqlite` (bundled) | stdlib `sqlite3` |
| Keychain | `zalando/go-keyring` | `keyring` crate | `keyring` |
| Testing | built-in `testing`, `teatest`, `rapid` | built-in + `insta`, `proptest` | `pytest`, `hypothesis`, `syrupy` |
| Learning depth | interfaces, concurrency, explicit errors | ownership, traits, async, types (deepest) | fastest iteration, dynamic typing |
| Distribution | single static binary | single binary | needs an interpreter (uv tool / pipx) |

**Recommendation: Go.** It has the best ecosystem fit for exactly these features (Bubble Tea's ANSI theming, x/oauth2, official Google client, built-in testing), and you'll ship faster and spend learning time on architecture, sync and TDD rather than fighting the borrow checker in async OAuth code. **Rust** is the pick if the main goal is learning language internals; expect roughly 2× the time. **Python** is the fastest to build, but teaches the least about types and compilation.

## 6. Milestones (each fully test-driven)
0. `git init`, `go mod init`, Makefile, golangci-lint, GitHub Actions, config/XDG paths, slog logging
1. Domain: Task, statuses, priorities, subtasks, dependencies (cycle detection), quick-add parser, filters
2. SQLite repo + migrations + FTS5 + trash + undo/redo commands; WAL/locking for multiple instances
3. Non-interactive CLI over the same app layer. This lets the core run end-to-end before any UI work
4. TUI shell: layout engine + breakpoints, vertical tab sidebar, list/detail/editor, keymap config, mouse (hit-testing, click/double-click, wheel, context menu, drag-resize/reorder/drop), virtualized list, fuzzy search, command palette, settings tab
5. Theme system (ANSI default, NO_COLOR, bg detection, user themes)
6. Text pipeline: normalize, Glamour render, `$EDITOR`, grapheme-width wrapping
7. Natural-language dates + RRULE recurrence (property tests)
8. ICS import/export + URL subscriptions; todo.txt/JSON/Markdown formats
9. Calendar pane (toggle, agenda/week/month)
10. OAuth (PKCE loopback, device code) + keychain; Google Tasks + Calendar
11. Microsoft Graph: To Do + Calendar
12. Two-way sync engine (outbox, delta tokens, conflicts) + convergence property tests
13. Live reload (DB change detection), reminders (`mimir notify` + launchd), time tracking/Pomodoro
14. Stats view, plugin/hook scripts

## Verification
- `go test ./... -race` and `golangci-lint run` are green at each milestone; every feature starts with a failing test. Coverage concentrates on domain, sync and ics.
- Manual: run the TUI under Catppuccin, a light theme, `NO_COLOR=1`, and `TERM=dumb`; resize to 40×10.
- Import real Google/Outlook `.ics` exports. Do a sync round-trip against test calendars, editing on both sides to force a conflict.

## 7. Progress & decisions log
This section is updated as work lands, so any machine or session can pick up where the last one stopped.

### Milestone 0 status
- [x] `git init`, `go mod init` (module `github.com/Vishnu-MnM/mimir`, `go 1.27.3`)
- [x] `config.ResolvePaths` (XDG paths), table-tested
- [x] Makefile, golangci-lint v2 config, GitHub Actions CI
- [x] `CLAUDE.md` working agreement added; §7 is the handoff between machines
- [ ] **In progress:** `Paths.EnsureDirs()`. The red tests are on branch `ensure-dirs`
  (`internal/config/paths_test.go`). Next step: check out that branch and implement
  `EnsureDirs` in `paths.go` (validate `filepath.IsAbs` on all dirs first, then
  `os.MkdirAll(dir, 0o700)`, wrapping errors with the path). Merge to `main` once `make check` is green.
- [ ] slog logging to a file in `Paths.State` (inject an `io.Writer` so it can be tested)
- Then Milestone 1 (domain).

### Decisions
**Config paths (`internal/config`)**
- `ResolvePaths(getenv func(string) string) (Paths, error)`: `getenv` is injected (usually `os.Getenv`), so tests never touch the real environment.
- Each directory is resolved independently: an absolute `$XDG_*_HOME` value gives `<that>/mimir`; otherwise `$HOME/<default>/mimir`. Relative or empty XDG values are ignored, as the spec requires.
- `HOME` must be absolute, and it's only required when some directory falls back to it. On failure the function returns the zero `Paths` plus an error wrapping the sentinel `ErrNoHome`.
- XDG paths are used on macOS too (`~/.config/mimir`), not `~/Library/Application Support`, so `os.UserConfigDir()` is deliberately not used. `adrg/xdg` was considered and skipped: hand-writing it is more instructive and it's only ~40 lines.
- `EnsureDirs` creates missing directories (and missing parents) with mode `0700`, as the XDG spec says. It leaves directories that already exist, and their modes, untouched. It rejects relative or empty paths before creating anything, and fails if a file sits where a directory should be.

**Tooling**
- `make check` runs lint and then test. The other targets are `test` (`go test -race ./...`), `cover`, `lint`, `fmt` (`golangci-lint fmt`), `tidy` and `build` (`go build ./...` until `cmd/mimir` exists).
- golangci-lint v2: `default: standard`, plus revive, errorlint, misspell, unconvert and gocritic. The formatters are gofumpt and goimports (local prefix: the module path).
- CI (`.github/workflows/ci.yml`): lint runs on ubuntu-latest; `make test` runs on ubuntu-latest and macos-latest with `fail-fast: false`. The Go version comes from `go.mod`, and the golangci-lint version is pinned to match local installs. The repo is public, so Actions minutes are free.

**Conventions**
- Tests live next to the code in `_test.go` files in the same directory. Use `package x_test` for black-box tests where that's practical. Fixtures go in `testdata/`. End-to-end tests of the binary may get a top-level directory later.
- Every package gets a `// Package x ...` doc comment.
- Format with gofumpt (format-on-save, or `make fmt`); never align code by hand.
- Use `t.TempDir()` for filesystem tests, and `t.Helper()` in assertion helpers.
- Name a parameter `_`, or leave it unnamed, when it's unused (revive's `unused-parameter`).
