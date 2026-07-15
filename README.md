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
│       ├── loom.db                # SQLite — events + claims for this project
│       └── settings.json          # per-project setting overrides (optional)
│
└── global/
    ├── global.db                  # SQLite — the opt-in global knowledge hub
    └── settings.json              # settings specific to global context
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

### Global context

Global context is opt-in — a separate event log not tied to any project. Useful for cross-project notes or decisions that span multiple repos.

```bash
loom global log "switching all projects to Postgres"
loom global show
```

### Config

```bash
loom config get <key>
loom config set <key> <value>
loom config list
```

---

## Why not just use X?

**Traycer / other agent IDEs** — these require installing a specific editor or environment. Loom is editor-agnostic and works with any agent that can run a CLI command.

**A shared file in the repo** — anything in the repo risks conflicts, accidental commits, and `.gitignore` noise. Loom keeps everything in `~/.loom`.

**Agent-native memory** — per-agent memory is siloed. Claude Code doesn't know what Cursor just did. Loom is the shared layer across all of them.

---

## Roadmap

- **MCP server** — expose all Loom operations as MCP tools so agents can interact with Loom directly without running CLI commands
- **AGENTS.md generation** — for agents that can't make live tool calls, Loom regenerates a section of `AGENTS.md` from current state on each update
- **Daemon** — optional background process that watches the filesystem and git history to auto-log events without manual `loom log` calls
- **Git integration** — automatically log events from git history, correlate claims with commits, surface recent changes as context
- **Markdown export** — generate a human-readable snapshot of project context as a Markdown document, useful for sharing state or onboarding a new agent

---

## License

[Blue Oak Model License 1.0.0](LICENSE.md)
