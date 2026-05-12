# gotask - Task Management for ADHD Productivity

A cross-platform Go application designed to help people with ADHD manage daily tasks, timers, and focus. Based on the original .NET "ConducteurOrchestre" project, gotask provides a lightweight, command-line-first interface with an optional TUI (Terminal User Interface).

## 🎯 Project Status

**Phase 1: ✅ Closed (Validated)**

- ✅ Go project structure with proper package layout
- ✅ Domain models (Task, Timer, Recurrence, Rules)
- ✅ Markdown inbox parser/writer (Obsidian Tasks compatible)
- ✅ SQLite logging for sessions and interventions
- ✅ CLI commands for inbox, timer, and focus management
- ✅ Timer state persistence between CLI invocations
- ✅ Interactive TUI with keyboard navigation and overlays

**Phase 2-4: In Progress**
- Advanced task editor (multi-field)
- Task sorting/filtering refinements
- Recurrence engine for automatic task generation
- System notifications (platform-specific)

## ✅ Phase 1 Acceptance Checklist

- [x] Build passes with `task build`
- [x] Test command passes with `task test`
- [x] Core CLI commands usable (`inbox`, `timer`, `focus`)
- [x] Timer state persists between CLI invocations
- [x] Markdown inbox read/write is functional
- [x] SQLite logging is functional
- [x] TUI keyboard navigation works for core flows
- [x] TUI task create/edit modal works (Enter save, Esc cancel)
- [x] Timer overlay with live progress is visible in TUI

Phase 1 closure decision: `GO` (closed and validated)

## 🚀 Quick Start

### Prerequisites

Install [Task](https://taskfile.dev):
```bash
brew install go-task
# or
go install github.com/go-task/task/v3/cmd/task@latest
```

### Build

```bash
task build           # Build for current OS
task build:windows   # Windows
task build:macos     # macOS
task build:linux     # Linux
task build:all       # Build all platforms
```

### Basic Usage

```bash
# Launch interactive TUI
./dist/gotask
# or
./dist/gotask tui

# Show top 3 tasks
./dist/gotask inbox top

# Add a new task
./dist/gotask inbox add --description "My new task"

# Start a timer
./dist/gotask timer start --preset "Pomodoro Standard" --line 1

# Check timer status
./dist/gotask timer status

# Stop the timer
./dist/gotask timer stop

# Show focus snapshot (active timer + top 3 tasks)
./dist/gotask focus snapshot
```

## 📁 Project Structure

```
cmd/gotask/              CLI entry point
internal/
  domain/                Core models (Task, Timer, Recurrence, Rule)
  storage/               Markdown inbox + SQLite logger
  cli/                   Command handlers
  tui/                   Terminal UI (Bubbletea)
  timer/                 Timer state management
  rules/                 Rule evaluation engine
  notifications/         Platform-specific notifications
```

## 📋 Inbox Format

Tasks are stored in `inbox.md` in Obsidian Tasks format:

```markdown
- [ ] Task description 📅 2026-05-15 ⏫ ⏱ 30m
  - [ ] Subtask
- [>] Task in progress 🔺 ⏱ 120m
- [x] Completed task ✅ 2026-05-12
```

**Status symbols:**
- `[ ]` = Paused
- `[>]` = In progress (only one at a time)
- `[x]` = Completed

**Priority emojis:**
- `🔺` = Highest
- `⏫` = High
- `🔼` = Medium
- `🔽` = Low
- `⏬` = Lowest

**Metadata:**
- `📅 YYYY-MM-DD` = Due date
- `⏱ Xm` or `⏱ Xh Ym` = Time spent
- `✅ YYYY-MM-DD` = Completion date

## ⚙️ Configuration

- Inbox location: `inbox.md` (current directory)
- Data storage: `~/.gotask/` (SQLite database, timer state)
- Logs: `~/.gotask/gotask.db`

## 🔧 Development

### Dependencies

- Go 1.21+
- Bubbletea (TUI framework)
- Lipgloss (terminal styling)
- go-sqlite3 (database)

### Install dependencies

```bash
task deps
```

### Run tests

```bash
task test              # Run all tests
task test:coverage     # Run with coverage report
```

### Code quality

```bash
task fmt               # Format code
task lint              # Run linter
task check             # Run all checks (fmt + lint + test)
```

### Clean build artifacts

```bash
task clean             # Remove build artifacts
task clean:all         # Deep clean (remove config too)
```

## 🎨 Design Principles

Per the project spec, gotask follows these core constraints:

1. **One intervention at a time** - Never show multiple alerts
2. **Max 3 infos displayed** - Keep visual simple (e.g., top 3 tasks)
3. **Max 2 choices offered** - Reduce decision fatigue
4. **Prefer simplicity** - Over elegance or features
5. **Visible persistence** - Store in readable format (Markdown + SQLite)
6. **Deterministic rules** - No opaque AI, explicit logic only
7. **Reduce cognitive load** - Each feature must decrease net load

## 📚 Key Files

- [objectif.md](docs/objectif.md) - Product spec and requirements
- [cli.md](docs/cli.md) - CLI interface specification
- [Roadmap.md](docs/Roadmap.md) - Development roadmap
- [inbox.md](inbox.md) - Example task file

## 🔄 Workflow

### For Task Management

1. User opens inbox (CLI or TUI)
2. System shows top 3 prioritized tasks
3. User selects a task and starts a timer (Pomodoro preset)
4. Timer runs; system tracks elapsed time
5. On completion, system suggests next action
6. User marks task done, system logs session to SQLite

### For Recurring Tasks

1. User defines recurrence pattern (daily, weekly, custom)
2. Background engine generates instances automatically
3. Instances appear in inbox with proper due dates
4. User completes instances, history is logged

## 🐛 Known Limitations (Phase 1)

- Recurrence engine not active (roadmap for Phase 2)
- Platform-specific notifications are stubs
- No background daemon for automatic task generation
- Advanced field editor in TUI not complete (description modal is implemented)

## 📝 Next Steps (Phase 2-3)

- [ ] Advanced TUI editor (priority, due date, duration, recurrence)
- [ ] Recurrence engine and automatic task generation
- [ ] Real system notifications (go-toast, osascript, d-bus)
- [ ] Background task evaluation loop
- [ ] Intervention/alert system
- [ ] Time tracking analytics

## 📄 License

Based on ConducteurOrchestre concept. See docs/ for original specifications.

## 🤝 Contributing

Contributions welcome! Please ensure:
- Go code follows standard fmt/lint checks
- New features respect product design principles
- Changes update relevant documentation
- Tests added for new functionality

## 📞 Questions?

Refer to [docs/objectif.md](docs/objectif.md) for product vision and [docs/cli.md](docs/cli.md) for interface specifications.
