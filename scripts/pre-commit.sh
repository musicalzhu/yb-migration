#!/bin/bash

# YB Migration Git Pre-commit Hook
# 在提交前运行质量检查，确保代码质量

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 YB Migration Pre-commit 质量检查${NC}"
echo "=================================="

# 检查是否有 Makefile
if [ ! -f "Makefile" ]; then
    echo -e "${RED}❌ 错误: 未找到 Makefile${NC}"
    exit 1
fi

# 检查是否有 golangci-lint
if ! command -v golangci-lint &> /dev/null; then
    echo -e "${YELLOW}⚠️  golangci-lint 未安装，正在安装...${NC}"
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
fi

# 检查是否有 goimports
if ! command -v goimports &> /dev/null; then
    echo -e "${YELLOW}⚠️  goimports 未安装，正在安装...${NC}"
    go install golang.org/x/tools/cmd/goimports@latest
fi

# 运行质量检查
echo -e "${BLUE}📝 1. 检查代码格式...${NC}"
if make format-check; then
    echo -e "${GREEN}✅ 格式检查通过${NC}"
else
    echo -e "${RED}❌ 格式检查失败${NC}"
    echo -e "${YELLOW}💡 运行 'make fix-format' 自动修复${NC}"
    exit 1
fi

echo -e "${BLUE}🔍 2. 运行轻量级 lint 检查...${NC}"
if make lint; then
    echo -e "${GREEN}✅ Lint 检查通过${NC}"
else
    echo -e "${RED}❌ Lint 检查失败${NC}"
    echo -e "${YELLOW}💡 运行 'make fix-lint' 尝试自动修复${NC}"
    exit 1
fi

echo -e "${BLUE}🧪 3. 运行单元测试...${NC}"
if make test; then
    echo -e "${GREEN}✅ 单元测试通过${NC}"
else
    echo -e "${RED}❌ 单元测试失败${NC}"
    exit 1
fi

# 检查是否有未提交的大文件
echo -e "${BLUE}📏 4. 检查文件大小...${NC}"
LARGE_FILES=$(git diff --cached --name-only | xargs -I {} find {} -type f -size +1M 2>/dev/null || true)
if [ -n "$LARGE_FILES" ]; then
    echo -e "${YELLOW}⚠️  警告: 发现大文件:${NC}"
    echo "$LARGE_FILES"
    echo -e "${YELLOW}💡 考虑使用 Git LFS 或减少文件大小${NC}"
fi

# 检查是否有敏感信息
echo -e "${BLUE}🔒 5. 检查敏感信息...${NC}"
SENSITIVE_FILES=$(git diff --cached --name-only --diff-filter=ACM | xargs -I {} grep -l -i "password\|secret\|key\|token" {} 2>/dev/null || true)
if [ -n "$SENSITIVE_FILES" ]; then
    echo -e "${RED}❌ 警告: 在暂存文件中发现可能的敏感信息:${NC}"
    echo "$SENSITIVE_FILES"
    echo -e "${RED}请检查并移除敏感信息后再提交${NC}"
    exit 1
else
    echo -e "${GREEN}✅ 未发现敏感信息${NC}"
fi

# 检查提交消息格式（如果是 git commit -m 的情况）
if [ -n "$1" ]; then
    echo -e "${BLUE}📝 6. 检查提交消息格式...${NC}"
    COMMIT_MSG="$1"
    
    # 检查提交消息长度
    if [ ${#COMMIT_MSG} -gt 72 ]; then
        echo -e "${YELLOW}⚠️  警告: 提交消息过长 (>72 字符)${NC}"
    fi
    
    # 检查是否以大写字母开头
    if [[ ! "$COMMIT_MSG" =~ ^[A-Z] ]]; then
        echo -e "${YELLOW}⚠️  建议: 提交消息以大写字母开头${NC}"
    fi
fi

echo ""
echo -e "${GREEN}🎉 所有 pre-commit 检查通过！${NC}"
echo -e "${GREEN}✅ 代码可以提交${NC}"
echo "=================================="
