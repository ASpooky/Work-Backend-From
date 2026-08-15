---
name: persona-review
description: Run this app's fixed 10-persona panel against the current frontend/backend for UX, design, and feature feedback, then synthesize it into docs/feedback-backlog.md. Use when the user asks for a persona review, wants feedback on the current UI, or wants the feedback backlog updated.
---

# Persona Review

This project doesn't have a connected browser-automation tool in every environment, so personas review the **source of truth for what renders**: the React components and CSS under `frontend/src/`, plus the API surface under `src/handler/httpapi/` and `src/usecase/` for feature-level behavior. If `claude-in-chrome` (or another browser tool) is available and the dev server is running, personas may additionally drive the live app for a closer-to-real read — prefer that when it's available, but code-based review is the baseline and is always valid.

## Roster

Read `personas.md` in this same directory. It is a **fixed roster** — reuse the same 10 people every run so feedback is comparable over time. Only edit an entry if it has stopped producing distinct feedback; don't regenerate the whole list per run.

## Process

1. Read `personas.md` and the existing `docs/feedback-backlog.md` (if present) for context on what's already been raised.
2. Spawn one subagent per persona, **in parallel, in a single message, in the foreground** (the next step depends on all of them). Each agent's prompt must include:
   - The persona's full description verbatim.
   - Instruction to review the current `frontend/src/` (and relevant `src/` backend pieces for feature-level gaps) and give feedback strictly from that persona's stated priorities — not a generic checklist.
   - Instruction to stay concrete: cite actual files, component names, copy strings, or flows. No generic advice ("improve accessibility") without pointing at what specifically fails and where.
   - Instruction to keep output short: 3-6 bullet points max, each tagged `[UX]`, `[Design]`, or `[Feature]`.
   - It's fine and expected for a persona to have little or nothing to say outside their concern — don't pad.
3. Once all 10 return, spawn **one more subagent** (general-purpose) as the synthesizer. Give it all 10 raw outputs plus the current `docs/feedback-backlog.md` content, and instruct it to:
   - De-duplicate overlapping points raised by multiple personas (note when 2+ personas independently raised the same issue — that's a strong signal, keep the count).
   - Group into themes, not persona-by-persona transcript.
   - Preserve any existing backlog items still unresolved; mark newly-fixed ones as resolved only if the synthesizer can verify the fix actually landed in the current code (don't take it on faith).
   - Write the updated result to `docs/feedback-backlog.md` using the structure below.
4. Report back to the user: a short summary (not the full backlog dump) of what's new this run, plus the file path.

## docs/feedback-backlog.md structure

```markdown
# Feedback Backlog

Last synthesized: <date>, from personas: <list>

## Open

### <theme>
- <item> — raised by: <persona names>
  ...

## Resolved
- <item> — resolved <date>, verified in <file/commit>
```

Keep entries terse — one line each where possible. This file accumulates across runs; never blow it away, only merge into it.
