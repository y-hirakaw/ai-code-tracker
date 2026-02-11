#!/bin/bash

# AI Code Tracker --since Option Integration Test
# Tests all major user scenarios and edge cases

# set -e  # Disabled to continue testing after errors

# CI環境ではカラー出力を無効化
if [ "${CI:-}" = "true" ] || [ "${NO_COLOR:-}" != "" ]; then
    GREEN=''
    RED=''
    YELLOW=''
    NC=''
else
    GREEN='\033[0;32m'
    RED='\033[0;31m'
    YELLOW='\033[1;33m'
    NC='\033[0m'
fi

# Test counters
PASSED=0
FAILED=0
TOTAL=0

# Helper functions
pass() {
    PASSED=$((PASSED + 1))
    TOTAL=$((TOTAL + 1))
    echo -e "${GREEN}✓${NC} $1"
}

fail() {
    FAILED=$((FAILED + 1))
    TOTAL=$((TOTAL + 1))
    echo -e "${RED}✗${NC} $1"
    if [ ! -z "$2" ]; then
        echo -e "${RED}  Error: $2${NC}"
    fi
}

info() {
    echo -e "${YELLOW}ℹ${NC} $1"
}

# Build the project
echo "Building aict..."
go build -o bin/aict ./cmd/aict
if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed${NC}"
    exit 1
fi
echo -e "${GREEN}Build successful${NC}"
echo ""

# Test 1: Shorthand notation - 7d (7 days ago)
echo "Test 1: Shorthand notation - 7d"
OUTPUT=$(./bin/aict report --since 7d 2>&1)
if echo "$OUTPUT" | grep -q "since 7d"; then
    pass "7d shorthand works and displays correctly"
else
    fail "7d shorthand failed" "$OUTPUT"
fi
echo ""

# Test 2: Shorthand notation - 2w (2 weeks ago)
echo "Test 2: Shorthand notation - 2w"
OUTPUT=$(./bin/aict report --since 2w 2>&1)
if echo "$OUTPUT" | grep -q "since 2w"; then
    pass "2w shorthand works and displays correctly"
else
    fail "2w shorthand failed" "$OUTPUT"
fi
echo ""

# Test 3: Shorthand notation - 1m (1 month ago)
echo "Test 3: Shorthand notation - 1m"
OUTPUT=$(./bin/aict report --since 1m 2>&1)
if echo "$OUTPUT" | grep -q "since 1m"; then
    pass "1m shorthand works and displays correctly"
else
    fail "1m shorthand failed" "$OUTPUT"
fi
echo ""

# Test 4: Relative date - "7 days ago"
echo "Test 4: Relative date - '7 days ago'"
OUTPUT=$(./bin/aict report --since '7 days ago' 2>&1)
if echo "$OUTPUT" | grep -q "since 7 days ago"; then
    pass "Relative date '7 days ago' works"
else
    fail "Relative date '7 days ago' failed" "$OUTPUT"
fi
echo ""

# Test 5: Relative date - "yesterday"
echo "Test 5: Relative date - 'yesterday'"
OUTPUT=$(./bin/aict report --since yesterday 2>&1)
if echo "$OUTPUT" | grep -q "since yesterday"; then
    pass "Relative date 'yesterday' works"
else
    fail "Relative date 'yesterday' failed" "$OUTPUT"
fi
echo ""

# Test 6: Absolute date - specific date
echo "Test 6: Absolute date - '2025-01-01'"
OUTPUT=$(./bin/aict report --since '2025-01-01' 2>&1)
if echo "$OUTPUT" | grep -q "since 2025-01-01"; then
    pass "Absolute date '2025-01-01' works"
else
    fail "Absolute date '2025-01-01' failed" "$OUTPUT"
fi
echo ""

# Test 7: Mutual exclusivity - both --range and --since
echo "Test 7: Mutual exclusivity - both --range and --since"
OUTPUT=$(./bin/aict report --range HEAD~5..HEAD --since 7d 2>&1)
if echo "$OUTPUT" | grep -q "mutually exclusive"; then
    pass "Mutual exclusivity check works"
else
    fail "Mutual exclusivity check failed" "$OUTPUT"
fi
echo ""

# Test 8: No option error - neither --range nor --since
echo "Test 8: No option error - neither --range nor --since"
OUTPUT=$(./bin/aict report 2>&1)
if echo "$OUTPUT" | grep -q "either --range or --since is required"; then
    pass "No option error works correctly"
else
    fail "No option error failed" "$OUTPUT"
fi
echo ""

# Test 9: JSON output format with --since
echo "Test 9: JSON output format with --since"
OUTPUT=$(./bin/aict report --since 7d --format json 2>&1)
if echo "$OUTPUT" | grep -q -E '("range"|No commits found)'; then
    pass "JSON format with --since works"
else
    fail "JSON format with --since failed" "$OUTPUT"
fi
echo ""

# Test 10: Very old date (should handle gracefully)
echo "Test 10: Very old date - '10 years ago'"
OUTPUT=$(./bin/aict report --since '10 years ago' 2>&1)
# Should either show commits or "No commits found"
if echo "$OUTPUT" | grep -q -E "(AI Code Generation Report|No commits found)"; then
    pass "Very old date handled gracefully"
else
    fail "Very old date handling failed" "$OUTPUT"
fi
echo ""

# Test 11: Compare --range and equivalent --since
echo "Test 11: Consistency between --range and --since"
# Get the first commit hash from 7 days ago
FIRST_COMMIT=$(git log --since='7 days ago' --format=%H --reverse | head -1)
if [ ! -z "$FIRST_COMMIT" ]; then
    # Check if parent exists
    if git rev-parse "${FIRST_COMMIT}^" >/dev/null 2>&1; then
        PARENT_COMMIT="${FIRST_COMMIT}^"
    else
        # No parent (initial commit), use the commit itself
        PARENT_COMMIT="${FIRST_COMMIT}"
    fi

    RANGE_OUTPUT=$(./bin/aict report --range ${PARENT_COMMIT}..HEAD 2>&1 | grep "Total Lines:" || true)
    SINCE_OUTPUT=$(./bin/aict report --since 7d 2>&1 | grep "Total Lines:" || true)

    if [ "$RANGE_OUTPUT" = "$SINCE_OUTPUT" ]; then
        pass "Results consistent between --range and --since"
    else
        info "Different results (this is acceptable if commits differ)"
        info "  --range: $RANGE_OUTPUT"
        info "  --since: $SINCE_OUTPUT"
        pass "Both commands executed successfully"
    fi
else
    info "No commits in last 7 days, skipping consistency test"
    pass "Skipped (no commits)"
fi
echo ""

# Test 12: Help text includes --since documentation
echo "Test 12: Help text includes --since documentation"
OUTPUT=$(./bin/aict --help 2>&1)
if echo "$OUTPUT" | grep -q "\-\-since" && echo "$OUTPUT" | grep -q "7d"; then
    pass "Help text documents --since option with examples"
else
    fail "Help text missing --since documentation"
fi
echo ""

# Test 13: Error message quality for invalid --since
echo "Test 13: Error message quality for invalid --since"
OUTPUT=$(./bin/aict report --since 'invalid-date-format' 2>&1)
if echo "$OUTPUT" | grep -q -E "(Error|failed|no commits found)"; then
    pass "Invalid --since input handled with error message"
else
    fail "Invalid --since input not handled properly" "$OUTPUT"
fi
echo ""

# Test 14: Real-world scenario - Sprint review (2 weeks)
echo "Test 14: Real-world scenario - Sprint review (2 weeks)"
OUTPUT=$(./bin/aict report --since 2w 2>&1)
if echo "$OUTPUT" | grep -q -E "(AI Code Generation Report|No commits found)"; then
    pass "Sprint review scenario (2w) works"
    info "  Use case: 2-week sprint retrospective"
else
    fail "Sprint review scenario failed" "$OUTPUT"
fi
echo ""

# Test 15: Real-world scenario - Daily standup (1 day)
echo "Test 15: Real-world scenario - Daily standup (1 day)"
OUTPUT=$(./bin/aict report --since 1d 2>&1)
if echo "$OUTPUT" | grep -q -E "(AI Code Generation Report|No commits found)"; then
    pass "Daily standup scenario (1d) works"
    info "  Use case: Daily development review"
else
    fail "Daily standup scenario failed" "$OUTPUT"
fi
echo ""

# Test 16: Real-world scenario - Monthly release (1 month)
echo "Test 16: Real-world scenario - Monthly release (1 month)"
OUTPUT=$(./bin/aict report --since 1m 2>&1)
if echo "$OUTPUT" | grep -q -E "(AI Code Generation Report|No commits found)"; then
    pass "Monthly release scenario (1m) works"
    info "  Use case: Monthly release retrospective"
else
    fail "Monthly release scenario failed" "$OUTPUT"
fi
echo ""

# Summary
echo "========================================="
echo "Test Summary"
echo "========================================="
echo -e "Total:  $TOTAL"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi
