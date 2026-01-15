#!/bin/bash

# pre-check.sh - 代码提交前检查脚本
# 用法: ./pre-check.sh

set -o pipefail

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 结果记录
FMT_RESULT=0
LINT_RESULT=0
TEST_RESULT=0

# 项目目录
PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"
SERVER_DIR="${PROJECT_ROOT}/server"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}       代码提交前检查 (Pre-Check)       ${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 检查 server 目录是否存在
if [ ! -d "$SERVER_DIR" ]; then
    echo -e "${RED}[错误] server 目录不存在: ${SERVER_DIR}${NC}"
    exit 1
fi

cd "$SERVER_DIR" || exit 1
echo -e "${YELLOW}工作目录: ${SERVER_DIR}${NC}"
echo ""

# ========== 1. go fmt ==========
echo -e "${BLUE}[1/3] 正在执行 go fmt ./...${NC}"
echo "----------------------------------------"

FMT_OUTPUT=$(go fmt ./... 2>&1)
FMT_RESULT=$?

if [ -n "$FMT_OUTPUT" ]; then
    echo -e "${YELLOW}以下文件已格式化:${NC}"
    echo "$FMT_OUTPUT"
    FMT_RESULT=1
else
    echo -e "${GREEN}所有文件格式正确${NC}"
fi

echo ""

# ========== 2. golangci-lint ==========
echo -e "${BLUE}[2/3] 正在执行 golangci-lint run ./...${NC}"
echo "----------------------------------------"

if ! command -v golangci-lint &> /dev/null; then
    echo -e "${RED}[错误] golangci-lint 未安装${NC}"
    echo "请安装: go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8"
    LINT_RESULT=1
else
    golangci-lint run ./...
    LINT_RESULT=$?

    if [ $LINT_RESULT -eq 0 ]; then
        echo -e "${GREEN}lint 检查通过${NC}"
    else
        echo -e "${RED}lint 检查发现问题${NC}"
    fi
fi

echo ""

# ========== 3. go test ==========
echo -e "${BLUE}[3/3] 正在执行 go test -gcflags=all=-l -v ./...${NC}"
echo "----------------------------------------"

go test -gcflags=all=-l -v ./...
TEST_RESULT=$?

if [ $TEST_RESULT -eq 0 ]; then
    echo -e "${GREEN}测试通过${NC}"
else
    echo -e "${RED}测试失败${NC}"
fi

echo ""

# ========== 汇总结果 ==========
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}                检查结果                ${NC}"
echo -e "${BLUE}========================================${NC}"

if [ $FMT_RESULT -eq 0 ]; then
    echo -e "  go fmt:        ${GREEN}✓ 通过${NC}"
else
    echo -e "  go fmt:        ${YELLOW}⚠ 有文件被格式化${NC}"
fi

if [ $LINT_RESULT -eq 0 ]; then
    echo -e "  golangci-lint: ${GREEN}✓ 通过${NC}"
else
    echo -e "  golangci-lint: ${RED}✗ 失败${NC}"
fi

if [ $TEST_RESULT -eq 0 ]; then
    echo -e "  go test:       ${GREEN}✓ 通过${NC}"
else
    echo -e "  go test:       ${RED}✗ 失败${NC}"
fi

echo ""

# 最终结果
TOTAL_RESULT=$((FMT_RESULT + LINT_RESULT + TEST_RESULT))
if [ $TOTAL_RESULT -eq 0 ]; then
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  所有检查通过，可以提交代码！${NC}"
    echo -e "${GREEN}========================================${NC}"
    exit 0
else
    echo -e "${RED}========================================${NC}"
    echo -e "${RED}  存在问题，请修复后再提交！${NC}"
    echo -e "${RED}========================================${NC}"
    exit 1
fi
