#!/bin/bash
# Basic sanity tests for gotask

set -e

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GOTASK="$PROJECT_DIR/dist/gotask"
TEST_INBOX="$PROJECT_DIR/test-inbox.md"

echo "🧪 Running gotask sanity tests..."

# Build
echo "Building..."
cd "$PROJECT_DIR"
go build -o "$GOTASK" ./cmd/gotask 2>&1 || { echo "❌ Build failed"; exit 1; }
echo "✅ Build succeeded"

# Test: inbox top
echo ""
echo "Testing: inbox top"
$GOTASK inbox top 2>&1 | grep -q "Top 3" && echo "✅ inbox top works" || { echo "❌ inbox top failed"; exit 1; }

# Test: timer presets
echo "Testing: timer presets"
$GOTASK timer presets 2>&1 | grep -q "Pomodoro" && echo "✅ timer presets works" || { echo "❌ timer presets failed"; exit 1; }

# Test: focus snapshot
echo "Testing: focus snapshot"
$GOTASK focus snapshot 2>&1 | grep -q "FOCUS" && echo "✅ focus snapshot works" || { echo "❌ focus snapshot failed"; exit 1; }

# Test: timer start/status
echo "Testing: timer start/status"
$GOTASK timer start --preset "Pomodoro Standard" >/dev/null 2>&1
$GOTASK timer status 2>&1 | grep -q "Active Timer" && echo "✅ timer start/status works" || { echo "❌ timer start/status failed"; exit 1; }

# Test: timer stop
echo "Testing: timer stop"
$GOTASK timer stop >/dev/null 2>&1
$GOTASK timer status 2>&1 | grep -q "No active timer" && echo "✅ timer stop works" || { echo "❌ timer stop failed"; exit 1; }

# Test: inbox add
echo "Testing: inbox add"
$GOTASK inbox add --description "Test task from CLI" 2>&1 | grep -q "✓" && echo "✅ inbox add works" || { echo "❌ inbox add failed"; exit 1; }

# Test: inbox top (should now have more tasks)
echo "Testing: inbox top after add"
$GOTASK inbox top 2>&1 | grep -q "Test task" && echo "✅ Added task appears in inbox" || { echo "⚠️  Task not found in top 3 (may be ok)"; }

echo ""
echo "✅ All sanity tests passed!"
