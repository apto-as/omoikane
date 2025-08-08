#!/bin/bash

# Week 1 Implementation Test Suite
# Trinity統合テストスクリプト

set -e

echo "🔮 Omoikane Week 1 Test Suite Starting..."
echo "================================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results
PASSED=0
FAILED=0

# Function to run tests
run_test() {
    local test_name=$1
    local test_cmd=$2
    
    echo -e "\n${BLUE}▶ Running: ${test_name}${NC}"
    echo "----------------------------------------"
    
    if eval $test_cmd; then
        echo -e "${GREEN}✓ ${test_name} passed${NC}"
        ((PASSED++))
    else
        echo -e "${RED}✗ ${test_name} failed${NC}"
        ((FAILED++))
    fi
}

# 1. Unit Tests
echo -e "\n${YELLOW}=== Phase 1: Unit Tests ===${NC}"

run_test "Subagent Manager Tests" \
    "go test ./internal/subagent -v -timeout 30s"

run_test "Security Auditor Tests" \
    "go test ./internal/security -v -timeout 30s"

# 2. Benchmark Tests
echo -e "\n${YELLOW}=== Phase 2: Performance Benchmarks ===${NC}"

run_test "Parallel Execution Benchmark" \
    "go test ./internal/benchmark -bench=BenchmarkParallelExecution -benchtime=10x -timeout 60s"

run_test "Memory Efficiency Benchmark" \
    "go test ./internal/benchmark -bench=BenchmarkMemoryEfficiency -benchtime=10x -timeout 60s"

run_test "Latency Benchmark" \
    "go test ./internal/benchmark -bench=BenchmarkLatency -benchtime=10x -timeout 60s"

# 3. Trinity Integration Tests
echo -e "\n${YELLOW}=== Phase 3: Trinity Integration ===${NC}"

run_test "Trinity Workflow Test" \
    "go test ./internal/subagent -v -run TestManagerTrinityIntegration -timeout 30s"

run_test "Trinity Benchmark" \
    "go test ./internal/benchmark -bench=BenchmarkTrinityWorkflow -benchtime=5x -timeout 60s"

# 4. Coverage Report
echo -e "\n${YELLOW}=== Phase 4: Coverage Analysis ===${NC}"

echo "Generating coverage report..."
go test ./... -coverprofile=coverage.out 2>/dev/null
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
echo -e "Total Coverage: ${GREEN}${COVERAGE}${NC}"

# 5. Race Condition Detection
echo -e "\n${YELLOW}=== Phase 5: Race Condition Check ===${NC}"

run_test "Race Detector" \
    "go test ./internal/subagent -race -timeout 60s"

# Summary
echo -e "\n${YELLOW}================================================${NC}"
echo -e "${BLUE}🎯 Test Summary${NC}"
echo -e "================================================"
echo -e "${GREEN}Passed: ${PASSED}${NC}"
echo -e "${RED}Failed: ${FAILED}${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "\n${GREEN}✨ All tests passed successfully!${NC}"
    echo -e "${GREEN}Week 1 implementation is working correctly.${NC}"
    exit 0
else
    echo -e "\n${RED}⚠️  Some tests failed. Please review the output above.${NC}"
    exit 1
fi