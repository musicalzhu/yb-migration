# YB Migration

YB Migration 是一个用于分析 MySQL 到 YB 数据库迁移兼容性的工具。它可以解析 SQL 语句、日志文件，并识别潜在的兼容性问题，提供详细的迁移建议。

## 功能特性

- **多格式输入支持**：支持 SQL 文件、MySQL General Log 和目录批量分析
- **智能兼容性检查**：检测语法、数据类型、函数等方面的兼容性问题
- **多格式报告输出**：支持 JSON、Markdown、HTML 格式的分析报告
- **可配置规则**：通过 YAML 配置文件自定义检查规则和建议
- **高性能解析**：基于 TiDB SQL 解析器的 AST 解析

## 项目结构

```
yb-migration/
├── cmd/                    # 命令行入口
│   ├── main.go            # 主程序入口
│   └── integration_test.go # 集成测试
├── configs/               # 配置文件
│   └── default.yaml       # 默认配置
├── internal/              # 内部模块
│   ├── analyzer/          # 分析器核心
│   ├── checker/           # 兼容性检查器
│   ├── config/            # 配置管理
│   ├── input-parser/      # 输入解析器
│   ├── model/             # 数据模型
│   ├── report/            # 报告生成器
│   ├── sql-parser/        # SQL 解析器
│   └── testutils/         # 测试工具
├── testdata/              # 测试数据
│   ├── mysql_queries.sql  # 示例 SQL
│   └── general_log_example.log # 示例日志
├── output-report/         # 报告输出目录
├── go.mod                 # Go 模块定义
├── go.sum                 # 依赖校验
├── README.md              # 项目文档
├── TESTING.md             # 测试文档
└── .gitlab-ci.yml         # CI/CD 配置
```

## 快速开始

### 安装

```bash
# 克隆项目
git clone <repository-url>
cd yb-migration

# 安装依赖
go mod download

# 构建项目
go build -o bin/ybMigration ./cmd
```

### 基本使用

```bash
# 分析 SQL 文件
./bin/ybMigration --config configs/default.yaml --path testdata/mysql_queries.sql

# 分析 MySQL General Log
./bin/ybMigration --config configs/default.yaml --path testdata/general_log_example.log

# 分析整个目录
./bin/ybMigration --config configs/default.yaml --path /path/to/sql/files

# 指定报告输出目录
./bin/ybMigration --config configs/default.yaml --path input.sql --reportPath ./reports
```

### 命令行参数

- `--config`: 配置文件路径（可选，默认查找 `configs/default.yaml`）
- `--path`: 待分析的文件或目录路径（必须）
- `--reportPath`: 报告输出目录（可选，默认为 `./output-report`）

## 配置说明

配置文件使用 YAML 格式，包含以下主要部分：

```yaml
report:
  formats: [json, markdown, html]  # 报告格式
parser:
  mode: ast                        # 解析模式
suggestions:
  GROUP_CONCAT: "使用 STRING_AGG(col, ',') 或自研聚合函数"
  IFNULL: "使用 COALESCE(expr1, expr2)"
  NOW: "使用 CURRENT_TIMESTAMP(6) 或 LOCALTIMESTAMP(6) 以保留精度"
```

### 检查器类型

项目包含多种检查器：

1. **语法检查器** (`syntax_checker`)：检查 SQL 语法兼容性
2. **数据类型检查器** (`datatype_checker`)：检查数据类型兼容性
3. **函数检查器** (`function_checker`)：检查函数使用兼容性

## 报告格式

工具会生成多种格式的报告：

- **JSON**：结构化数据，便于程序处理
- **Markdown**：可读性强的文档格式
- **HTML**：可视化报告，便于浏览器查看

报告包含以下信息：
- 原始 SQL 语句
- 发现的兼容性问题
- 问题严重级别
- 修复建议
- 来源文件信息

## 开发指南

### 运行测试

```bash
# 运行所有测试
go test -v ./...

# 运行单元测试
go test -v -race ./...

# 运行集成测试
go test -v -tags=integration ./...

# 运行性能测试
go test -bench=. -benchmem ./...
```

### 代码检查

```bash
# 运行 golangci-lint
golangci-lint run ./...

# 格式化代码
go fmt ./...
```

### 添加新的检查器

1. 在 `internal/checker/` 目录下创建新的检查器文件
2. 实现 `checker.Checker` 接口
3. 在 `internal/checker/checker.go` 中注册新检查器
4. 添加相应的测试用例

## 依赖项

- `github.com/pingcap/tidb/pkg/parser`: SQL 解析器
- `github.com/stretchr/testify`: 测试框架
- Go 1.25.1+

## 许可证

[请添加许可证信息]

## 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 支持

如有问题或建议，请提交 Issue 或联系维护团队。