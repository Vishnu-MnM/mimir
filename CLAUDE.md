# mimir

A terminal (TUI) todo app in Go, built test-first. It's a **learning project**: the point is to understand how Go, its libraries and the architecture fit together, not just to ship.

- **Design and plan:** `DESIGN.md` (features, architecture, edge cases, milestones).
- **Current progress and decisions:** `DESIGN.md` §7. This is the source of truth for where work stopped.

## How we work
- **I write the code myself.** Show proposed code, tests and config in chat as complete, copy-pasteable snippets. Don't create or edit project files, and don't commit or push, unless I explicitly ask.
- **Go step by step.** One small piece at a time; wait for me before moving on to the next step.
- **Explain the why.** Say why it's written that way, which Go idioms it uses, and which choices I could reasonably make differently. Recommend one option instead of listing every alternative.
- **Strict TDD:** red → green → refactor. Tests come first. A test file that doesn't compile because the function is missing counts as red. When useful, give implementation hints so I can try writing it before seeing a full version.
- **Don't re-run tests I say are passing.** When I paste an error, explain the cause and the fix. Look at the file if that's needed to be accurate.
- **Git is mine, entirely.** Never run git commands that change anything (add, commit,
  branch, switch, push, merge, etc.), and don't suggest commit commands. Read-only
  checks (`git status`, `git log`, `git diff`) are fine.

## Start of every session
1. I work from two machines: Linux (`/home/sunny/Projects/mimir`) and macOS (`/Users/hermes/Projects/mimir`). Claude sessions and memory are **not** shared between them; this repo is the only shared state.
2. Check `git log`, `git status` and `DESIGN.md` §7 to see what changed, possibly on the other machine.
3. When a step finishes or a decision is made, update §7 (once I agree), so the other machine's session sees it.

## When I say "wrapping up"
1. Draft the `DESIGN.md` §7 update (progress, decisions) and wait for my OK before editing.
2. Suggest a `CLAUDE.md` change only if the way we work has changed.
3. Run `git status` and warn me about anything uncommitted or unpushed, including
   the doc edits from steps 1–2.

## Commands
- `make check`: lint, then test. Run it before every commit.
- `make test` (`go test -race ./...`), `make lint`, `make fmt` (gofumpt + goimports), `make cover`, `make tidy`, `make build`.
- `golangci-lint config verify` checks `.golangci.yml`.
- When the formatter complains, run `make fmt`; never align code by hand.

## Things to keep in mind
- **The repo is public.** Never commit secrets, tokens, `.env` files, or machine-specific absolute paths in code.
- Code must work on both Linux and macOS (CI runs tests on both). Use the XDG paths on macOS too.
- Conventions such as test layout, package doc comments, injected dependencies and `t.TempDir()` are listed in `DESIGN.md` §7.
