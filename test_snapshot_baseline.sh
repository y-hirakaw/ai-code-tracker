#!/bin/bash
# Test script for snapshot-based baseline functionality
# Tests that existing code is not counted when aict is first initialized

set -e

TEST_DIR="/tmp/aict_snapshot_test_$$"
AICT_BIN="$(pwd)/bin/aict"

echo "=== Testing Snapshot-Based Baseline Functionality ==="
echo ""

# Setup test repository
setup_test_repo() {
    echo "Setting up test repository..."
    rm -rf "$TEST_DIR"
    mkdir -p "$TEST_DIR"
    cd "$TEST_DIR"

    git init
    git config user.name "Test User"
    git config user.email "test@example.com"

    # Create existing code (100 lines)
    cat > main.go << 'EOF'
package main

import "fmt"

func main() {
    fmt.Println("Line 1")
    fmt.Println("Line 2")
    fmt.Println("Line 3")
    fmt.Println("Line 4")
    fmt.Println("Line 5")
    fmt.Println("Line 6")
    fmt.Println("Line 7")
    fmt.Println("Line 8")
    fmt.Println("Line 9")
    fmt.Println("Line 10")
    fmt.Println("Line 11")
    fmt.Println("Line 12")
    fmt.Println("Line 13")
    fmt.Println("Line 14")
    fmt.Println("Line 15")
    fmt.Println("Line 16")
    fmt.Println("Line 17")
    fmt.Println("Line 18")
    fmt.Println("Line 19")
    fmt.Println("Line 20")
    fmt.Println("Line 21")
    fmt.Println("Line 22")
    fmt.Println("Line 23")
    fmt.Println("Line 24")
    fmt.Println("Line 25")
}
EOF

    git add main.go
    git commit -m "Initial commit with existing code"

    echo "✓ Test repository created with 25 lines of existing code"
}

# Test 1: Initialize aict and create baseline checkpoint
test_initial_checkpoint() {
    echo ""
    echo "Test 1: Initial checkpoint should not count existing code"
    echo "-------------------------------------------------------"

    # Initialize aict
    "$AICT_BIN" init

    # Create initial checkpoint (should create baseline snapshot)
    "$AICT_BIN" checkpoint --author "Human"

    # Check that no changes were recorded
    if [ -f .git/aict/checkpoints/*.json ]; then
        CHECKPOINT_FILE=$(ls .git/aict/checkpoints/*.json | head -1)
        CHANGES_COUNT=$(grep -o '"changes":{' "$CHECKPOINT_FILE" | wc -l)

        if [ "$CHANGES_COUNT" -eq 1 ]; then
            CHANGES_CONTENT=$(grep -A 1 '"changes":' "$CHECKPOINT_FILE")
            if echo "$CHANGES_CONTENT" | grep -q '{}'; then
                echo "✓ Initial checkpoint has empty changes (correct)"
            else
                echo "✗ Initial checkpoint has non-empty changes (incorrect)"
                cat "$CHECKPOINT_FILE"
                return 1
            fi
        fi
    fi

    # Report should show 0 changes for initial checkpoint
    echo "Report output:"
    "$AICT_BIN" report --range HEAD~1..HEAD
    echo ""
}

# Test 2: Add 5 lines with AI, should only count those 5 lines
test_ai_addition() {
    echo ""
    echo "Test 2: AI adds 5 lines, should count exactly 5 lines"
    echo "------------------------------------------------------"

    # Add 5 new lines
    cat >> main.go << 'EOF'
    fmt.Println("AI Line 1")
    fmt.Println("AI Line 2")
    fmt.Println("AI Line 3")
    fmt.Println("AI Line 4")
    fmt.Println("AI Line 5")
EOF

    # Create AI checkpoint
    "$AICT_BIN" checkpoint --author "Claude"

    # Commit
    git add main.go
    git commit -m "AI added 5 lines"
    "$AICT_BIN" commit

    # Check report
    echo "Report output:"
    REPORT=$("$AICT_BIN" report --range HEAD~1..HEAD)
    echo "$REPORT"

    # Verify AI lines count
    AI_LINES=$(echo "$REPORT" | grep "🤖 AI Generated:" | awk '{print $4}')
    if [ "$AI_LINES" = "5" ]; then
        echo "✓ AI lines correctly counted as 5"
    else
        echo "✗ AI lines counted as $AI_LINES (expected 5)"
        return 1
    fi

    # Verify human lines count
    HUMAN_LINES=$(echo "$REPORT" | grep "👤 Human Written:" | awk '{print $4}' || echo "0")
    if [ "$HUMAN_LINES" = "0" ] || [ -z "$HUMAN_LINES" ]; then
        echo "✓ Human lines correctly counted as 0"
    else
        echo "✗ Human lines counted as $HUMAN_LINES (expected 0)"
        return 1
    fi
    echo ""
}

# Test 3: Human adds 3 lines
test_human_addition() {
    echo ""
    echo "Test 3: Human adds 3 lines, should count exactly 3 lines"
    echo "---------------------------------------------------------"

    # Add 3 new lines
    cat >> main.go << 'EOF'
    fmt.Println("Human Line 1")
    fmt.Println("Human Line 2")
    fmt.Println("Human Line 3")
EOF

    # Create human checkpoint
    "$AICT_BIN" checkpoint --author "Human"

    # Commit
    git add main.go
    git commit -m "Human added 3 lines"
    "$AICT_BIN" commit

    # Check report for this commit
    echo "Report for last commit:"
    REPORT=$("$AICT_BIN" report --range HEAD~1..HEAD)
    echo "$REPORT"

    # Check cumulative report
    echo ""
    echo "Cumulative report:"
    CUMULATIVE=$("$AICT_BIN" report --range HEAD~2..HEAD)
    echo "$CUMULATIVE"

    # Verify totals
    TOTAL_AI=$(echo "$CUMULATIVE" | grep "🤖 AI Generated:" | awk '{print $4}')
    TOTAL_HUMAN=$(echo "$CUMULATIVE" | grep "👤 Human Written:" | awk '{print $4}')

    if [ "$TOTAL_AI" = "5" ] && [ "$TOTAL_HUMAN" = "3" ]; then
        echo "✓ Cumulative counts correct: AI=5, Human=3"
    else
        echo "✗ Cumulative counts incorrect: AI=$TOTAL_AI, Human=$TOTAL_HUMAN (expected AI=5, Human=3)"
        return 1
    fi
    echo ""
}

# Cleanup
cleanup() {
    cd /
    rm -rf "$TEST_DIR"
    echo "Cleaned up test directory"
}

# Run tests
trap cleanup EXIT

setup_test_repo

if test_initial_checkpoint && test_ai_addition && test_human_addition; then
    echo ""
    echo "=========================================="
    echo "✓ All snapshot baseline tests passed!"
    echo "=========================================="
    exit 0
else
    echo ""
    echo "=========================================="
    echo "✗ Some tests failed"
    echo "=========================================="
    exit 1
fi
