# YB Migration Makefile
# 用于本地开发和 CI/CD 质量检查的便捷命令

.PHONY: help test test-coverage test-integration quality-check security-check format-check build clean lint fix-format fix-lint install-tools

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
APP_NAME := ybMigration
GO_VERSION := 1.25.1
COVERAGE_FILE := coverage.txt
COVERAGE_HTML := coverage.html
QUALITY_REPORT := quality-report.html
LINT_REPORT := lint-report.json

# 颜色定义
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

help: ## 显示帮助信息
	@echo "$(BLUE)YB Migration 开发工具$(NC)"
	@echo ""
	@echo "$(GREEN)质量检查命令:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(YELLOW)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST) | grep -E "(quality|security|format|lint|test)"
	@echo ""
	@echo "$(GREEN)构建和清理:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(YELLOW)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST) | grep -E "(build|clean|install)"
	@echo ""
	@echo "$(GREEN)修复命令:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(YELLOW)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST) | grep -E "(fix)"

# ============================================================================
# 质量检查命令
# ============================================================================

quality-check: ## 运行完整的代码质量检查
	@echo "$(BLUE)🔍 运行代码质量检查...$(NC)"
	@if [ ! -f .golangci.yml ]; then \
		echo "$(RED)❌ 错误: .golangci.yml 配置文件不存在$(NC)"; \
		exit 1; \
	fi
	@golangci-lint run --config .golangci.yml --timeout 15m ./...
	@echo "$(GREEN)✅ 代码质量检查通过$(NC)"

quality-report: ## 生成详细的质量报告
	@echo "$(BLUE)📊 生成质量报告...$(NC)"
	@golangci-lint run --config .golangci.yml --out-format=html --timeout 15m ./... > $(QUALITY_REPORT)
	@golangci-lint run --config .golangci.yml --out-format=json --timeout 15m ./... > quality-report.json
	@echo "$(GREEN)✅ 质量报告已生成: $(QUALITY_REPORT)$(NC)"

security-check: ## 运行安全扫描
	@echo "$(BLUE)🔒 运行安全扫描...$(NC)"
	@golangci-lint run --config .golangci.yml --enable-only=gosec --timeout 10m ./...
	@echo "$(BLUE)🔍 检查敏感信息泄露...$(NC)"
	@if grep -r -i "password\|secret\|key\|token" --include="*.go" --include="*.yaml" --include="*.yml" --exclude-dir=.git . 2>/dev/null; then \
		echo "$(RED)⚠️  警告: 发现可能的敏感信息$(NC)"; \
	else \
		echo "$(GREEN)✅ 未发现敏感信息泄露$(NC)"; \
	fi
	@echo "$(GREEN)✅ 安全扫描完成$(NC)"

format-check: ## 检查代码格式
	@echo "$(BLUE)📝 检查代码格式...$(NC)"
	@if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "$(RED)❌ 以下文件需要格式化:$(NC)"; \
		gofmt -s -l .; \
		echo "$(YELLOW)💡 运行 'make fix-format' 自动修复$(NC)"; \
		exit 1; \
	else \
		echo "$(GREEN)✅ Go 格式检查通过$(NC)"; \
	fi
	@if command -v goimports >/dev/null 2>&1; then \
		if [ "$(goimports -l . | wc -l)" -gt 0 ]; then \
			echo "$(RED)❌ 以下文件的 import 语句需要格式化:$(NC)"; \
			goimports -l .; \
			echo "$(YELLOW)💡 运行 'make fix-format' 自动修复$(NC)"; \
			exit 1; \
		else \
			echo "$(GREEN)✅ Import 格式检查通过$(NC)"; \
		fi; \
	else \
		echo "$(YELLOW)⚠️  goimports 未安装，跳过 import 格式检查$(NC)"; \
	fi

lint: ## 运行轻量级 lint 检查（跳过格式问题）
	@echo "$(BLUE)🔍 运行轻量级 lint 检查...$(NC)"
	@golangci-lint run --config .golangci.yml --disable=godot,whitespace --timeout 10m ./...
	@echo "$(GREEN)✅ Lint 检查完成$(NC)"

lint-report: ## 生成 lint 报告
	@echo "$(BLUE)📊 生成 lint 报告...$(NC)"
	@golangci-lint run --config .golangci.yml --out-format=json --timeout 10m ./... > $(LINT_REPORT)
	@echo "$(GREEN)✅ Lint 报告已生成: $(LINT_REPORT)$(NC)"

# ============================================================================
# 测试命令
# ============================================================================

test: ## 运行单元测试
	@echo "$(BLUE)🧪 运行单元测试...$(NC)"
	@go test -v -race ./...

test-coverage: ## 运行测试并生成覆盖率报告
	@echo "$(BLUE)🧪 运行测试并生成覆盖率报告...$(NC)"
	@go test -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@go tool cover -func=$(COVERAGE_FILE)
	@COVERAGE=$$(go tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print $$3}' | sed 's/%//'); \
	if (( $$(echo "$$COVERAGE >= 60" | bc -l) )); then \
		echo "$(GREEN)✅ 覆盖率检查通过 ($$COVERAGE% ≥ 60%)$(NC)"; \
	else \
		echo "$(RED)❌ 覆盖率不达标 ($$COVERAGE% < 60%)$(NC)"; \
		exit 1; \
	fi

test-coverage-html: ## 生成 HTML 格式的覆盖率报告
	@echo "$(BLUE)📊 生成 HTML 覆盖率报告...$(NC)"
	@go test -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "$(GREEN)✅ 覆盖率报告已生成: $(COVERAGE_HTML)$(NC)"

test-integration: ## 运行集成测试
	@echo "$(BLUE)🔗 运行集成测试...$(NC)"
	@go test -v -race -tags=integration ./...
	@go test -v -race ./cmd/... -tags=integration

test-bench: ## 运行性能基准测试
	@echo "$(blue)⚡ 运行性能基准测试...$(NC)"
	@go test -bench=. -benchmem -run=^$$ ./...

# ============================================================================
# 构建和清理
# ============================================================================

build: ## 构建应用程序
	@echo "$(BLUE)🏗️  构建应用程序...$(NC)"
	@mkdir -p bin
	@go build -ldflags="-w -s" -o bin/$(APP_NAME) ./cmd
	@echo "$(GREEN)✅ 构建完成: bin/$(APP_NAME)$(NC)"
	@ls -la bin/

build-all: ## 构建多平台版本
	@echo "$(BLUE)🏗️  构建多平台版本...$(NC)"
	@mkdir -p bin
	# Linux AMD64
	@GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o bin/$(APP_NAME)-linux-amd64 ./cmd
	# Linux ARM64
	@GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o bin/$(APP_NAME)-linux-arm64 ./cmd
	# Windows AMD64
	@GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o bin/$(APP_NAME)-windows-amd64.exe ./cmd
	# macOS AMD64
	@GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o bin/$(APP_NAME)-darwin-amd64 ./cmd
	# macOS ARM64
	@GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o bin/$(APP_NAME)-darwin-arm64 ./cmd
	@echo "$(GREEN)✅ 多平台构建完成$(NC)"
	@ls -la bin/

clean: ## 清理构建产物和临时文件
	@echo "$(BLUE)🧹 清理构建产物...$(NC)"
	@rm -rf bin/
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML) $(QUALITY_REPORT) $(LINT_REPORT)
	@rm -f quality-report.json lint-report.json
	@go clean -cache
	@go clean -testcache
	@echo "$(GREEN)✅ 清理完成$(NC)"

# ============================================================================
# 修复命令
# ============================================================================

fix-format: ## 自动修复代码格式问题
	@echo "$(BLUE)🔧 修复代码格式...$(NC)"
	@gofmt -s -w .
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
		echo "$(GREEN)✅ 格式修复完成$(NC)"; \
	else \
		echo "$(YELLOW)⚠️  goimports 未安装，仅修复 gofmt 格式$(NC)"; \
	fi

fix-lint: ## 尝试自动修复 lint 问题
	@echo "$(BLUE)🔧 尝试自动修复 lint 问题...$(NC)"
	@if golangci-lint run --config .golangci.yml --fix ./...; then \
		echo "$(GREEN)✅ 自动修复完成$(NC)"; \
	else \
		echo "$(YELLOW)⚠️  部分问题需要手动修复$(NC)"; \
	fi

# ============================================================================
# 工具安装
# ============================================================================

install-tools: ## 安装开发工具
	@echo "$(BLUE)📦 安装开发工具...$(NC)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/sonatypecommunity/nancy/cmd/nancy@latest
	@echo "$(GREEN)✅ 开发工具安装完成$(NC)"

check-tools: ## 检查开发工具是否已安装
	@echo "$(BLUE)🔍 检查开发工具...$(NC)"
	@command -v golangci-lint >/dev/null 2>&1 && echo "$(GREEN)✅ golangci-lint$(NC)" || echo "$(RED)❌ golangci-lint 未安装$(NC)"
	@command -v goimports >/dev/null 2>&1 && echo "$(GREEN)✅ goimports$(NC)" || echo "$(YELLOW)⚠️  goimports 未安装$(NC)"
	@command -v nancy >/dev/null 2>&1 && echo "$(GREEN)✅ nancy$(NC)" || echo "$(YELLOW)⚠️  nancy 未安装$(NC)"

# ============================================================================
# 组合命令
# ============================================================================

ci-quality: ## 运行完整的 CI 质量检查流程
	@echo "$(BLUE)🚀 运行 CI 质量检查流程...$(NC)"
	@$(MAKE) quality-check
	@$(MAKE) security-check
	@$(MAKE) format-check
	@$(MAKE) test-coverage
	@echo "$(GREEN)🎉 所有质量检查通过！$(NC)"

ci-test: ## 运行完整的 CI 测试流程
	@echo "$(BLUE)🚀 运行 CI 测试流程...$(NC)"
	@$(MAKE) ci-quality
	@$(MAKE) test-integration
	@echo "$(GREEN)🎉 所有测试通过！$(NC)"

pre-commit: ## Git pre-commit 钩子检查
	@echo "$(BLUE)🔍 Pre-commit 检查...$(NC)"
	@$(MAKE) format-check
	@$(MAKE) lint
	@$(MAKE) test
	@echo "$(GREEN)✅ Pre-commit 检查通过$(NC)"

# ============================================================================
# 开发辅助
# ============================================================================

dev-setup: ## 初始化开发环境
	@echo "$(BLUE)🚀 初始化开发环境...$(NC)"
	@$(MAKE) install-tools
	@$(MAKE) check-tools
	@go mod download
	@go mod verify
	@echo "$(GREEN)✅ 开发环境初始化完成$(NC)"

status: ## 显示项目状态
	@echo "$(BLUE)📊 项目状态:$(NC)"
	@echo "Go 版本: $$(go version)"
	@echo "模块路径: $$(go list -m)"
	@echo "依赖数量: $$(go list -m all | wc -l)"
	@if [ -f $(COVERAGE_FILE) ]; then \
		COVERAGE=$$(go tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print $$3}'); \
		echo "当前覆盖率: $$COVERAGE"; \
	fi
	@echo "最后构建: $$(ls -la bin/ 2>/dev/null | grep $(APP_NAME) | awk '{print $$6, $$7, $$8}' || echo '未构建')"
