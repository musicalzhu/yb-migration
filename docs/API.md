# YB Migration API 文档

## 概述

YB Migration 是一个命令行工具，用于分析和迁移 SQL 语句到目标数据库兼容格式。本文档描述了工具的编程接口和使用方法。

## 目录

- [安装和配置](#安装和配置)
- [命令行接口](#命令行接口)
- [编程接口](#编程接口)
- [配置接口](#配置接口)
- [分析器接口](#分析器接口)
- [报告生成接口](#报告生成接口)
- [错误处理](#错误处理)
- [示例代码](#示例代码)

## 安装和配置

### 安装

```bash
git clone https://github.com/musicalzhu/yb-migration.git
cd yb-migration
go mod tidy
go build -o ybMigration cmd/main.go
```

### 配置文件

配置文件使用 YAML 格式，默认查找位置：
1. `./config.yaml`
2. `~/.yb-migration/config.yaml`
3. `/etc/yb-migration/config.yaml`

## 命令行接口

### 基本用法

```bash
# 基本分析
./ybMigration /path/to/sql/file.sql

# 指定配置文件
./ybMigration --config /path/to/config.yaml /path/to/sql/file.sql

# 指定输出目录
./ybMigration --reportPath /path/to/output /path/to/sql/file.sql

# 完整参数
./ybMigration --config config.yaml --path input.sql --reportPath ./reports
```

### 命令行参数

| 参数 | 类型 | 必需 | 默认值 | 描述 |
|------|------|------|--------|------|
| `--config` | string | 否 | 自动查找 | 配置文件路径 |
| `--path` | string | 是 | - | 待分析的SQL文件、日志文件或目录路径 |
| `--reportPath` | string | 否 | `./output-report` | 分析报告输出目录 |

### 退出状态码

| 状态码 | 含义 |
|--------|------|
| 0 | 成功 |
| 1 | 参数错误 |
| 2 | 配置错误 |
| 3 | 分析错误 |

## 编程接口

### 主要包结构

```
github.com/example/ybMigration/
├── cmd/                    # 命令行入口
├── internal/
│   ├── analyzer/          # 分析器核心
│   ├── checker/           # 检查器实现
│   ├── config/            # 配置管理
│   ├── input-parser/      # 输入解析
│   ├── model/             # 数据模型
│   ├── report/            # 报告生成
│   └── sql-parser/        # SQL解析
└── docs/                  # 文档
```

### 核心接口

#### Analyzer 接口

```go
type Analyzer interface {
    Analyze(input string, parser SQLParser, checkers []Checker) (*AnalysisResult, error)
}

type AnalysisResult struct {
    InputPath     string
    SQLStatements []SQLStatement
    Issues        []Issue
    Transformations []Transformation
    Summary       AnalysisSummary
}
```

#### Checker 接口

```go
type Checker interface {
    Check(stmt SQLStatement) []Issue
    GetName() string
    GetCategory() string
    GetDescription() string
}
```

#### SQLParser 接口

```go
type SQLParser interface {
    Parse(sql string) ([]SQLStatement, error)
    ParseFile(filePath string) ([]SQLStatement, error)
}
```

## 配置接口

### 配置结构

```go
type Config struct {
    Rules []Rule `yaml:"rules"`
}

type Rule struct {
    Name        string   `yaml:"name"`
    Category    string   `yaml:"category"`
    Description string   `yaml:"description"`
    Enabled     bool     `yaml:"enabled"`
    Severity    string   `yaml:"severity"`
    Parameters  map[string]interface{} `yaml:"parameters"`
}
```

### 配置管理

```go
// 加载配置
config, err := config.LoadConfig("/path/to/config.yaml")

// 获取默认配置路径
configPath, err := config.GetDefaultConfigPath()

// 按类别获取规则
functionRules := config.GetRulesByCategory("function")
datatypeRules := config.GetRulesByCategory("datatype")
```

## 分析器接口

### 分析器工厂

```go
// 创建分析器工厂
factory, err := analyzer.NewAnalyzerFactory("/path/to/config.yaml")

// 从配置创建检查器
checkers, err := factory.CreateCheckersFromConfig()

// 执行分析
result, err := analyzer.AnalyzeInput("/path/to/input", sqlParser, checkers)
```

### 内置检查器

| 检查器 | 类别 | 描述 |
|--------|------|------|
| `FunctionChecker` | function | 检查不兼容的函数调用 |
| `DatatypeChecker` | datatype | 检查不兼容的数据类型 |
| `SyntaxChecker` | syntax | 检查语法兼容性 |
| `CharsetChecker` | charset | 检查字符集兼容性 |

## 报告生成接口

### 支持的报告格式

| 格式 | 描述 | 文件扩展名 |
|------|------|------------|
| JSON | 结构化数据 | `.json` |
| HTML | 可视化报告 | `.html` |
| Markdown | 文档报告 | `.md` |
| SQL | 转换后的SQL | `.sql` |

### 报告生成

```go
// 生成所有格式的报告
err := report.GenerateReports("/path/to/output", result, config, checkers)

// 生成特定格式的报告
jsonReport, err := report.GenerateJSON(result, config)
htmlReport, err := report.GenerateHTML(result, config)
mdReport, err := report.GenerateMarkdown(result, config)

// 保存转换后的SQL
err := report.SaveTransformedSQL(result, "/path/to/output.sql")
```

## 错误处理

### 错误类型

```go
// 配置错误
type ConfigError struct {
    Path string
    Err  error
}

// 分析错误
type AnalysisError struct {
    Input string
    Err   error
}

// 解析错误
type ParseError struct {
    SQL string
    Err error
}
```

### 错误处理示例

```go
result, err := analyzer.AnalyzeInput(inputPath, parser, checkers)
if err != nil {
    switch {
    case errors.Is(err, &config.ConfigError{}):
        log.Fatal("配置错误:", err)
    case errors.Is(err, &analyzer.AnalysisError{}):
        log.Fatal("分析错误:", err)
    case errors.Is(err, &sqlparser.ParseError{}):
        log.Fatal("解析错误:", err)
    default:
        log.Fatal("未知错误:", err)
    }
}
```

## 示例代码

### 基本使用示例

```go
package main

import (
    "log"
    "github.com/example/ybMigration/internal/analyzer"
    "github.com/example/ybMigration/internal/config"
    "github.com/example/ybMigration/internal/report"
    sqlparser "github.com/example/ybMigration/internal/sql-parser"
)

func main() {
    // 1. 加载配置
    configPath, err := config.GetDefaultConfigPath()
    if err != nil {
        log.Fatal(err)
    }

    // 2. 创建分析器工厂
    factory, err := analyzer.NewAnalyzerFactory(configPath)
    if err != nil {
        log.Fatal(err)
    }

    // 3. 创建检查器
    checkers, err := factory.CreateCheckersFromConfig()
    if err != nil {
        log.Fatal(err)
    }

    // 4. 创建SQL解析器
    parser := sqlparser.NewSQLParser()

    // 5. 执行分析
    result, err := analyzer.AnalyzeInput("input.sql", parser, checkers)
    if err != nil {
        log.Fatal(err)
    }

    // 6. 生成报告
    err = report.GenerateReports("./output", result, factory.GetConfig(), checkers)
    if err != nil {
        log.Fatal(err)
    }

    log.Println("分析完成！")
}
```

### 自定义检查器示例

```go
package main

import (
    "github.com/example/ybMigration/internal/checker"
    "github.com/example/ybMigration/internal/model"
)

type CustomChecker struct {
    name string
}

func (c *CustomChecker) Check(stmt model.SQLStatement) []model.Issue {
    // 实现自定义检查逻辑
    var issues []model.Issue
    
    // 检查逻辑...
    
    return issues
}

func (c *CustomChecker) GetName() string {
    return c.name
}

func (c *CustomChecker) GetCategory() string {
    return "custom"
}

func (c *CustomChecker) GetDescription() string {
    return "自定义检查器"
}

func main() {
    customChecker := &CustomChecker{name: "MyCustomChecker"}
    // 使用自定义检查器...
}
```

## 版本信息

- **当前版本**: v2.0
- **Go 版本**: 1.25.1+
- **许可证**: MIT

## 支持和反馈

- **GitHub**: https://github.com/musicalzhu/yb-migration
- **Issues**: https://github.com/musicalzhu/yb-migration/issues
- **文档**: https://github.com/musicalzhu/yb-migration/docs

---

*最后更新: 2026-02-03*
