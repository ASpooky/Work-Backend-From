# Feedback Backlog

Last synthesized: 2026-08-15, from personas: Rina, Kenji, Yui, Sora, Haruka, Jun, Ai, Taku, Nozomi, Marina

## Open

### Task goal-picker doesn't scale (#5)
- Plain `<select>` resets to blank on every visit, no memory of the last-used goal for a daily heavy user — raised by: Rina — *partially fixed 2026-08-15: RegisterPage now remembers the last-used goal via localStorage, and shows each goal's remaining days inline.*
- Same `<select>` becomes an unusable long title-only dropdown once a user has ~10 goals, with no search/filter — raised by: Nozomi — *still open*

### WeekPathView doesn't scale with goal count (#7)
- One table row per goal with no cap, filter, or pin; 5-10 goals produce an endless scroll with no way to focus — raised by: Yui, Nozomi

### Weekly view has no trend/export/history (#9)
- No trend graph/sparkline across weeks; the view only ever shows the current week — raised by: Sora
- No CSV/JSON export exists — raised by: Sora
- Backend `DailyTask` stores no completion timestamp, so there's no raw data to build a trend feature on without a schema change — raised by: Sora
- (The original numerator/denominator and zero-vs-all-missed complaints this issue also raised are moot as of 2026-08-16: the weekly % display they were about was removed entirely, see Resolved below.)

### Multi-goal management gaps (backend) (#14)
- No goal edit/delete/priority field in the backend usecases; a goal's title is immutable after creation and there's no priority-based ordering — raised by: Nozomi
- No goal-list or single-goal focus view to orient across multiple concurrent goals — raised by: Nozomi
- `WeekPathView`'s goal-title column has no truncation-avoidance strategy for long/similar names — raised by: Nozomi

## Resolved

- **Today's task completion had no working toggle** (#1 root cause, 5 personas independently: Rina, Yui, Haruka, Taku, Ai) — resolved 2026-08-15. Added `dailytask.UpdateDailyTaskDoneUsecase`, `DailyTaskRepository.UpdateDone`, `PATCH /daily-tasks/{id}`, and a real checkbox with `onChange` in `CalendarView.tsx`.
- **CORS blocked PATCH requests from the browser** (found live after shipping the toggle above, not a persona-review item) — resolved 2026-08-15. Widened `Access-Control-Allow-Methods` beyond `GET, POST, OPTIONS`.
- **CalendarView screen order (#1)** — closed as won't-fix 2026-08-15: product owner explicitly requested week-view-first, today's-tasks-below, overriding the persona recommendation.
- **RegisterPage form structure (#2)** — resolved 2026-08-15: task-add form now renders above goal-creation.
- **Missing labels / placeholder-only form fields (#3)** — resolved 2026-08-15: every RegisterPage input now has a real `<label>`.
- **Language/tone inconsistency (#4)** — resolved 2026-08-15: placeholders, buttons, the mode select, and empty-state strings are Japanese throughout.
- **WeekPathView decoration vs. information density (#6)** — resolved 2026-08-15: rebuilt as a semantic `<table>` with stepper nodes (done/missed/today-pending/upcoming/milestone) adapted from a reference mockup, replacing the calc()-positioned card+line lanes.
- **Settings AI-integration section looks live but isn't (#11)** — resolved 2026-08-15: added a visible "未実装" badge next to the heading.
- **Sidebar visual identity inconsistency (#12)** — resolved 2026-08-15: active nav item now shows an accent-colored border tied to the shared `--accent` token.
- **Accessibility gaps (#13)** — resolved 2026-08-15: aria-labels on week nav buttons, ~44px touch targets, brightened Sidebar contrast to clear WCAG AA, and (via #6's rebuild) real table semantics for the goal × day grid.
- **Error handling & copy quality (#15)** — resolved 2026-08-15: raw exception text no longer reaches the UI (`errors.ts`'s `toUserMessage`).
- **Today's task list omits strict/want distinction (#17)** — resolved 2026-08-15: strict-mode tasks show a "必達" badge.
- **No relapse-recovery nudge (#16)** — resolved 2026-08-16: `entity.Goal.Postpone()` is now actually called. A missed day for a strict-mode goal is carried forward automatically (`CatchUpMissedTasksUsecase`, runs at server startup) instead of sitting unaddressed.
- **Failure-state ("未") styling — contested read (#8)** — resolved 2026-08-16 by reframing, not picking a side: once missed days are actively carried forward (#16), there's no permanent failure state left to argue about styling for. The transient "missed" node (visible only until the next reconciliation) is now a muted dot instead of a red ✕.
- **Progress % as gamification / failure framing (#10)** — resolved 2026-08-16: removed the weekly % bar and N/M total entirely, since a weekly evaluation score stopped making sense once missed days became postponements rather than failures.
