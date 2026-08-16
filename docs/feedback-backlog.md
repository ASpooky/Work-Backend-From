# Feedback Backlog

Last synthesized: 2026-08-17, from personas: Rina, Kenji, Yui, Sora, Haruka, Jun, Ai, Taku, Nozomi, Marina

## Open

### 今日のタスクがWeekPathViewの下に埋もれる
- `CalendarView.tsx` renders `WeekPathView` (up to 7 goals × 7 days) above the `今日のタスク` checklist — the one section that answers "what do I do right now" requires scanning/scrolling past the week grid first — raised by: Rina, Yui, Haruka
- On phone widths this is worse than it looks: `.week-path-table { min-width: 40rem }` (`WeekPathView.css` line 31) forces horizontal scroll, and each goal row is `height: 72px` (line 98) — up to 7 rows before the today-checklist is even reached — raised by: Haruka
- Note: this is related to, but distinct from, the already-closed won't-fix "CalendarView screen order (#1)" — that decision was about week-view-first vs. today-first ordering, made explicitly by the product owner. This item is about the week view's own footprint being large enough to push the actionable list off first paint even under a week-first layout; it doesn't reopen the ordering decision, but is worth the product owner's attention as a distinct sizing/scroll problem.

### 今日のタスクリスト自体の使い勝手
- `daily_task_repository.go`'s `FindByDate`/`FindByDateAndWorkspaceID` (lines 92-113) both `ORDER BY created_at` only — no priority or strict-first ordering, and `CalendarView.tsx` renders `dayTasks.map` as-is with no client-side sort, so completed and incomplete tasks stay interleaved as you work through a list — raised by: Rina
- The "必達" badge (`task-item-strict`, `CalendarView.tsx` lines 94-98) is display-only; there's no way to filter to just strict tasks or pin them above "want" tasks — raised by: Rina
- No remaining-count / done-vs-total indicator on the today list — only an empty-state string when there are zero tasks (`CalendarView.tsx` line 103) — raised by: Rina

### その場でのタスク追加・初回ゴール作成の摩擦
- `CalendarView.tsx` has no inline "quick add" — adding a task requires navigating to the separate `/register` route (`App.tsx` lines 141-151), so a 10-second glance can check off an existing task but can't log a new one without leaving the screen — raised by: Haruka
- Creating a goal (`RegisterPage.tsx` lines 244-274) requires 4 mandatory free-text/date fields (タイトル, 詳細, 達成条件, 期限) with no defaults or skip path — since a task can't be logged without an existing goal, this is a hard gate in front of even the fastest quick-add flow — raised by: Haruka, Taku
- `RegisterPage.tsx` line 30 defaults new-goal `mode` to `'strict'` (必達) rather than `'want'` (努力目標), with no inline copy explaining what the difference means for how misses are handled — a first-time user is nudged into the highest-pressure framing by default — raised by: Taku
- 達成条件 (`RegisterPage.tsx` line 263-270) is a plain unvalidated text input — not a precision problem, a volume one: it's one more required field standing between a lazy/guilt-prone beginner and being able to save anything at all — raised by: Taku

### 達成率・残日数の定義がUIに出ない
- `get_goal_stats.go`'s `AchievementRate` (line 61) is well-scoped — only tasks on/before "now" count toward the denominator — but `GoalDetailPage.tsx` (lines 441-442) prints the bare `{achievementPercent}%` under "達成率" with no indication future-dated tasks are excluded from the denominator — raised by: Sora
- `daysRemaining` (`get_goal_stats.go` line 64) is computed against `goal.EndDate`, which `Goal.Postpone()` (`goal.go` lines 50-54) silently pushes forward by a day per missed strict-mode day. `GoalDetailPage.tsx` shows both `あと${days_remaining}日` and the postpone count, but never states that days-remaining is counted from the *moved* deadline, not the original one — raised by: Sora

### 完了データはあるが活用されていない（トレンド・履歴・エクスポート）
- `entity.DailyTask.CompletedAt` is now fully wired end-to-end (see Resolved) but `get_goal_stats.go` never reads it — only `Done bool` feeds `ScheduledCount`/`DoneCount`, so a timestamp that could drive streaks/time-of-day patterns is captured and unused — raised by: Sora
- No trend/history view exists anywhere in `frontend/src` — `GoalDetailPage.tsx`'s three stat cards are a single point-in-time snapshot, not a series, despite per-day `DailyTask` rows already being keyed by date — raised by: Sora
- No CSV/JSON export path exists in either the handler layer (`goal_handler.go`) or the frontend — raised by: Sora

### 「先延ばし回数」のフレーミング (multi-persona, related but distinct angles)
- `GoalDetailPage.tsx`'s `.goal-detail-stats` (lines 439-454) shows 先延ばし回数 as an equal-weight third stat-card next to 達成率 and 完了/予定, with no framing, every time the goal is opened — risks reading as a judgment tally rather than useful info — raised by: Yui
- Same stat, product framing angle: a notebook wouldn't compute and show back a "times you gave up" counter as a headline number — raised by: Jun
- Adjacent but separate: the `aria-label="先送り中"` on the missed-day node (`WeekPathView.tsx` line 168) literally reads "being put off" — a small word-choice concern for a guilt-sensitive user, independent of the stat card — raised by: Taku
- Also from Taku: `DayCell` still renders the task caption under missed days (`content && state !== 'no-task'` doesn't exclude `state === 'missed'`, `WeekPathView.tsx` line 153) — a missed day doesn't just fade, it keeps showing exactly what you didn't do
- These three angles point at the same general area (postpone-count/missed-day framing) but aren't the same fix — a synthesizer note, not a single ticket.

### AI機能の面数増加 (tension — record, don't auto-action)
- Sidebar routes to five top-level destinations (カレンダー/目標/登録/AIと計画/設定) for what's fundamentally one goal-tracking concept; カレンダー・目標・登録 are three different views onto the same goal data — raised by: Jun
- `GoalDetailPage.tsx` alone has three independent AI entry points — "AIと見直す" revision chat (lines 292-398), "AIによるサマリー" analyze button (line 112), plus `AIPlanPage.tsx`'s separate creation wizard — three flows to learn instead of one editable field — raised by: Jun
- `AIPlanPage.tsx`'s plan-review step (lines 241-321) is a full secondary editable-proposal UI (per-task 削除/戻す, editable goal fields) sandwiched between "chat" and "save" — raised by: Jun
- **Tension, not a directive**: this reads as feature-creep from a minimalist-persona lens, but the AI features themselves (goal review chat, AI plan creation, AI summary) were explicitly requested by the actual product owner in an earlier session — this is a real tradeoff between "fewer surfaces" and already-committed product direction, for a human to weigh, not something to action as "remove AI features."

### デザイントークンの一貫性
- `--accent`/Inter Variable/6-10px radii/hairline borders in `index.css` reads as an unmodified Linear/shadcn-style AI-scaffold default, not a distinct identity — raised by: Kenji
- `Sidebar.css` hardcodes its own permanent dark-navy palette (`--sidebar-bg: #1a1d2e`, etc.) that never responds to the light/dark tokens in `index.css` — in light mode this bolts a navy sidebar onto a white canvas instead of one coherent system — raised by: Kenji
- Border weight is inconsistently 0.5px (`.week-path`, `.stat-card`, `.goals-list-card`) vs. full 1px (`.ai-plan-task-list li`, `.workspace-switcher-menu`, all inputs/buttons) — raised by: Kenji
- Spacing has no visible scale — App.css/AIPlanPage.css/GoalDetailPage.css cycle through 0.15-0.85rem gaps with no 4/8px rhythm — raised by: Kenji
- The same accent-pill treatment is re-declared from scratch in five places (`.badge-enabled`, `.task-item-strict`, `.week-path-goal-workspace`, `.goals-list-card-mode.strict`, `.ai-conversation-list li button.active`) instead of one shared Badge class — raised by: Kenji
- Caption/label text repeatedly drops to 0.6-0.65rem (`.week-path-caption`, `.badge-unimplemented`, `.goals-list-reorder button`, `.task-item-strict`, `.goals-list-card-mode`) — raised by: Kenji (see also Ai's contrast finding on `.week-path-caption` below, same element from a different angle)

### 色だけに依存した状態表現・キーボード操作性
- `WeekPathView.tsx`'s `NodeIcon` (lines 158-193): `missed`, `today-pending`, and `upcoming` all render the identical dot glyph (`week-path-node-dot`), differentiated only by CSS border/fill color — a colorblind or grayscale user can't distinguish "missed" from "upcoming" by shape — raised by: Ai
- `.node-missed`'s border (`WeekPathView.css` line 156) uses `--border`, which is `rgba(15,16,20,0.09)` — 9% opacity — on `--bg-raised`, making arguably the most important state to notice the weakest visual presence of all six. **Tension**: this same low-contrast, unshouty styling is exactly what Yui and Taku independently praised as the right low-guilt choice (see Working well) — a fix here needs to add distinguishability (shape/icon, not necessarily color intensity) without reintroducing a red/alarming missed-state — raised by: Ai
- `.week-path-caption` (`WeekPathView.css` lines 203-210): `color: var(--text)` (`#6f7280`) at 0.65rem on `--bg-raised` (`#f6f6f8`) is reported at ≈4.4:1, under WCAG AA's 4.5:1 minimum for small text — raised by: Ai
- `WorkspaceSwitcher.tsx`'s trigger button (line 58) has no `aria-expanded`/`aria-haspopup`, and the opened menu (line 64) has no `role="menu"`/menuitem roles, no focus moved on open, and no Escape-to-close or outside-click handler — confirmed still present in current code — raised by: Ai

### 複数ゴール運用時の優先度・スケール
- `GoalsListPage.tsx`'s `handleMove` (lines 40-58) only swaps two adjacent goals per click, each firing two awaited API calls — moving a goal from position 10 to 1 takes 9 sequential click-and-wait cycles — raised by: Nozomi
- The ▲/▼ reorder controls are hidden entirely in all-workspaces mode (`{!isAllWorkspaces && (...)}`, `GoalsListPage.tsx` line 73) — priority is workspace-scoped only, so there's no way to rank goals across all clients/workspaces at once — raised by: Nozomi
- `WeekPathView`'s `DEFAULT_VISIBLE_GOAL_COUNT = 7` cap (see Resolved: row-capping shipped) is silent and priority-ordered — any goal ranked below 7th disappears from the daily calendar by default and only reappears via "もっと見る," which is exactly the deprioritized-but-still-active goal most likely to get missed — raised by: Nozomi
- In "すべて" mode, `GoalsListPage.tsx` still renders one flat `<ul>` with only a small workspace badge (line 104) — no section headers, grouping, or filter-by-workspace — at 10 goals across several clients this is still a badge-scan wall, not the triage view multi-project users need — raised by: Nozomi
- No month or multi-week overview anywhere — `CalendarView.tsx` drives `WeekPathView` off a single `anchor` with only prev/next-week navigation (lines 60-76), so comparing last week's delivery against this week requires manual back-and-forth rather than a stacked view — raised by: Nozomi
- Carryover from #14: `WeekPathView`'s goal-title column (`.week-path-goal-cell { white-space: nowrap }`, `WeekPathView.css` line 62) still has no truncation-avoidance (no `text-overflow`/`max-width`) for long or similar goal names — confirmed still present in current code

### 表記ゆれ・生の内部文字列の露出
- `Sidebar.tsx` line 25 (`<div className="sidebar-brand">Goal Tracker</div>`) and `App.tsx` line 114 (`<h1>Goal Tracker</h1>`) are the only English strings in an otherwise fully Japanese nav/chrome — raised by: Marina (both locations confirmed during verification; Marina's note cited the Sidebar instance)
- `AIPlanPage.tsx` line 173 surfaces the raw env-var name directly in user-facing copy: "...バックエンドに GEMINI_API_KEY 環境変数が設定されていないため、現在利用できません。" (`SettingsPage.tsx` line 48 has the same pattern, wrapped in a `<code>` tag but still exposing the literal var name to end users) — raised by: Marina
- Parenthesis style is inconsistent: `GoalDetailPage.tsx` line 348 and `AIPlanPage.tsx` line 243 use half-width `()` for review-flow headings, while `GoalDetailPage.tsx` lines 429-431 use full-width `（あと◯日）` for the same kind of aside elsewhere on the same page — raised by: Marina
- Empty-state punctuation is inconsistent: `AIPlanPage.tsx` line 202 "まだ会話がありません" has no closing "。" while `GoalsListPage.tsx` line 66 "まだ目標がありません。「登録」から作成してください。" is fully punctuated — raised by: Marina

## Working well (protect this)

- **Muted missed-state styling** — the gray dot + `opacity: 0.55` treatment for missed days (no red, no ✕) is explicitly the right low-guilt call, independently praised by Yui and Taku. A future "make failure more visible" pass should not reintroduce alarming color here — see the color/shape tension noted above under Ai's feedback for how to add distinguishability without violating this.
- **Absence of gamification** — Jun specifically checked `GoalDetailPage.tsx` and `AIPlanPage.tsx` for badges/streaks/confetti/celebratory copy and found none. That restraint is correct and should stay a constraint on future feature additions (including any future AI-surfaced "encouragement" feature, see Taku's re-engagement-copy point which pulls the opposite direction and should be weighed against this, not assumed to win).
- **`GoalsListPage.tsx` reorder buttons** — real `<button>` elements, `aria-label` with the goal title baked in, `disabled` at list boundaries rather than a no-op click. Ai flagged this as the template other controls in the app should follow (e.g. WorkspaceSwitcher's menu, above).
- **1-tap done toggle** — `CalendarView.tsx`'s checkbox-based done toggle (`.task-item`, min-height 2.75rem) with optimistic UI update is exactly the friction-free flow Haruka needs; no changes suggested.
- **Same-day quick-add, once a goal exists** — `RegisterPage.tsx`'s default "今日だけ" task-creation path is genuinely 2 fields + 1 submit; the friction Haruka and Taku flagged above is specifically the goal-creation gate in front of it, not this path itself.

## Resolved

- **Today's task completion had no working toggle** (#1 root cause, 5 personas independently: Rina, Yui, Haruka, Taku, Ai) — resolved 2026-08-15. Added `dailytask.UpdateDailyTaskDoneUsecase`, `DailyTaskRepository.UpdateDone`, `PATCH /daily-tasks/{id}`, and a real checkbox with `onChange` in `CalendarView.tsx`.
- **CORS blocked PATCH requests from the browser** (found live after shipping the toggle above, not a persona-review item) — resolved 2026-08-15. Widened `Access-Control-Allow-Methods` beyond `GET, POST, OPTIONS`.
- **CalendarView screen order (#1)** — closed as won't-fix 2026-08-15: product owner explicitly requested week-view-first, today's-tasks-below, overriding the persona recommendation. (See new "今日のタスクがWeekPathViewの下に埋もれる" item above for a related-but-distinct sizing concern this doesn't cover.)
- **RegisterPage form structure (#2)** — resolved 2026-08-15: task-add form now renders above goal-creation.
- **Missing labels / placeholder-only form fields (#3)** — resolved 2026-08-15: every RegisterPage input now has a real `<label>`. Re-verified 2026-08-17 against current `RegisterPage.tsx` — still holds.
- **Language/tone inconsistency (#4)** — resolved 2026-08-15: placeholders, buttons, the mode select, and empty-state strings are Japanese throughout. (Note: Marina's 2026-08-17 pass found two narrower, previously-unflagged English/leak strings — "Goal Tracker" brand text and the raw `GEMINI_API_KEY` name — tracked as new items above; those are new findings, not a regression of this resolution.)
- **WeekPathView decoration vs. information density (#6)** — resolved 2026-08-15: rebuilt as a semantic `<table>` with stepper nodes (done/missed/today-pending/upcoming/milestone) adapted from a reference mockup, replacing the calc()-positioned card+line lanes.
- **Settings AI-integration section looks live but isn't (#11)** — resolved 2026-08-15: added a visible "未実装" badge next to the heading.
- **Sidebar visual identity inconsistency (#12)** — resolved 2026-08-15: active nav item now shows an accent-colored border tied to the shared `--accent` token.
- **Accessibility gaps (#13)** — resolved 2026-08-15: aria-labels on week nav buttons, ~44px touch targets, brightened Sidebar contrast to clear WCAG AA, and (via #6's rebuild) real table semantics for the goal × day grid. (Note: Ai's 2026-08-17 pass found further, more specific a11y gaps — color-only state differentiation, WorkspaceSwitcher menu semantics, caption contrast — tracked as new items above; those are new findings in areas this pass didn't cover, not a regression.)
- **Error handling & copy quality (#15)** — resolved 2026-08-15: raw exception text no longer reaches the UI (`errors.ts`'s `toUserMessage`).
- **Today's task list omits strict/want distinction (#17)** — resolved 2026-08-15: strict-mode tasks show a "必達" badge.
- **No relapse-recovery nudge (#16)** — resolved 2026-08-16: `entity.Goal.Postpone()` is now actually called. A missed day for a strict-mode goal is carried forward automatically (`CatchUpMissedTasksUsecase`, runs at server startup) instead of sitting unaddressed.
- **Failure-state ("未") styling — contested read (#8)** — resolved 2026-08-16 by reframing, not picking a side: once missed days are actively carried forward (#16), there's no permanent failure state left to argue about styling for. The transient "missed" node (visible only until the next reconciliation) is now a muted dot instead of a red ✕.
- **Progress % as gamification / failure framing (#10)** — resolved 2026-08-16: removed the weekly % bar and N/M total entirely, since a weekly evaluation score stopped making sense once missed days became postponements rather than failures.
- **Task goal-picker doesn't scale (#5)** — fully resolved 2026-08-17. `RegisterPage.tsx` now (a) remembers the last-used goal via localStorage (`taskGoalId` initial state, line 32) and shows each goal's remaining days inline in the `<option>` label (lines 133-140, shipped 2026-08-15), and (b) once `goals.length > 6`, shows a live goal-name filter input above the `<select>` (lines 118-128, shipped this session, commit `f9d2c75`) that narrows the option list while always keeping the current selection visible even if it doesn't match the filter (lines 49-53).
- **WeekPathView doesn't scale with goal count (#7)** — resolved 2026-08-17: `WeekPathView.tsx` now caps rows at `DEFAULT_VISIBLE_GOAL_COUNT = 7` with a "もっと見る (+N)" / "折りたたむ" toggle (lines 11, 35-37, 80-84; commit `4e6ecaf`), so 5-10 goals no longer produce unbounded scroll. Note: this cap's own new failure mode (goals ranked below 7th silently disappearing from the daily view) is tracked as a new item under "複数ゴール運用時の優先度・スケール" above — the original "no cap" complaint is resolved, but Nozomi's follow-up on the cap's behavior is a distinct, still-open concern.
- **Multi-goal management gaps (backend) (#14) — edit/delete/priority** — resolved 2026-08-17. Verified in code: `src/usecase/goal/update_goal.go` (title/detail/condition/end_date/mode are all editable, wired to `GoalDetailPage.tsx`'s edit form), `src/usecase/goal/delete_goal.go` + `GoalRepository.Delete`'s cascading transaction (tasks + AI conversations + conversation messages, `goal_repository.go` lines 47-70), and `entity.Goal.Priority` + `src/usecase/goal/reorder_goal.go` + `GoalRepository.UpdatePriority`/`ORDER BY priority ASC, created_at ASC` (`goal_repository.go` lines 83-86, 107, 120) all present and connected end to end, including the frontend delete confirm + reorder buttons in `GoalsListPage.tsx`.
- **Multi-goal management gaps (backend) (#14) — no goal-list/focus view** — resolved 2026-08-17: `GoalsListPage.tsx` (list) and `GoalDetailPage.tsx` (single-goal focus, stats, edit, AI review) both exist and are routed in `App.tsx` (lines 130-140, 157). (The #14 sub-point about `WeekPathView`'s goal-title truncation is carried forward as still open — see above.)
- **Backend `DailyTask` has no completion timestamp** (the specific blocker #9 cited for why a trend feature couldn't be built) — resolved 2026-08-17: `entity.DailyTask.CompletedAt` is fully wired — persisted on insert/update (`daily_task_repository.go` `Save`/`UpdateDone`, lines 18-29, 48-58), cleared on un-done, and serialized via `json:"completed_at,omitempty"` through `DailyTaskHandler.List`. The overarching #9 feature request (trend view, export, and actually reading this timestamp in `get_goal_stats.go`) remains open — see "完了データはあるが活用されていない" above, which supersedes the old blocker note.
