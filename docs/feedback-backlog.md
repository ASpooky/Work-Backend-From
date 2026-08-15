# Feedback Backlog

Last synthesized: 2026-08-15, from personas: Rina, Kenji, Yui, Sora, Haruka, Jun, Ai, Taku, Nozomi, Marina

## Open

Each theme below is filed as its own GitHub issue (#1-#17) for tracking; this file stays the source-of-truth for the raw persona reasoning behind each one.

### CalendarView screen order buries today's tasks (#1)
- WeekPathView (decorative goal lanes) renders above the today's-task list, delaying the one thing time-pressed/strict users need first on open — raised by: Rina, Haruka

### RegisterPage form structure (#2)
- The full goal-creation form (all fields required) is stacked above the daily task-add form, forcing a scroll/read past long-term-planning UI just to log one quick task — raised by: Haruka, Jun, Yui
- Task's goal `<select>` shows only "Select a goal" and dead-ends with no onboarding nudge when the user has no goals yet — raised by: Yui
- Goal creation form requires all 5 fields including a free-text "Achievement condition," a high-friction first step for someone prone to giving up — raised by: Taku

### Missing labels / placeholder-only form fields (#3)
- Every input on `RegisterPage.tsx` uses placeholder text with no `<label>`/`aria-label`; text disappears once focused/typed, and screen readers announce no field name — raised by: Kenji, Yui, Ai

### Language/tone inconsistency (#4)
- `RegisterPage.tsx` headings are Japanese while placeholders/buttons stay English (Title/Detail/Achievement condition/Create Goal/Add Task, etc.) — raised by: Kenji, Marina
- Empty-state strings ("No tasks for today." / "No goals yet.") remain in English against an otherwise Japanese UI — raised by: Marina
- Goal mode `<select>` exposes raw internal values `strict`/`want` instead of localized labels — raised by: Marina

### Task goal-picker doesn't scale (#5)
- Plain `<select>` resets to blank on every visit, no memory of the last-used goal for a daily heavy user — raised by: Rina
- Same `<select>` becomes an unusable long title-only dropdown once a user has ~10 goals, with no due-date/priority cue — raised by: Nozomi

### WeekPathView decoration vs. information density (#6)
- `calc()`-based line/arrow connectors between goal cards are purely decorative, adding visual noise without new information — raised by: Rina, Jun
- Cards render at ~11px font with 6-7px padding and a single-character "済/未" badge, prioritizing density over legibility (the connector concept itself is seen as a good differentiator) — raised by: Kenji
- No numeric summary (completion rate, streak count) per goal — only the one-character "済/未" badge — raised by: Sora
- The same "済"/"未" color+text pairing is read two opposite ways: Ai flags it as a genuine accessibility win (status isn't color-only); Jun calls it redundant, over-decorated for a single checkmark's worth of information — raised by: Ai (positive), Jun (negative)
- A reference design (`goal_tracker_variable_tasks_and_goal_progress.html`) replaces cards+connectors with a stepper/node timeline and a real numeric cumulative-progress column — candidate replacement, not just a patch

### WeekPathView doesn't scale with goal count (#7)
- One `.lane` stacks per goal with no cap, filter, or pin; 5-10 goals produce an endless scroll with no way to focus — raised by: Yui, Nozomi

### Failure-state ("未") styling — contested read (#8)
- Yui reads the red `border-left` + muted "未" text as appropriately restrained (no big × or warning background) for a streak break
- Taku reads the same styling as discouraging: muted "済" vs. danger-colored "未" repeated every time a slacked-off week is reopened, an asymmetry that works against people trying to restart
— raised by: Yui (positive), Taku (negative)

### Progress % trustworthiness and depth (#9)
- `weekProgress` shows a percentage and bar with no numerator/denominator breakdown, so identical percentages can mean different underlying counts — raised by: Sora
- A zero-task week and an all-missed week are visually indistinguishable — raised by: Sora
- No trend graph/sparkline; `GoalProgressList` only ever shows the current week's snapshot — raised by: Sora
- No CSV/JSON export exists — raised by: Sora
- Backend `GoalCalendar.Days` stores only status+content, no completion timestamp/duration, so there's no raw data to build a future trend feature on — raised by: Sora

### Progress % as gamification / failure framing (#10)
- The weekly % bar is the kind of gamification a minimalist distrusts, imposing a weekly evaluation metric — raised by: Jun
- Same % bar has no positive counterpart (e.g. streak length); a bad week just shows a falling number, functioning as pure "failure visualization" — raised by: Jun, Taku

### Settings AI-integration section looks live but isn't (#11)
- The `sk-...` key input and save button read as a working feature; the "not implemented" note is plain body text with no visual distinction — raised by: Kenji, Jun
- Counterpoint: the disclosure text itself ("not yet referenced by any feature") is honest and manages expectations well — raised by: Yui

### Visual identity inconsistency (sidebar vs. main theme) (#12)
- `Sidebar.css` hardcodes its own dark-navy palette (`--sidebar-bg: #1a1d2e`, etc.) with no connection to `index.css`'s Linear-style `--accent: #5e6ad2` token system, reading as two mismatched templates stitched together — raised by: Kenji

### Accessibility gaps (#13)
- Week-navigation buttons render as literal "<"/">" text with no `aria-label` — raised by: Ai
- `Sidebar`'s `--sidebar-text-dim` (#7d8099) on `--sidebar-bg` (#1a1d2e) measures ~4.1:1, under WCAG AA's 4.5:1 — raised by: Ai
- `WeekPathView`'s goal × day grid is plain div+CSS grid with no `table`/`role="grid"`/`scope` semantics — raised by: Ai
- `.period-nav` buttons and `.task-list li` have no padding/min-height, missing ~44px touch targets — raised by: Haruka

### Multi-goal management gaps (backend) (#14)
- No goal edit/delete/priority field in the backend usecases; a goal's title is immutable after creation and there's no priority-based ordering — raised by: Nozomi
- `Sidebar.tsx` only switches between Calendar/Register/Settings; no goal-list or single-goal focus view to orient across multiple concurrent projects — raised by: Nozomi
- `GoalProgressList`'s `.goal-progress-title` is a fixed `8rem` + ellipsis, truncating long goal names indistinguishably — raised by: Nozomi

### Error handling & copy quality (#15)
- Errors surface via `setError(String(err))`, leaking raw exception text to the end user — raised by: Marina

### No relapse-recovery nudge (#16)
- `entity/goal.go`'s `Postpone()` (strict-mode relief logic) is never called from any usecase, and no UI affordance encourages resuming after a missed day — raised by: Taku

### Today's task list omits strict/want distinction (#17)
- `Goal` has `mode: 'strict' | 'want'`, but the today's-task list only renders `content`+`done`, so users can't tell must-do items from optional ones at a glance — raised by: Rina

## Resolved

- **Today's task completion had no working toggle** — resolved 2026-08-15. Added `dailytask.UpdateDailyTaskDoneUsecase`, `DailyTaskRepository.UpdateDone`, `PATCH /daily-tasks/{id}`, and a real checkbox with `onChange` in `CalendarView.tsx` (raised by: Rina, Yui, Haruka, Taku, Ai). Verified via `go test ./...` and a live PATCH against the running server.
- **CORS blocked PATCH requests from the browser** (not a persona-review item — found live after shipping the toggle above) — resolved 2026-08-15. `httpapi.WithCORS`'s `Access-Control-Allow-Methods` only listed `GET, POST, OPTIONS`; the preflight for `PATCH /daily-tasks/{id}` succeeded but the browser then blocked the real request, surfacing as `TypeError: Failed to fetch`. Added `PATCH, PUT, DELETE` to the allow-list. Verified via a manual preflight (`OPTIONS` + `Access-Control-Request-Method: PATCH`) against the running server.
