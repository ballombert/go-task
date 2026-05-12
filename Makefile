# ⚠️  DEPRECATED: This Makefile has been replaced by Taskfile.yml
#
# Install Task: brew install go-task
# Or: go install github.com/go-task/task/v3/cmd/task@latest
#
# Usage: task [TASK_NAME]
#
# Common commands:
#   task build       - Build for current OS
#   task build:all   - Build all platforms
#   task test        - Run tests
#   task clean       - Clean artifacts
#   task --list-all  - Show all tasks
#
# For backward compatibility, you can still use 'make' if you prefer:

default:
	@echo "⚠️  NOTICE: This project now uses Taskfile.yml instead of Makefile"
	@echo ""
	@echo "Install Task: brew install go-task"
	@echo "Then use: task [TASK_NAME]"
	@echo ""
	@echo "Examples:"
	@echo "  task build       - Build for current OS"
	@echo "  task test        - Run tests"
	@echo "  task clean       - Clean artifacts"
	@echo "  task --list-all  - List all tasks"
	@echo ""
	@echo "For more info, see README.md or Taskfile.yml"

# Fallback targets for backward compatibility
.PHONY: build test clean help

build:
	@command -v task >/dev/null 2>&1 && task build || make_error "Task not installed"

test:
	@command -v task >/dev/null 2>&1 && task test || make_error "Task not installed"

clean:
	@command -v task >/dev/null 2>&1 && task clean || make_error "Task not installed"

help:
	@command -v task >/dev/null 2>&1 && task help || make_error "Task not installed"

make_error:
	@echo "❌ Error: Task is not installed"
	@echo "Install with: brew install go-task"
	@exit 1

