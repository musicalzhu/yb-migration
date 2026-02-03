# YB Migration

YB Migration 是一个用于分析 MySQL 到 YB 数据库迁移兼容性的工具。它可以解析 SQL 语句、MySQL General Log 日志文件，并识别潜在的兼容性问题，提供详细的迁移建议。

**版本**: v2.0 - 零 lint 问题版本  
**最后更新**: 2026-02-03  
**状态**: 生产就绪，企业级标准

---

## 🎯 功能特性

- **多格式输入支持**：支持 SQL 文件（.sql）、MySQL General Log（.log）和目录批量分析
- **智能兼容性检查**：检测语法、数据类型、函数等方面的兼容性问题
- **多格式报告输出**：支持 JSON、Markdown、HTML 格式的分析报告
- **可配置规则**：通过 YAML 配置文件自定义检查规则和建议
- **高性能解析**：基于 TiDB SQL 解析器的 AST 解析
- **AST 转换与优化**：智能 AST 节点转换，确保 SQL 格式正确、关键字大写、标识符反引号
- **SQL 质量保证**：确保转换后的 SQL 格式正确、关键字大写、标识符反引号
- **统一报告接口**：简化的报告生成接口，支持多种输出格式

---

## 🏗️ 项目结构

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
│   ├── constants/         # 常量定义 (新增)
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
├── docs/                  # 项目文档
│   ├── API.md             # API 文档
│   ├── ADR/               # 架构决策记录
│   ├── DEVELOPMENT.md     # 开发者指南
│   ├── WORKFLOW.md        # 开发工作流程
│   └── TESTING.md         # 测试指南
└── .gitlab-ci.yml         # CI/CD 配置
```

---

## 🚀 快速开始

### 环境要求

- **Go 版本**: 1.25.1 或更高版本
- **操作系统**: Windows、Linux、macOS
- **依赖**: 自动通过 Go modules 管理

### 安装

```bash
# 克隆项目
git clone <repository-url>
cd yb-migration

# 安装依赖
go mod download

# 编译项目
go build -o bin/yb-migration ./cmd
```

### 基本使用

```bash
# 分析 SQL 文件
./bin/yb-migration -f testdata/mysql_queries.sql

# 分析日志文件
./bin/yb-migration -f testdata/general_log_example.log

# 批量分析目录
./bin/yb-migration -d ./sql-files/

# 使用自定义配置
./bin/yb-migration -c configs/custom.yaml -f input.sql

# 生成 HTML 报告
./bin/yb-migration -f input.sql -o output-report/ --format html
```

---

## 📖 文档

### 📋 项目文档
- [README.md](README.md) - 项目介绍
- [Quality-Gate-Guide.md](Quality-Gate-Guide.md) - CI/CD 质量门禁完整指南
- [TESTING.md](TESTING.md) - 测试指南
- [GitLab-Community-Guide.md](GitLab-Community-Guide.md) - GitLab 社区版部署指南
- [golangci-config-review.md](golangci-config-review.md) - golangci-lint 配置审查报告

### 📊 项目统计
- [PROJECT_STATS.md](PROJECT_STATS.md) - 项目统计报告
- [CODE_REVIEW_REPORT.md](CODE_REVIEW_REPORT.md) - 代码审查报告

---

## 🔧 开发者上手

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

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### 代码检查

```bash
# 运行 golangci-lint
golangci-lint run ./...

# 格式化代码
gofmt ./...
```

### 构建项目

```bash
# 本地构建
go build -o bin/yb-migration ./cmd

# 交叉编译
GOOS=linux GOARCH=amd64 go build -o bin/yb-migration-linux ./cmd
GOOS=windows GOARCH=amd64 go build -o bin/yb-migration.exe ./cmd
```

---

## 🏗️ 架构设计

### 核心组件

#### 1. 分析器 (Analyzer)
- **SQLAnalyzer**: 主要的分析器实现
- **功能**: 协调各个组件完成 SQL 分析流程

#### 2. 检查器 (Checker)
- **DataTypeChecker**: 数据类型兼容性检查
- **FunctionChecker**: 函数兼容性检查
- **SyntaxChecker**: SQL 语法检查
- **CharsetChecker**: 字符集兼容性检查

#### 3. 解析器 (Parser)
- **SQLParser**: 基于 TiDB Parser 的 SQL 解析器
- **SQLFileParser**: SQL 文件解析器
- **GeneralLogFileParser**: MySQL General Log 解析器
- **StringParser**: 字符串解析器

#### 4. 报告生成器 (Generator)
- **JSONGenerator**: JSON 格式报告生成
- **MarkdownGenerator**: Markdown 格式报告生成
- **HTMLGenerator**: HTML 格式报告生成

### 设计模式

- **工厂模式**: AnalyzerFactory 创建检查器
- **策略模式**: 多种检查器实现
- **访问者模式**: AST 遍历和检查
- **模板方法模式**: 报告生成器

---

## 📝 配置说明

### 配置文件结构

```yaml
# configs/default.yaml
rules:
  datatype:
    - pattern: "TINYINT"
      suggestion: "建议使用 SMALLINT 替代 TINYINT"
      severity: "warning"
      description: "TINYINT 在 YB 中可能有性能问题"
    
  function:
    - pattern: "NOW()"
      suggestion: "使用 CURRENT_TIMESTAMP 替代 NOW()"
      severity: "info"
      description: "NOW() 函数在 YB 中的行为可能不同"
    
  syntax:
    - pattern: "ENGINE=InnoDB"
      suggestion: "YB 不支持 ENGINE 选项"
      severity: "error"
      description: "YB 会自动处理存储引擎"

output:
  format: "json"  # json, markdown, html
  path: "./output-report"
  include-suggestions: true
  include-transformed-sql: true
```

### 自定义规则

```yaml
# 添加自定义规则
rules:
  custom:
    - pattern: "OLD_PASSWORD()"
      suggestion: "使用 PASSWORD() 替代 OLD_PASSWORD()"
      severity: "error"
      description: "OLD_PASSWORD() 函数已弃用"
```

---

## 📊 质量指标

### 代码质量
- **Lint 问题**: 0 个 (完美状态)
- **测试覆盖率**: 28.8% (核心模块良好)
- **代码行数**: 3,777 行 (业务 2,024 行，测试 1,753 行)
- **函数数量**: 129 个 (业务 90 个，测试 39 个)

### 高覆盖率模块
- **internal/config**: 84.2%
- **internal/input-parser**: 80.8%
- **internal/sql-parser**: 66.7%

### CI/CD 状态
- **质量门禁**: 35 个 linters，零问题
- **测试通过率**: 100%
- **构建状态**: 成功
- **部署状态**: 就绪

---

## 🔄 CI/CD 集成

### GitLab CI/CD

项目包含完整的 GitLab CI/CD 流水线配置：

```yaml
# .gitlab-ci.yml
stages:
  - prepare
  - quality
  - test
  - security
  - build
  - deploy
  - notify
```

### 质量门禁

- **代码检查**: golangci-lint (35 个 linters)
- **测试覆盖**: 自动化测试和覆盖率报告
- **安全扫描**: gosec 安全检查
- **格式检查**: gci + gofmt 自动格式化

### 报告生成

- **Lint 报告**: HTML + JSON + Checkstyle 格式
- **覆盖率报告**: HTML 可视化报告
- **质量指标**: 实时质量统计

---

## 🎯 添加新的检查器

### 1. 创建检查器文件

```go
// internal/checker/new_checker.go
package checker

import (
    "github.com/example/ybMigration/internal/model"
)

type NewChecker struct {
    rules []model.Rule
}

func NewNewChecker(rules []model.Rule) *NewChecker {
    return &NewChecker{rules: rules}
}

func (c *NewChecker) Name() string {
    return "new_checker"
}

func (c *NewChecker) Inspect(node interface{}) []model.Issue {
    // 实现检查逻辑
    return issues
}
```

### 2. 注册检查器

```go
// internal/analyzer/factory.go
func (f *AnalyzerFactory) CreateCheckers(categories ...string) ([]checker.Checker, error) {
    var checkers []checker.Checker
    
    for _, category := range categories {
        switch category {
        case "new_checker":
            rules := f.config.GetRulesByCategory("new_checker")
            checkers = append(checkers, checker.NewNewChecker(rules))
        // ... 其他检查器
        }
    }
    
    return checkers, nil
}
```

### 3. 添加测试

```go
// internal/checker/new_checker_test.go
func TestNewChecker_Check(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected []model.Issue
    }{
        // 测试用例...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 测试逻辑
        })
    }
}
```

---

## 🐛 故障排除

### 常见问题

#### 1. 解析错误
```bash
# 检查 SQL 语法
./bin/yb-migration -f invalid.sql --debug

# 查看详细错误信息
./bin/yb-migration -f input.sql --verbose
```

#### 2. 配置问题
```bash
# 验证配置文件
./bin/yb-migration --validate-config configs/custom.yaml

# 查看默认配置
./bin/yb-migration --show-default-config
```

#### 3. 性能问题
```bash
# 启用性能分析
./bin/yb-migration -f large.sql --profile

# 调整并发数
./bin/yb-migration -f large.sql --workers 4
```

### 调试模式

```bash
# 启用详细日志
./bin/yb-migration -f input.sql --debug --verbose

# 生成调试报告
./bin/yb-migration -f input.sql --debug-report debug.json
```

---

## 🤝 贡献指南

### 开发流程

1. **Fork 项目**
2. **创建功能分支**
   ```bash
   git checkout -b feature/new-feature
   ```
3. **编写代码**
4. **添加测试**
5. **运行检查**
   ```bash
   go test -v ./...
   golangci-lint run ./...
   ```

### 文档资源

- **[API 文档](docs/API.md)**: 详细的 API 接口文档
- **[开发者指南](docs/DEVELOPMENT.md)**: 完整的开发指南和最佳实践
- **[开发工作流程](docs/WORKFLOW.md)**: 详细的开发流程和 CI/CD 配置
- **[架构决策记录](docs/adr/)**: 重要的架构决策和设计选择
- **[测试指南](docs/TESTING.md)**: 测试策略和最佳实践
6. **提交更改**
   ```bash
   git commit -m "feat: add new feature"
   ```
7. **推送分支**
   ```bash
   git push origin feature/new-feature
   ```
8. **创建 Pull Request**

### 代码规范

- **命名**: 遵循 Go 官方命名约定
- **注释**: 为导出函数添加详细注释
- **测试**: 为新功能添加相应测试
- **文档**: 更新相关文档

### 提交信息规范

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

类型：
- `feat`: 新功能
- `fix`: 修复
- `docs`: 文档
- `style`: 格式
- `refactor`: 重构
- `test`: 测试
- `chore`: 构建/工具

---

## 📄 许可证

[请添加许可证信息]

---

## 🆘 支持

如有问题或建议，请：

1. **提交 Issue**: 在项目仓库中创建 Issue
2. **查看文档**: 参考 [Quality-Gate-Guide.md](Quality-Gate-Guide.md)
3. **联系维护团队**: 通过邮件或其他方式联系

---

## 🏆 致谢

感谢以下开源项目：

- [TiDB Parser](https://github.com/pingcap/tidb) - SQL 解析器
- [Testify](https://github.com/stretchr/testify) - 测试框架
- [Golangci-lint](https://github.com/golangci/golangci-lint) - 代码检查工具

---

## 📈 版本历史

### v2.0 (2026-02-03)
- ✅ **零 lint 问题**: 企业级代码质量标准
- ✅ **gci 集成**: 完美解决导入分组问题
- ✅ **企业级 CI/CD**: 完整的多阶段流水线
- ✅ **复杂度优化**: 高复杂度函数拆分完成
- ✅ **常量集中化**: 统一管理，消除重复
- ✅ **文档合并**: 统一质量门禁指南

### v1.0 (2026-01-XX)
- 🎉 初始版本发布
- 📝 基础功能实现
- 🧪 测试框架搭建
- 📚 文档完善

---

**YB Migration - 让 MySQL 到 YB 的迁移更简单、更可靠！** 🚀

**项目状态**: 生产就绪，企业级标准 ✅
