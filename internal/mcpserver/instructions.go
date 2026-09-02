package mcpserver

// instructions is returned to every connecting client in the MCP initialize
// response. It's the only "documentation" an agent is guaranteed to see, so
// it needs to stand alone: what Loom is, what each tool does, and how to use
// them well — not just what arguments they take.
const instructions = `Loom gives coding agents shared awareness of a project: what has happened, and what's being worked on right now. You are one of potentially several agents (or a human) working on this codebase — Loom is how you avoid stepping on each other and how you learn what came before you, without re-deriving it from a diff.

MENTAL MODEL
- Events are a permanent, append-only history: what happened and why. They're the project's memory — anything another agent (or a future you) would want to know later belongs here.
- Claims are temporary "I'm working on this" flags on a path. They are advisory, not a lock: Loom does not block anyone from editing a claimed path. Claiming only makes the conflict visible so agents can coordinate instead of silently colliding. The discipline is on you.
- Every project-scoped tool takes a required cwd argument: your actual current working directory right now, not wherever the session started. This server is a long-lived process whose own working directory never changes, so it cannot infer where you are — you must state it on every call. Loom resolves the project from cwd by walking up for a .git folder, or, if none is found, treating cwd itself as an ad-hoc project root. You never pass a project identifier directly.
- Every call is attributed automatically to your own client identity (e.g. "claude-code", "cursor"). There is no "agent" argument on any tool, and nothing to configure.

FULL TOOL REFERENCE

Project-scoped (this project only — cwd is required on every one of these):
- loom_log(cwd, message) — record a permanent event.
- loom_show(cwd, limit?) — recent events for this project, newest first. Omit limit to use the configured default (usually 20).
- loom_claim(cwd, path) — flag a path as in progress.
- loom_release(cwd, path) — clear a claim you made.
- loom_status(cwd) — currently active claims for this project.

Cross-project:
- loom_global_log(message) — a note that is not scoped to any one project, e.g. "migrating every service to Postgres." A separate stream from loom_log — writing here does not show up in loom_show.
- loom_global_show(limit?) — reads back that manual cross-project stream only.
- loom_global_all(limit?) — a read-only rollup of active claims and recent events across every project Loom has touched on this machine, plus the global stream above, merged into one timeline and labeled by project. This already includes the current project's own data, so it's the right default call at the start of a session (see RECOMMENDED WORKFLOW below) — one call gives you full situational awareness, not just this project. Reach for loom_show/loom_status instead when you specifically want to filter to just this project without cross-project noise, e.g. mid-session.

Config (read-only from here):
- loom_config_get(key) — a single setting's value (default_agent, default_limit).
- loom_config_list() — every setting.
- There is no config-writing tool. Changing settings is deliberately a human/CLI action (loom config set), not something an agent does on its own.

RECOMMENDED WORKFLOW
1. Start of a session, before making changes: call loom_global_all. It covers this project's active claims and recent history plus everything happening in parallel elsewhere — other agents, other directories, the same directory — in one call. This is how you find out what's going on before you touch anything, not just in this codebase but across every project Loom knows about.
2. Before editing a path, loom_claim it. If it's already claimed by someone else, that's a signal to coordinate or work elsewhere — Loom won't stop you either way.
3. When you're done with a path — including if you're abandoning it, not just finishing it — loom_release it. A stale claim left behind is worse than no claim at all: it tells the next agent something is in progress when it isn't.
4. Log meaningful events with loom_log as you go: decisions and their reasoning, not routine noise. Every entry is permanently tagged with your identity and a timestamp, so treat it as a message to whoever picks up this code next.

WRITING GOOD LOG MESSAGES
This is the single highest-leverage thing you can do with Loom. A log message should answer "why," not "what" — the diff already shows what changed.

  Weak: "updated auth.go"
  Good: "switched auth middleware to JWT because sessions weren't surviving the new load balancer"

  Weak: "fixed bug"
  Good: "fixed a race in claim release — two agents releasing the same path at once could both succeed and double-log the release event"

  Weak: "tried caching"
  Good: "tried an in-memory cache for project lookups, reverted — invalidation got complicated for no measurable win at this scale"

That last example matters as much as the other two: a documented dead end saves the next agent from re-trying it.`
