# Work-Backend-From

A personal goal-tracking app: break a long-term goal down into daily tasks, work through them one day at a time, and let an AI coach help you set the goal up and revisit it later.

Single-user, no authentication — built for personal / NAS-hosted use.

## Concept

A goal (title, detail, an achievement condition, a deadline) turns into a straight line of daily tasks connecting today to that deadline. Two commitment modes:

- **strict** — missing a day doesn't fail the goal, it postpones it: the whole remaining schedule and the deadline shift forward by one day. How many times this has happened is tracked as the goal's postpone count.
- **want** — missing a day just... doesn't move anything. No pressure.

Goals live inside **workspaces** (e.g. "private", "work") so unrelated goals don't crowd each other, with an "all workspaces" view when you want the full picture at once.

## Features

- **Calendar** — a week-at-a-time view of every goal as a row of stepper nodes (done / missed-and-postponed / upcoming / milestone), plus a checklist of just today's tasks.
- **Goals list** — every goal in the current workspace (or across all of them), independent of what's scheduled this week.
- **Goal detail** — achievement rate, done/scheduled counts, postpone count, days remaining; edit the goal directly.
- **AI goal planning** — chat with an AI coach (Gemini) to shape a new goal's intensity and deadline, then have it propose a goal plus a recurring daily-task schedule; review and edit before anything is created.
- **AI goal review** — the same idea for an *existing* goal: chat about what needs to change (extend the deadline, adjust the condition, ...) with the AI already aware of the goal's current state and progress, then review its proposed revision before saving it.
- **AI summary** — a short natural-language read on how a goal is actually going, given its real progress data.
- **Conversations persist** — every AI chat is saved per workspace (or per goal, for goal-review threads) so past context isn't lost between sessions.

AI features are optional: the backend runs fine without a Gemini API key, it just skips registering the `/ai/*` routes.

## Tech stack

- **Backend**: Go, `net/http`, SQLite (`modernc.org/sqlite`, pure Go / no CGO — keeps cross-compilation simple for a future NAS or desktop target)
- **Frontend**: React 19, TypeScript, Vite, `react-router-dom`
- **AI**: Google Gemini API, called directly over REST (not the CLI) for real JSON-schema-constrained structured output

## Running it

### Backend

```sh
go run ./src
```

Listens on `:8080`, creates/migrates `app.db` in the working directory on first run, and seeds a default `private` workspace.

To enable the AI features, put a Gemini API key in `.env` (or `src/.env`) before starting:

```
GEMINI_API_KEY=...
GEMINI_MODEL=gemini-3.7-flash   # optional, this is the default
```

### Frontend

```sh
cd frontend
npm install
npm run dev
```

Expects the backend at `http://localhost:8080`.

### Seeding sample data

```sh
go run ./cmd/seed
```

Populates `dev.db` (a separate database, never `app.db`) with a few example goals and a plausible task history. Idempotent — safe to re-run. Point it at a different file with `-db path`; the default deliberately isn't `app.db` so this can't accidentally seed real data.

## Screenshots

_TODO_

## Layout

```
src/
  entity/            domain model (Goal, DailyTask, Workspace, Conversation, ...)
  usecase/            application logic, one subpackage per aggregate
  repository/sqlite/  persistence
  handler/httpapi/    HTTP layer
  infra/              clock, id generation, the Gemini client
frontend/src/
  pages/               routed views
  components/          shared UI (calendar, sidebar, ...)
cmd/seed/              sample-data generator
```

See `entity.md` for the full data model.
