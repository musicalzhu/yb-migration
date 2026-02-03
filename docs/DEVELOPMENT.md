# YB Migration 开发者指南

## 概述

YB Migration 是一个用于分析和迁移 SQL 语句到目标数据库兼容格式的 Go 语言 CLI 工具。本指南将帮助开发者理解项目架构、开发流程和最佳实践。

## 目录

- [快速开始](#快速开始)
- [项目架构](#项目架构)
- [开发环境设置](#开发环境设置)
- [开发工作流程](#开发工作流程)
- [代码规范](#代码规范)
- [测试指南](#测试指南)
- [调试技巧](#调试技巧)
- [性能优化](#性能优化)
- [发布流程](#发布流程)
- [故障排除](#故障排除)
- [贡献指南](#贡献指南)

## 快速开始

### 环境要求

- **Go**: 1.25.1+
- **Git**: 2.0+
- **IDE**: 推荐 VS Code 或 GoLand

### 克隆项目

```bash
git clone https://github.com/musicalzhu/yb-migration.git
cd yb-migration
```

### 安装依赖

```bash
go mod tidy
```

### 构建项目

```bash
go build -o ybMigration cmd/main.go
```

### 运行测试

```bash
go test -v ./...
```

## 项目架构

### 目录结构

```
yb-migration/
├── cmd/                    # 命令行入口
│   └── main.go            # 主程序入口
├── internal/              # 内部包
│   ├── analyzer/          # 分析器核心
│   │   ├── analyzer.go
│   │   ├── analyzer_test.go
│   │   └── factory.go
│   ├── checker/           # 检查器实现
│   │   ├── checker.go
│   │   ├── checker_test.go
│   │   ├── function_checker.go
│   │   ├── datatype_checker.go
│   │   ├── syntax_checker.go
│   │   └── charset_checker.go
│   ├── config/            # 配置管理
│   │   ├── config.go
│   │   └── config_test.go
│   ├── constants/         # 常量定义
│   │   └── permissions.go
│   ├── input-parser/      # 输入解析
│   │   ├── input-parser.go
│   │   ├── sqlfile_parser.go
│   │   ├── general_log_parser.go
│   │   └── *_test.go
│   ├── model/             # 数据模型
│   │   └── models.go
│   ├── report/            # 报告生成
│   │   ├── generator.go
│   │   ├── html_generator.go
│   │   ├── json_generator.go
│   │   ├── markdown_generator.go
│   │   ├── sql_saver.go
│   │   └── utils.go
│   └── sql-parser/        # SQL 解析
│       ├── sql_parser.go
│       └── sql_parser_test.go
├── docs/                  # 文档
│   ├── API.md
│   ├── adr/              # 架构决策记录
│   └── DEVELOPMENT.md
├── scripts/               # 脚本文件
│   ├── pre-commit.sh
│   ├── setup-dev.sh
│   └── setup-gitlab.sh
├── .github/               # GitHub 配置
│   └── workflows/
├── .gitlab-ci.yml        # GitLab CI 配置
├── .golangci.yml         # golangci-lint 配置
├── go.mod                # Go 模块文件
├── go.sum                # 依赖校验文件
├── Makefile              # 构建脚本
├── README.md             # 项目说明
└── LICENSE               # 许可证
```

### 核心组件

#### 1. 分析器 (Analyzer)
负责协调整个分析流程，包括：
- 输入解析
- SQL 解析
- 兼容性检查
- 报告生成

#### 2. 检查器 (Checker)
实现各种兼容性检查：
- 函数兼容性检查
- 数据类型兼容性检查
- 语法兼容性检查
- 字符集兼容性检查

#### 3. 配置管理 (Config)
管理配置文件的加载和验证：
- YAML 配置解析
- 默认配置合并
- 配置验证

#### 4. 报告生成 (Report)
生成多种格式的分析报告：
- JSON 格式
- HTML 格式
- Markdown 格式
- SQL 格式

## 开发环境设置

### IDE 配置

#### VS Code
推荐安装以下扩展：
- Go (官方)
- GitLens
- YAML Support
- Better Comments

配置文件 `.vscode/settings.json`:
```json
{
    "go.useLanguageServer": true,
    "go.formatTool": "goimports",
    "go.lintTool": "golangci-lint",
    "go.lintFlags": ["--fast"],
    "go.testFlags": ["-v"],
    "go.coverOnSave": true,
    "go.coverageDecorator": {
        "type": "gutter",
        "coveredHighlightColor": "rgba(64,128,64,0.5)",
        "uncoveredHighlightColor": "rgba(128,64,64,0.25)"
    }
}
```

#### GoLand
1. 启用 Go Modules 支持
2. 配置代码格式化为 goimports
3. 启用 golangci-lint 集成
4. 配置测试运行器

### 开发工具

#### 必需工具
```bash
# 安装 golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 安装 swag (用于 API 文档)
go install github.com/swaggo/swag/cmd/swag@latest

# 安装 gci (导入分组)
go install github.com/daixiang0/gci@latest
```

#### 可选工具
```bash
# 安装 go-swagger (API 文档)
go install github.com/go-swagger/go-swagger/cmd/swagger@latest

# 安装 mockgen (模拟生成)
go install github.com/golang/mock/mockgen@latest

# 安装 golangci-lint 配置生成器
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 环境变量

创建 `.env` 文件（可选）：
```bash
# 开发环境配置
GO_VERSION=1.25.1
GOPROXY=https://goproxy.cn,direct
GOSUMDB=sum.golang.org

# 项目配置
YB_MIGRATION_LOG_LEVEL=debug
YB_MIGRATION_CONFIG_PATH=./config.yaml
```

## 开发工作流程

### 1. 创建功能分支

```bash
git checkout -b feature/new-checker
```

### 2. 开发代码

#### 添加新检查器
1. 在 `internal/checker/` 目录下创建新的检查器文件
2. 实现 `Checker` 接口
3. 添加单元测试
4. 更新配置文件

```go
// 示例：新的检查器
type NewChecker struct {
    name string
    config map[string]interface{}
}

func (c *NewChecker) Check(stmt model.SQLStatement) []model.Issue {
    // 实现检查逻辑
    return issues
}

func (c *NewChecker) GetName() string {
    return c.name
}

func (c *NewChecker) GetCategory() string {
    return "new_category"
}

func (c *NewChecker) GetDescription() string {
    return "新检查器描述"
}
```

#### 添加新报告格式
1. 在 `internal/report/` 目录下创建新的生成器
2. 实现 `Reporter` 接口
3. 添加单元测试
4. 注册到报告管理器

### 3. 运行测试

```bash
# 运行所有测试
go test -v ./...

# 运行特定包的测试
go test -v ./internal/checker

# 运行测试并生成覆盖率报告
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### 4. 代码检查

```bash
# 运行 golangci-lint
golangci-lint run

# 运行格式化检查
gci write -s standard -s default -s "prefix(github.com/example/ybMigration)" .

# 检查导入分组
goimports -w .
```

### 5. 提交代码

```bash
# 添加文件
git add .

# 提交（遵循提交信息规范）
git commit -m "feat: 添加新的检查器

- 实现 NewChecker 用于检查新功能
- 添加相关单元测试
- 更新配置文件模板

Closes #123"
```

### 6. 推送和创建 PR

```bash
# 推送到远程
git push origin feature/new-checker

# 创建 Pull Request
```

## 代码规范

### 命名规范

#### 包名
- 使用小写字母
- 简短、有意义
- 避免与标准库冲突

```go
// 好的包名
package analyzer
package checker
package config

// 不好的包名
package util
package common
package misc
```

#### 变量和函数名
- 使用驼峰命名法
- 导出的变量/函数首字母大写
- 私有的变量/函数首字母小写

```go
// 好的命名
var DefaultConfigPath = "./config.yaml"
func LoadConfig(path string) (*Config, error)
func validateInput(input string) error

// 不好的命名
var configPath = "./config.yaml"
func load(path string) (*Config, error)
func check(input string) error
```

#### 常量名
- 使用大写字母和下划线
- 分组相关的常量

```go
const (
    ExitSuccess     = 0
    ExitInvalidArgs = 1
    ExitConfigError = 2
    ExitAnalysisErr = 3
)
```

### 注释规范

#### 包注释
```go
// Package analyzer 提供 SQL 分析功能，包括语法解析、
// 兼容性检查和转换建议生成。
package analyzer
```

#### 函数注释
```go
// AnalyzeInput 分析输入文件并返回分析结果。
// 它支持 SQL 文件、日志文件和目录作为输入。
//
// 参数:
//   - inputPath: 输入文件或目录路径
//   - parser: SQL 解析器实例
//   - checkers: 检查器列表
//
// 返回:
//   - *AnalysisResult: 分析结果
//   - error: 错误信息
func AnalyzeInput(inputPath string, parser SQLParser, checkers []Checker) (*AnalysisResult, error) {
    // 实现
}
```

#### 结构体注释
```go
// Config 表示应用程序的配置信息。
type Config struct {
    // Rules 检查规则配置
    Rules []Rule `yaml:"rules"`
    
    // Reports 报告配置
    Reports Reports `yaml:"reports"`
    
    // Global 全局配置
    Global Global `yaml:"global"`
}
```

### 错误处理

#### 错误定义
```go
// 定义错误类型
type ConfigError struct {
    Path string
    Err  error
}

func (e *ConfigError) Error() string {
    return fmt.Sprintf("配置文件错误 %s: %v", e.Path, e.Err)
}

// 使用错误包装
return fmt.Errorf("加载配置失败: %w", err)
```

#### 错误处理模式
```go
// 检查错误
result, err := someFunction()
if err != nil {
    return nil, fmt.Errorf("someFunction 失败: %w", err)
}

// 处理特定错误类型
var configErr *ConfigError
if errors.As(err, &configErr) {
    // 处理配置错误
    log.Printf("配置文件问题: %s", configErr.Path)
}
```

### 并发安全

#### 使用互斥锁
```go
type SafeCounter struct {
    mu    sync.RWMutex
    count int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

func (c *SafeCounter) Value() int {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.count
}
```

#### 使用通道
```go
func processFiles(files []string, workers int) error {
    fileChan := make(chan string, len(files))
    errChan := make(chan error, len(files))
    
    // 启动工作协程
    for i := 0; i < workers; i++ {
        go func() {
            for file := range fileChan {
                if err := processFile(file); err != nil {
                    errChan <- err
                }
            }
        }()
    }
    
    // 发送文件
    for _, file := range files {
        fileChan <- file
    }
    close(fileChan)
    
    // 等待完成
    for i := 0; i < len(files); i++ {
        if err := <-errChan; err != nil {
            return err
        }
    }
    
    return nil
}
```

## 测试指南

### 单元测试

#### 测试文件命名
- 测试文件以 `_test.go` 结尾
- 与被测试文件在同一包中

#### 测试函数命名
```go
func TestFunctionName(t *testing.T) {
    // 测试实现
}

func TestFunctionName_EdgeCase(t *testing.T) {
    // 边界情况测试
}

func TestFunctionName_ErrorCase(t *testing.T) {
    // 错误情况测试
}
```

#### 测试示例
```go
func TestConfig_LoadConfig(t *testing.T) {
    tests := []struct {
        name     string
        path     string
        expected *Config
        wantErr  bool
    }{
        {
            name:    "valid config",
            path:    "testdata/config.yaml",
            expected: &Config{ /* 期望的配置 */ },
            wantErr: false,
        },
        {
            name:    "invalid config",
            path:    "testdata/invalid.yaml",
            expected: nil,
            wantErr:  true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := LoadConfig(tt.path)
            if (err != nil) != tt.wantErr {
                t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.expected) {
                t.Errorf("LoadConfig() = %v, want %v", got, tt.expected)
            }
        })
    }
}
```

### 集成测试

#### 测试数据库
```go
func TestIntegration_Analyzer(t *testing.T) {
    if testing.Short() {
        t.Skip("跳过集成测试")
    }
    
    // 设置测试环境
    config := setupTestConfig(t)
    defer cleanupTestConfig(t)
    
    // 执行测试
    result, err := AnalyzeInput("testdata/sample.sql", parser, checkers)
    require.NoError(t, err)
    assert.NotNil(t, result)
}
```

### 基准测试

```go
func BenchmarkAnalyzer_Analyze(b *testing.B) {
    analyzer := setupAnalyzer()
    input := "testdata/large.sql"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := analyzer.Analyze(input)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### 测试覆盖率

```bash
# 生成覆盖率报告
go test -coverprofile=coverage.out ./...

# 查看覆盖率
go tool cover -func=coverage.out

# 生成 HTML 报告
go tool cover -html=coverage.out -o coverage.html
```

## 调试技巧

### 使用 Delve

```bash
# 安装 Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 调试测试
dlv test ./internal/checker

# 调试主程序
dlv debug cmd/main.go
```

### 日志调试

```go
import "log"

// 添加调试日志
log.Printf("解析配置文件: %s", configPath)
log.Printf("检查器数量: %d", len(checkers))
log.Printf("分析结果: %+v", result)
```

### 性能分析

```bash
# CPU 分析
go test -cpuprofile=cpu.prof -bench=.

# 内存分析
go test -memprofile=mem.prof -bench=.

# 查看分析结果
go tool pprof cpu.prof
go tool pprof mem.prof
```

## 性能优化

### 内存优化

#### 避免内存泄漏
```go
// 使用 defer 确保资源释放
func processFile(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()
    
    // 处理文件
    return nil
}
```

#### 使用对象池
```go
var statementPool = sync.Pool{
    New: func() interface{} {
        return &SQLStatement{}
    },
}

func parseStatement(sql string) *SQLStatement {
    stmt := statementPool.Get().(*SQLStatement)
    defer statementPool.Put(stmt)
    
    // 重置并使用
    *stmt = SQLStatement{}
    // 解析逻辑
    return stmt
}
```

### 并发优化

#### 使用工作池
```go
type WorkerPool struct {
    workers int
    jobs    chan Job
    results chan Result
}

func (wp *WorkerPool) Start() {
    for i := 0; i < wp.workers; i++ {
        go wp.worker()
    }
}

func (wp *WorkerPool) worker() {
    for job := range wp.jobs {
        result := job.Execute()
        wp.results <- result
    }
}
```

## 发布流程

### 版本管理

#### 语义化版本
- 主版本号：不兼容的 API 修改
- 次版本号：向下兼容的功能性新增
- 修订号：向下兼容的问题修正

#### 标签管理
```bash
# 创建标签
git tag -a v2.0.0 -m "Release version 2.0.0"

# 推送标签
git push origin v2.0.0
```

### 构建发布

#### Makefile
```makefile
.PHONY: build test lint clean release

# 构建
build:
	go build -o bin/ybMigration cmd/main.go

# 测试
test:
	go test -v ./...

# 代码检查
lint:
	golangci-lint run

# 清理
clean:
	rm -rf bin/

# 发布
release: clean test lint build
	@echo "构建完成，准备发布"
```

#### 交叉编译
```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o bin/ybMigration-linux-amd64 cmd/main.go

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o bin/ybMigration-windows-amd64.exe cmd/main.go

# macOS AMD64
GOOS=darwin GOARCH=amd64 go build -o bin/ybMigration-darwin-amd64 cmd/main.go
```

## 故障排除

### 常见问题

#### 依赖问题
```bash
# 清理模块缓存
go clean -modcache

# 重新下载依赖
go mod download

# 更新依赖
go mod tidy
```

#### 编译问题
```bash
# 检查 Go 版本
go version

# 检查模块路径
go list -m

# 检查依赖关系
go mod graph
```

#### 测试问题
```bash
# 运行特定测试
go test -run TestFunctionName ./internal/checker

# 详细测试输出
go test -v ./internal/checker

# 跳过缓存
go test -count=1 ./internal/checker
```

### 性能问题

#### 内存使用
```bash
# 查看内存统计
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```

#### CPU 使用
```bash
# 查看 CPU 分析
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof
```

## 贡献指南

### 提交信息规范

使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

#### 类型
- `feat`: 新功能
- `fix`: 修复 bug
- `docs`: 文档更新
- `style`: 代码格式化
- `refactor`: 重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动

#### 示例
```
feat(checker): 添加新的函数兼容性检查器

- 实现 FunctionChecker 用于检查 MySQL 函数兼容性
- 添加相关单元测试
- 更新配置文件模板

Closes #123
```

### Pull Request 流程

1. **Fork 项目**
2. **创建功能分支**
3. **开发代码**
4. **编写测试**
5. **运行检查**
6. **提交代码**
7. **创建 PR**
8. **代码审查**
9. **合并代码**

### 代码审查清单

#### 功能性
- [ ] 功能是否按预期工作
- [ ] 边界情况是否处理
- [ ] 错误处理是否完善

#### 代码质量
- [ ] 代码是否清晰易读
- [ ] 命名是否规范
- [ ] 注释是否充分

#### 测试
- [ ] 测试覆盖率是否足够
- [ ] 测试是否有意义
- [ ] 是否有集成测试

#### 性能
- [ ] 是否有性能问题
- [ ] 内存使用是否合理
- [ ] 并发是否安全

---

## 📚 **相关资源**

### 官方文档
- [Go 官方文档](https://golang.org/doc/)
- [Go Modules 文档](https://golang.org/cmd/go/#hdr-Modules__module_versions_and_more)
- [golangci-lint 配置](https://golangci-lint.run/)

### 项目文档
- [API 文档](API.md)
- [架构决策记录](adr/)
- [测试指南](TESTING.md)

### 工具和库
- [TiDB Parser](https://github.com/pingcap/tidb/tree/master/pkg/parser)
- [YAML 库](https://github.com/go-yaml/yaml)
- [Testify](https://github.com/stretchr/testify)

---

*最后更新: 2026-02-03*  
*维护者: YB Migration Team*
