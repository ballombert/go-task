# Phase 2-4 Synced Roadmap (Current State)

Last sync: 2026-05-12

This file is aligned with the current codebase and replaces the old assumption that TUI key handling was still in stub state.

## Phase 1 Closure Gate

Validation checklist:
- [x] `task build` passes
- [x] `task test` passes
- [x] CLI foundation complete (inbox/timer/focus)
- [x] Persistent timer state in place
- [x] TUI core navigation and modals delivered
- [x] Timer overlay and live refresh delivered

Decision: `GO` -> Phase 1 is closed. Work proceeds under Phase 2-4 scope only.

## Current Delivery Snapshot

### Completed
- TUI keyboard navigation for Tasks, Focus, Logs.
- Global view switching (t/f/l and left/right).
- Move mode for task reordering.
- Timer start shortcuts (AZERTY-friendly aliases included).
- Focus timer full-screen overlay with large timer and progress bar.
- 1-second UI tick refresh for timer display.
- Task modal input using Bubble Tea text input.
- Advanced field editor in modal (description, priority, due date, duration).
- Focus shortlist sorting/filtering by priority and due date (completed tasks excluded).
- Logs panel migrated to Bubble Tea viewport.

### Partially completed
- TUI hardening is in progress (initial regression suite added, more flow coverage pending).

### Not started
- Recurrence engine implementation.
- Real OS notifications.
- Background evaluator loop.
- Analytics package and export.
- Config package and config file loading.

## Phase 2: TUI Completion (1-2 weeks)

### Priority 1: Advanced Task Editor

Status: done

Target file:
- internal/tui/app.go
- internal/tui/views.go

Scope:
- Edit description, priority, due date, duration.
- Tab / Shift+Tab field navigation.
- Enter save, Esc cancel.
- Reuse current centered modal behavior.

### Priority 2: Tasks Sorting and Filtering

Status: done

Current TODO:
- internal/tui/views.go (getTopNTasks)

Scope:
- Sort by effective priority and due date.
- Exclude completed tasks from focus shortlist by default.
- Keep predictable ordering for subtasks.

### Priority 3: TUI Hardening

Status: in progress

Scope:
- Add tests for keybinding flows in Tasks/Focus/Logs.
- Stabilize modal/timer overlay precedence rules.
- Reduce duplicate state fields where possible.

Delivered in this pass:
- Added TUI regression tests for modal field navigation/save/validation.
- Added sorting/filtering test for focus shortlist behavior.
- Added move mode cancel test and timer-overlay-precedence test.

## Phase 3: Recurrence + Notifications + Daemon (2-3 weeks)

### Priority 1: Recurrence Engine

Status: not started

Target file:
- internal/rules/recurrence.go (new)

Minimal v1 scope:
- Daily and weekly recurrence first.
- Generate instances up to a horizon date.
- Append generated tasks to inbox safely.

### Priority 2: System Notifications

Status: scaffold exists, real impl missing

Current TODOs:
- internal/notifications/notifier.go

Scope:
- Windows: go-toast.
- macOS: osascript.
- Linux: notify-send or dbus.
- Add cooldown and dedupe guard.

### Priority 3: Background Evaluator Loop

Status: not started

Target file:
- internal/daemon/evaluator.go (new)

Scope:
- Evaluate rules every 30s.
- Send notifications and log interventions.
- Clean start/stop lifecycle.

## Phase 4: Analytics + Config (1-2 weeks)

### Priority 1: Analytics

Status: not started

Target file:
- internal/analytics/tracker.go (new)

Scope:
- Daily stats, session stats, intervention counts.
- Optional JSON export.

### Priority 2: Config Support

Status: not started

Target file:
- internal/config/config.go (new)

Scope:
- Load ~/.gotask/config.yaml.
- Support inbox path, notification toggle, timer presets.

## Testing Strategy (Updated)

### TUI tests
```bash
go test ./internal/tui/... -v
```

Add targeted tests for:
- create/edit modal lifecycle,
- move mode confirm/cancel,
- timer overlay behavior,
- logs viewport scrolling.

### Rules and recurrence tests
```bash
go test ./internal/rules/... -v
```

### End-to-end sanity
```bash
./dist/gotask inbox top
./dist/gotask timer status
./dist/gotask focus snapshot
```

## Known Gaps and Risks

- Notifications currently print to stdout (platform backends not implemented).
- Recurrence is not yet integrated.
- TUI hardening suite is still partial (additional Focus/Logs flows to cover).
- No dedicated daemon process yet.

## Success Criteria (Updated)

### Phase 2
- Full keyboard-only TUI with advanced field editor.
- Stable task ordering/sorting in focus and task lists.
- Regression tests for keybindings and modal behavior.

### Phase 3
- Recurring tasks generated reliably.
- Notifications working on Windows/macOS/Linux.
- Background evaluation loop active every 30s.

### Phase 4
- Analytics available via CLI output and optional export.
- Config file loading and override behavior documented.

---

Estimated remaining timeline: 4-6 weeks focused execution.

Reference behavior notes:
- docs/cli.md
