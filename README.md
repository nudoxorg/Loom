# Loom

Shared context for coding agents.

When you run multiple AI coding agents on the same codebase — Claude Code, Cursor, Codex, or anything else — they have no awareness of each other. One agent can overwrite what another is working on. There's no shared history of what happened or why. Loom fixes that.

Loom is a small, local CLI tool that gives agents a shared event log and a coordination layer. It requires no account, no server, no VS Code fork, and nothing running in the background unless you want it to. It just works from the command line.

---

## Mental model

Loom tracks two things:

**Events** are an append-only log. Every time something meaningful happens — an agent refactors a module, you make an architectural decision, a bug gets fixed — that gets logged. Events are permanent and ordered. They answer the question: *what has happened in this project?*

**Claims** are temporary flags. When an agent starts working on a file or path, it claims it. Other agents can see what's claimed and avoid stepping on in-progress work. When the work is done, the claim is released — but it stays in the event history. Claims answer the question: *what is being worked on right now?*

Both are scoped to the project by default. Loom determines the project root by walking up from the current directory looking for a `.git` folder — the same way most tools do. No init step required.

---

## Storage

Everything lives under `~/.loom`, outside your project repo. Nothing touches your working directory, nothing needs a `.gitignore` entry.

```
~/.loom/
├── settings.json                  # global settings (CLI-managed via `loom config`)
│
├── projects/
│   └── <slug>/                    # slug = slugified absolute project path
│       └── loom.db                # SQLite — events + claims for this project
│
└── global/
    └── global.db                  # SQLite — the opt-in global knowledge hub
```

**Project scope is the default.** Any `loom` command run from inside a project resolves its slug from the current directory (walking up to find `.git`) and reads/writes `projects/<slug>/loom.db`.

**Global scope is opt-in only.** `loom global log` / `loom global show` explicitly target `global/global.db` — it is never touched by default project-scoped commands.

**Slug derivation** mirrors Claude Code's `~/.claude/projects/` convention: the absolute project root path is slugified (e.g. `/Users/alice/code/myapp` → `Users-alice-code-myapp`). This is deterministic and collision-proof — no registry or `loom init` step required.

Each project gets its own SQLite database. Two agents working in the same directory will always resolve to the same database with no coordination required.

---

## Installation

```bash
go install github.com/nudoxorg/loom/cmd/loom@latest
```

Make sure `~/go/bin` is on your `$PATH`:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

Add that line to your `~/.zshrc` or `~/.bashrc` to make it permanent.

---

## Usage

### Events

```bash
# Log an event
loom log "refactored auth middleware"

# Show recent events
loom show
```

### Claims

```bash
# Claim a path you're working on
loom claim internal/auth

# See what's currently claimed
loom status

# Release when done
loom release internal/auth
```

Re-claiming a path you already hold refreshes it instead of adding a duplicate row, and any claim left unreleased for 4 hours expires automatically — useful since automated claiming (e.g. from a hook) has no natural "done" signal the way running `loom release` does.

### Global context

Global context is opt-in — a separate event log not tied to any project. Useful for cross-project notes or decisions that span multiple repos.

```bash
loom global log "switching all projects to Postgres"
loom global show
```

To see everything happening across *every* project Loom knows about at once — active claims and recent events, each labeled with which project they're from, merged with the global log above — use `loom global all`:

```bash
loom global all
loom global all --limit 50
```

This doesn't require anything to be logged manually; it reads directly from every project's own `loom.db` under `~/.loom/projects/`.

### Config

```bash
loom config get <key>
loom config set <key> <value>
loom config list
```

Supported keys: `default_agent` (used on every `log`/`claim`/`release`/`global log` unless overridden by MCP client identity — see below) and `default_limit` (default row count for `show`/`global show`, overridable per-call with `--limit`).

### Hooks

Agents forget to check Loom unless you remind them — so instead of relying on you to say it every session, Loom can generate project-scoped Claude Code hooks that do it automatically:

```bash
loom hooks install
```

This writes into the current project only — never `~/.claude/`, never anything outside the repo:

- `.claude/hooks/loom-remind-sessionstart.sh` and `.claude/hooks/loom-remind-pretooluse.sh` — small, static scripts that just print a reminder
- an entry in `.claude/settings.json` wiring them to the `SessionStart` and `PreToolUse` (`Write|Edit|MultiEdit`) hook events, merged into whatever's already in that file rather than overwriting it

Both hooks only inject a static reminder into context — "this project uses Loom, use the MCP tools" — pointing the agent at `loom_status`/`loom_claim`/etc. They never call `loom` themselves. `SessionStart` fires once per session; `PreToolUse` fires again before every file edit, so the reminder survives context drift instead of only being said once at the top and forgotten.

Safe to re-run: an unchanged script isn't rewritten, and an event already wired to Loom's hook isn't duplicated. Also available as the `loom_hooks_install` MCP tool, so an agent can set this up for a project itself when asked to.

---

## MCP server

Every Loom operation above is also exposed as an MCP tool over stdio — no network exposure, no accounts, same local trust model as the CLI. This is how agents use Loom directly instead of shelling out to `loom` themselves.

```bash
loom mcp
```

Point your agent's MCP client config at the `loom` binary, e.g.:

```json
{
  "mcpServers": {
    "loom": {
      "command": "loom",
      "args": ["mcp"]
    }
  }
}
```

Tools exposed: `loom_log`, `loom_show`, `loom_claim`, `loom_release`, `loom_status`, `loom_global_log`, `loom_global_show`, `loom_global_all`, `loom_config_get`, `loom_config_list`, `loom_hooks_install`. (`loom config set` stays CLI-only — global settings changes require a human at the terminal.)

The server ships with detailed instructions in the MCP `initialize` response — the mental model, when to claim/release, and how to write a log message that's actually useful to the next agent. Any MCP-aware client surfaces these automatically, so there's nothing extra to read or configure.

Events and claims created via MCP are attributed to the connecting client's own reported identity (e.g. `claude-code`, `cursor`) automatically, falling back to `default_agent` only if a client doesn't report one — no `agent` argument to pass, no config to keep in sync per client. Since every client we've seen reports that same name for every session, the server also appends a random ID generated once per `loom mcp` process, so two concurrent sessions of the same client (e.g. two Claude Code windows) stay distinguishable instead of silently sharing — and potentially releasing — each other's claims.

---

## Why not just use X?

**Traycer / other agent IDEs** — these require installing a specific editor or environment. Loom is editor-agnostic and works with any agent that can run a CLI command.

**A shared file in the repo** — anything in the repo risks conflicts, accidental commits, and `.gitignore` noise. Loom keeps everything in `~/.loom`.

**Agent-native memory** — per-agent memory is siloed. Claude Code doesn't know what Cursor just did. Loom is the shared layer across all of them.

---

## Roadmap

Config, the local MCP server, and project-scoped hooks are done (see above). What's next, in priority order — see `IDEAS.md` for full detail on each:

1. **Agent-to-agent thought sharing** — let an agent ask *why* a path was implemented a certain way and get another agent's reasoning, not just a diff
2. **AGENTS.md generation + Markdown export** — human/fallback-facing snapshots of project state for agents without MCP support, and for sharing or onboarding
3. **Daemon** and **Git integration** — background automation for auto-logging events; nice-to-have, not load-bearing

---

## License

[Blue Oak Model License 1.0.0](LICENSE.md)
