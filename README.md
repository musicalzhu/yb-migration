# YB Migration

YB Migration 是一个用于分析 MySQL 到 YB 数据库迁移兼容性的工具。它可以解析 SQL 语句、MySQL General Log 日志文件，并识别潜在的兼容性问题，提供详细的迁移建议。

本文档面向开发者，重点描述项目架构、扩展点（checker / rules）、以及开发调试与测试方式。

## 功能特性

- **多格式输入支持**：支持 SQL 文件（.sql）、MySQL General Log（.log）和目录批量分析
- **智能兼容性检查**：检测语法、数据类型、函数等方面的兼容性问题
- **多格式报告输出**：支持 JSON、Markdown、HTML 格式的分析报告
- **可配置规则**：通过 YAML 配置文件自定义检查规则和建议
- **高性能解析**：基于 TiDB SQL 解析器的 AST 解析
- **AST 转换与优化**：智能 AST 节点转换，确保 SQL 格式正确、关键字大写、标识符反引号
- **SQL 质量保证**：确保转换后的 SQL 格式正确、关键字大写、标识符反引号
- **统一报告接口**：简化的报告生成接口，支持多种输出格式

## 项目结构

```
yb-migration/
├── cmd/                    # 命令行入口
│   ├── main.go            # 主程序入口
│   └── main_integration_test.go # 集成测试
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

## 开发者上手

### 环境要求

- Go 1.25.1+

### 构建与运行

```bash
# 安装依赖
go mod download

# 构建
go build -o bin/ybMigration ./cmd

# 运行（示例：分析 SQL 文件 / 日志 / 目录）
./bin/ybMigration --config configs/default.yaml --path testdata/mysql_queries.sql
./bin/ybMigration --config configs/default.yaml --path testdata/general_log_example.log
./bin/ybMigration --config configs/default.yaml --path testdata
```

### 输入类型与文件类型约束

- `internal/analyzer.AnalyzeInput` 支持：
  - `string`：
    - 若路径存在：作为文件/目录输入处理
    - 若路径不存在：作为 SQL 字符串处理
  - `io.Reader`：读取后作为 SQL 字符串处理
- **文件类型仅支持**：`.sql` 与 `.log`
- 目录遍历仅分析 `.sql/.log`，其他文件会跳过

## 配置说明

默认配置文件为 `configs/default.yaml`。

目前配置文件的核心是 `rules` 列表（与仓库中的默认配置保持一致）：

```yaml
rules:
  - name: "TINYINT_to_SMALLINT"
    description: "MySQL TINYINT 转换为标准 SMALLINT"
    category: "datatype"
    when:
      pattern: "TINYINT"
    then:
      action: "replace_type"
      target: "SMALLINT"
      mapping:
        - from: "TINYINT"
          to: "SMALLINT"
```

其中：

- `category` 当前支持：`datatype` / `function` / `syntax` / `charset`
- `when.pattern` 表示规则触发的匹配模式
- `then` 表示规则触发后的动作与映射

默认配置文件路径查找逻辑见：`internal/config.GetDefaultConfigPath()`。

### 检查器类型

项目包含多种检查器：

1. **语法检查器**（category: `syntax`）：检查 SQL 语法兼容性
2. **数据类型检查器**（category: `datatype`）：检查数据类型兼容性
3. **函数检查器**（category: `function`）：检查函数使用兼容性
4. **字符集检查器**（category: `charset`）：检查字符集与排序规则兼容性

检查器的创建入口位于 `internal/analyzer` 的 `Factory.CreateCheckers(...)`。

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

默认报告输出目录由 `internal/config.GetDefaultReportPath()` 决定：

- 优先当前工作目录的 `./output-report`
- 兜底可执行文件目录的 `./output-report`

## API 使用指南

### 推荐的使用方式

```go
// 1. 创建分析器工厂
factory, err := NewAnalyzerFactory("")
if err != nil {
    return fmt.Errorf("创建工厂失败: %w", err)
}

// 2. 创建检查器
checkers, err := factory.CreateCheckers("datatype", "function")
if err != nil {
    return fmt.Errorf("创建检查器失败: %w", err)
}

// 3. 创建 SQL 解析器
sqlParser := sqlparser.NewSQLParser()

// 4. 字符串分析器
stringAnalyzer, err := NewSQLAnalyzer(
    inputparser.NewStringParser(), 
    sqlParser, 
    checkers,
)
if err != nil {
    return fmt.Errorf("创建字符串分析器失败: %w", err)
}

// 5. 文件分析器
fileAnalyzer, err := NewSQLAnalyzer(
    inputparser.NewSQLFileParser(), 
    sqlParser, 
    checkers,
)
if err != nil {
    return fmt.Errorf("创建文件分析器失败: %w", err)
}

// 6. 分析 SQL 字符串
result, err := stringAnalyzer.AnalyzeSQL("SELECT * FROM users", "input_string")
if err != nil {
    var analysisErr *model.AnalysisError
    if errors.As(err, &analysisErr) {
        switch analysisErr.Type {
        case model.ErrorTypeParse:
            // 处理解析错误
        case model.ErrorTypeNoSQL:
            // 处理无SQL错误
        case model.ErrorTypeTransform:
            // 处理转换错误
        }
    }
    return err
}

// 7. 分析文件或目录（推荐使用 AnalyzeInput）
result, err := analyzer.AnalyzeInput(filePath, sqlParser, checkers)
if err != nil {
    return fmt.Errorf("分析输入失败: %w", err)
}
```

### 错误处理最佳实践

```go
// 使用 errors.As 检查具体错误类型
var analysisErr *model.AnalysisError
if errors.As(err, &analysisErr) {
    switch analysisErr.Type {
    case model.ErrorTypeParse:
        log.Printf("SQL 解析失败: %s, 源文件: %s", analysisErr.Message, analysisErr.Source)
    case model.ErrorTypeNoSQL:
        log.Printf("未找到有效 SQL: %s", analysisErr.Message)
    case model.ErrorTypeTransform:
        log.Printf("SQL 转换失败: %s", analysisErr.Message)
    }
}

// 使用 errors.Is 检查预定义错误
if errors.Is(err, model.ErrParse) {
    // 处理解析错误
}

## 架构与数据流（概览)

核心流程可以概括为：

1. `AnalyzeInput` 识别输入（文件/目录/SQL 字符串/Reader})
2. `input-parser` 解析输入为 SQL 文本（`.sql` 读取；`.log` 提取 Query 语句；字符串直接透传）
3. `sql-parser` 将 SQL 文本解析为 AST
4. `checker` 在 AST 上执行兼容性检查与转换
5. `analyzer.generateSQL` 将转换后的 AST 重新生成为优化后的 SQL
6. `report` 生成多格式输出

对应目录：

- `internal/analyzer`：入口与编排（选择 parser、组织 checker、聚合结果、SQL 生成）
- `internal/input-parser`：输入解析（`.sql` / `.log` / string）
- `internal/sql-parser`：SQL AST 解析
- `internal/checker`：规则检查与转换
- `internal/report`：报告生成

### 包职责

- `internal/analyzer`：分析器核心，负责输入识别、parser 选择、checker 组织、结果聚合、SQL 生成
- `internal/input-parser`：输入解析，负责将输入转换为 SQL 文本
- `internal/sql-parser`：SQL AST 解析，负责将 SQL 文本转换为 AST
- `internal/checker`：规则检查与转换，负责在 AST 上执行兼容性检查与转换
- `internal/report`：报告生成，负责生成多格式输出

### 核心数据流

- 输入识别与 parser 选择
- SQL 文本解析为 AST
- AST 上的兼容性检查与转换
- **转换后的 AST 重新生成为优化 SQL**
- 多格式报告生成

### AST 转换与 SQL 生成详解

#### AST 解析过程
```go
// 1. SQL 文本解析为 AST
stmts, err := sqlParser.ParseSQL(sqlText)
// 返回 []ast.StmtNode - 抽象语法树节点列表
```

#### AST 转换过程
```go
// 2. 检查器在 AST 上执行转换
checkResult := checker.Check(stmts, a.checkers...)
// 返回包含转换后 AST 的结果
// - TransformedStmts: 转换后的 AST 节点
// - Issues: 发现的兼容性问题
```

#### SQL 生成过程
```go
// 3. 将转换后的 AST 重新生成为 SQL
transformedSQL, err := a.generateSQL(checkResult.TransformedStmts)
// 使用 TiDB 的 Restore API 将 AST 转换回 SQL 文本
```

#### SQL 生成优化特性
1. **关键字大写**：自动将 SQL 关键字转换为大写格式
2. **标识符反引号**：为表名、字段名添加反引号保护
3. **格式标准化**：统一的空格、换行和缩进格式
4. **字符串优化**：使用 `strings.Builder` 提高字符串拼接性能

#### SQL 生成示例
```sql
-- 输入 SQL
select * from users where name = 'test'

-- 输出 SQL（优化后）
SELECT * FROM `users` WHERE `name`='test'
```

#### 性能优化
- **预分配容量**：根据 AST 节点数量预分配 slice 容量
- **并发安全**：每个 SQL 生成操作独立，支持并发处理
- **内存管理**：及时释放 AST 节点，避免内存泄漏

### checker 扩展指南

1. 在 `internal/checker/` 中新增 checker（实现 `checker.Checker` 接口）
2. 为新 checker 增加构造函数（与现有 `NewDataTypeChecker` 等保持一致风格）
3. 在 `internal/analyzer` 的 `Factory.CreateCheckers(...)` 中增加分支注册（category 名称建议全小写）
4. 在 `configs/default.yaml` 中补充对应 `category` 的规则（如需要）
5. 增加/更新单元测试（建议新增到对应包的 `*_test.go`）

## 开发指南

### 运行测试

```bash
# 运行所有测试
go test -v ./...

# 运行单元测试
go test -v -race ./...

# 运行集成测试
go test -v ./cmd

# 运行性能测试
go test -bench=. -benchmem ./...
```

### 调试建议

- 从 `cmd/main.go` 的命令行入口开始断点调试
- 重点关注：
  - `internal/analyzer.AnalyzeInput`
  - `internal/input-parser` 中的 `.sql/.log` 解析器
  - `internal/checker.Check`

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