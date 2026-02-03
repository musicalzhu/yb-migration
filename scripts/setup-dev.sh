#!/bin/bash

# YB Migration 开发环境安装脚本
# 设置 Git hooks、安装工具、初始化开发环境

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 YB Migration 开发环境安装${NC}"
echo "=================================="

# 检查是否在项目根目录
if [ ! -f "go.mod" ] || [ ! -f "Makefile" ]; then
    echo -e "${RED}❌ 错误: 请在项目根目录运行此脚本${NC}"
    exit 1
fi

# 创建 scripts 目录
mkdir -p scripts

# 安装 Git hooks
echo -e "${BLUE}🔧 安装 Git hooks...${NC}"
if [ -f "scripts/pre-commit.sh" ]; then
    chmod +x scripts/pre-commit.sh
    if [ -d ".git" ]; then
        ln -sf "../../scripts/pre-commit.sh" .git/hooks/pre-commit
        echo -e "${GREEN}✅ Pre-commit hook 安装成功${NC}"
    else
        echo -e "${YELLOW}⚠️  不是 Git 仓库，跳过 hook 安装${NC}"
    fi
else
    echo -e "${RED}❌ Pre-commit 脚本不存在${NC}"
fi

# 安装开发工具
echo -e "${BLUE}📦 安装开发工具...${NC}"
make install-tools

# 验证工具安装
echo -e "${BLUE}🔍 验证工具安装...${NC}"
make check-tools

# 下载 Go 依赖
echo -e "${BLUE}📥 下载 Go 依赖...${NC}"
go mod download
go mod verify

# 运行初始质量检查
echo -e "${BLUE}🔍 运行初始质量检查...${NC}"
if make format-check && make lint; then
    echo -e "${GREEN}✅ 初始质量检查通过${NC}"
else
    echo -e "${YELLOW}⚠️  质量检查发现问题，建议修复后再开始开发${NC}"
    echo -e "${YELLOW}💡 运行 'make fix-format' 和 'make fix-lint' 尝试修复${NC}"
fi

# 创建本地配置文件（如果不存在）
if [ ! -f ".env.local" ]; then
    echo -e "${BLUE}📝 创建本地配置文件...${NC}"
    cat > .env.local << EOF
# YB Migration 本地配置
# 这个文件不会被 Git 跟踪，可以包含本地开发配置

# 开发模式
GO_ENV=development

# 调试模式
DEBUG=true

# 测试数据库（如果需要）
# TEST_DB_HOST=localhost
# TEST_DB_PORT=3306
# TEST_DB_USER=root
# TEST_DB_PASSWORD=
EOF
    echo -e "${GREEN}✅ 创建 .env.local 文件${NC}"
fi

# 创建 IDE 配置
echo -e "${BLUE}⚙️  创建 IDE 配置...${NC}"

# VS Code 配置
mkdir -p .vscode
if [ ! -f ".vscode/settings.json" ]; then
    cat > .vscode/settings.json << EOF
{
    "go.lintTool": "golangci-lint",
    "go.lintFlags": [
        "--fast"
    ],
    "go.formatTool": "goimports",
    "go.useLanguageServer": true,
    "go.testFlags": ["-v", "-race"],
    "go.coverOnSave": true,
    "go.coverageDecorator": {
        "type": "gutter",
        "coveredHighlightColor": "rgba(64,128,64,0.5)",
        "uncoveredHighlightColor": "rgba(128,64,64,0.25)"
    },
    "files.exclude": {
        "**/bin": true,
        "**/coverage.txt": true,
        "**/coverage.html": true,
        "**/quality-report.html": true,
        "**/quality-report.json": true
    },
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
        "source.organizeImports": true
    }
}
EOF
    echo -e "${GREEN}✅ 创建 VS Code 配置${NC}"
fi

# Git 配置
echo -e "${BLUE}🔧 配置 Git...${NC}"
git config core.autocrlf false
git config core.eol lf
git config core.safecrlf warn

# 显示项目状态
echo -e "${BLUE}📊 项目状态:${NC}"
make status

echo ""
echo -e "${GREEN}🎉 开发环境安装完成！${NC}"
echo ""
echo -e "${BLUE}🚀 快速开始:${NC}"
echo "  make help           # 查看所有可用命令"
echo "  make dev-setup      # 重新初始化开发环境"
echo "  make quality-check  # 运行质量检查"
echo "  make test           # 运行测试"
echo "  make build          # 构建应用"
echo ""
echo -e "${BLUE}📚 更多信息:${NC}"
echo "  - 查看 CI-CD-Quality-Guide.md 了解质量门禁"
echo "  - 查看 Makefile 了解所有可用命令"
echo "  - 查看 .golangci.yml 了解代码质量规则"
echo ""
echo -e "${GREEN}Happy Coding! 🎯${NC}"
