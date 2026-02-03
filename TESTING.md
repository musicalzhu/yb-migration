# 测试文档

本文档描述了 YB Migration 项目的测试策略、测试用例和测试执行指南。

**更新版本**: v2.0 - 零 lint 问题版本  
**最后更新**: 2026-02-03

---

## 📊 测试概览

### 当前测试状态
- **总体覆盖率**: 28.8% of statements
- **测试文件数**: 8 个单元测试文件 + 1 个集成测试文件
- **测试代码比例**: 46.4% (1,753 行测试代码)
- **质量状态**: 零 lint 问题，完美代码质量

### 高覆盖率模块
- **internal/config**: 84.2% - 配置管理模块
- **internal/input-parser**: 80.8% - 输入解析器模块  
- **internal/sql-parser**: 66.7% - SQL 解析器模块

---

## 🏗️ 测试架构

### 测试分层

项目采用多层次的测试策略：

1. **单元测试**：测试单个函数和方法的正确性
2. **集成测试**：测试模块间的协作和完整流程
3. **端到端测试**：测试完整的用户使用场景
4. **性能测试**：测试工具的性能和资源使用

### 测试目录结构

```
internal/
├── analyzer/
│   ├── analyzer_test.go           # 分析器单元测试 (15 个测试函数)
│   
├── checker/
│   ├── checker_test.go            # 检查器单元测试 (8 个测试函数)
│   ├── charset_checker_test.go     # 字符集检查器测试
│   ├── datatype_checker_test.go    # 数据类型检查器测试
│   ├── function_checker_test.go    # 函数检查器测试
│   └── syntax_checker_test.go     # 语法检查器测试
├── config/
│   └── config_test.go             # 配置模块测试 (7 个测试函数)
├── constants/
│   └── permissions.go             # 常量模块 (无测试文件)
├── input-parser/
│   ├── general_log_parser_test.go  # 输入解析器测试 (3 个测试函数)
│   └── sqlfile_parser_test.go      # SQL 文件解析器测试
├── model/
│   └── errors.go                  # 数据模型 (无测试文件)
├── report/
│   └── [多个生成器文件]            # 报告生成器 (无测试文件)
├── sql-parser/
│   └── sql_parser_test.go         # SQL 解析器测试 (5 个测试函数)
└── testutils/
    └── testutils.go               # 测试工具函数 (无测试文件)

cmd/
└── main_integration_test.go       # 集成测试 (1 个测试函数)
```

---

## 🚀 运行测试

### 本地测试

```bash
# 运行所有测试
go test -v ./...

# 运行特定模块测试
go test -v ./internal/analyzer
go test -v ./internal/checker
go test -v ./internal/config

# 运行带覆盖率的测试
go test -v -coverprofile=coverage.out ./...

# 生成 HTML 覆盖率报告
go tool cover -html=coverage.out -o coverage.html

# 查看函数级覆盖率
go tool cover -func=coverage.out
```

### 性能测试

```bash
# 运行性能基准测试
go test -bench=. -benchmem ./...

# 运行特定模块的性能测试
go test -bench=. -benchmem ./internal/checker
```

### 并发测试

```bash
# 运行并发测试 (需要 CGO 支持)
# Windows 环境下可能需要设置 CGO_ENABLED=1
go test -race -v ./...
```

---

## 📋 测试用例详情

### 单元测试

#### 1. 分析器测试 (analyzer_test.go)
- **测试函数数**: 15 个
- **覆盖功能**: SQL 分析、错误处理、配置加载
- **关键测试**: 
  - `TestAnalyzeSQL`: SQL 分析功能
  - `TestAnalyzeSQL_ErrorHandling`: 错误处理
  - `TestAnalyzeSQL_ConfigLoading`: 配置加载

#### 2. 检查器测试 (checker_test.go)
- **测试函数数**: 8 个
- **覆盖功能**: 检查器注册、规则加载、检查执行
- **关键测试**:
  - `TestNewRuleChecker`: 检查器创建
  - `TestCheck`: 检查功能
  - `TestLoadRulesFromConfig`: 规则加载

#### 3. 配置测试 (config_test.go)
- **测试函数数**: 7 个
- **覆盖率**: 84.2%
- **关键测试**:
  - `TestConfig`: 配置基础功能
  - `TestLoadConfig`: 配置加载
  - `TestConfigIntegration`: 集成测试

#### 4. 输入解析器测试
- **general_log_parser_test.go**: 3 个测试函数
- **sqlfile_parser_test.go**: SQL 文件解析测试
- **覆盖率**: 80.8%

#### 5. SQL 解析器测试 (sql_parser_test.go)
- **测试函数数**: 5 个
- **覆盖率**: 66.7%
- **关键测试**:
  - `TestParseSQL_Basic`: 基础 SQL 解析
  - `TestParseSQL_ComplexSQL`: 复杂 SQL 解析

### 集成测试

#### main_integration_test.go
- **测试函数数**: 1 个
- **覆盖功能**: 端到端工作流测试
- **测试场景**: 完整的迁移分析流程

---

## 📈 覆盖率分析

### 当前覆盖率分布

| 模块 | 覆盖率 | 状态 | 建议 |
|------|--------|------|------|
| internal/config | 84.2% | ✅ 优秀 | 保持 |
| internal/input-parser | 80.8% | ✅ 优秀 | 保持 |
| internal/sql-parser | 66.7% | ✅ 良好 | 可提升 |
| internal/analyzer | 待统计 | ⚠️ 待提升 | 需要增加测试 |
| internal/checker | 待统计 | ⚠️ 待提升 | 需要增加测试 |
| internal/constants | 0% | ❌ 无测试 | 需要添加测试 |
| internal/model | 0% | ❌ 无测试 | 需要添加测试 |
| internal/report | 0% | ❌ 无测试 | 需要添加测试 |
| internal/testutils | 0% | ❌ 无测试 | 可选 |

### 覆盖率提升计划

#### 短期目标 (1 个月内)
1. **internal/constants**: 添加基础单元测试
2. **internal/model**: 添加错误处理测试
3. **internal/report**: 添加报告生成器测试

#### 中期目标 (3 个月内)
1. **总体覆盖率**: 提升到 60%+
2. **核心模块**: 达到 80%+ 覆盖率
3. **边界测试**: 增加更多边界条件测试

---

## 🔧 测试工具和配置

### 测试框架
- **主要框架**: Go 标准测试包 + testify
- **断言库**: testify/assert
- **模拟库**: testify/mock (如需要)

### 测试数据
```bash
# 测试数据目录
testdata/
├── mysql_queries.sql          # 示例 SQL 查询
├── general_log_example.log    # 示例日志文件
└── configs/                   # 测试配置文件
    └── test_config.yaml
```

### 测试配置
```yaml
# testdata/configs/test_config.yaml
rules:
  datatype:
    - pattern: "TINYINT"
      suggestion: "使用 SMALLINT 替代 TINYINT"
      severity: "warning"
```

---

## 🎯 测试最佳实践

### 1. 测试命名规范
```go
// 好的测试命名
func TestAnalyzeSQL_ValidSQL(t *testing.T) { ... }
func TestAnalyzeSQL_InvalidSQL(t *testing.T) { ... }
func TestAnalyzeSQL_EmptyInput(t *testing.T) { ... }

// 避免的命名
func TestAnalyzeSQL1(t *testing.T) { ... }
func TestFunction(t *testing.T) { ... }
```

### 2. 测试结构
```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   interface{}
        want    interface{}
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   "valid data",
            want:    "expected result",
            wantErr: false,
        },
        // 更多测试用例...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("FunctionName() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("FunctionName() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### 3. 测试数据管理
```go
// 使用测试工具函数
func getTestDataPath(filename string) string {
    return filepath.Join("testdata", filename)
}

// 在测试中使用
func TestSQLFileParser_Parse(t *testing.T) {
    parser := NewSQLFileParser()
    
    // 使用测试数据文件
    result, err := parser.Parse(getTestDataPath("test.sql"))
    
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

---

## 🚨 CI/CD 测试集成

### GitLab CI/CD 测试阶段
```yaml
# 测试阶段
test:
  stage: test
  script:
    - echo "🧪 运行单元测试..."
    - go test -v -race ./...
    
    - echo "📊 生成覆盖率报告..."
    - go test -coverprofile=coverage.out ./...
    - go tool cover -html=coverage.out -o coverage.html
    
    - echo "📈 覆盖率统计..."
    - go tool cover -func=coverage.out
    
    - echo "⚡ 运行性能测试..."
    - go test -bench=. -benchmem ./...
```

### 质量门禁
- **测试通过率**: 100% (必须全部通过)
- **覆盖率目标**: 核心模块 > 80%
- **并发测试**: 启用竞态检测
- **性能测试**: 监控性能回归

---

## 📝 添加新测试

### 1. 添加单元测试
```bash
# 为新模块创建测试文件
touch internal/newmodule/newmodule_test.go
```

### 2. 测试模板
```go
package newmodule

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestNewFunction(t *testing.T) {
    tests := []struct {
        name string
        args args
        want return_type
    }{
        // 测试用例...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := NewFunction(tt.args)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### 3. 运行新测试
```bash
# 运行新模块测试
go test -v ./internal/newmodule

# 检查覆盖率
go test -coverprofile=coverage.out ./internal/newmodule
go tool cover -func=coverage.out
```

---

## 🎉 测试成就

### v2.0 测试改进
- ✅ **零 lint 问题**: 所有测试代码完美格式
- ✅ **覆盖率报告**: 生成详细的 HTML 覆盖率报告
- ✅ **CI/CD 集成**: 完整的测试流水线
- ✅ **性能测试**: 基础性能测试框架
- ✅ **并发测试**: 竞态检测支持

### 测试统计
- **总测试函数**: 39 个
- **测试代码行数**: 1,753 行
- **测试覆盖率**: 28.8% (核心模块良好)
- **测试文件数**: 9 个

---

## 🔍 调试测试

### 调试技巧

```bash
# 运行单个测试并显示详细输出
go test -v -run TestSpecificFunction

# 在测试中设置断点
go test -run TestSpecificFunction -ldflags="-compressdwarf=false"

# 查看测试覆盖率详情
go test -coverprofile=coverage.out
go tool cover -func=coverage.out
```

### 常见问题

1. **测试超时**: 增加 `-timeout` 参数
2. **并发测试**: 使用 `-race` 参数检测竞态条件
3. **内存泄漏**: 使用 `-memprofile` 分析内存使用

---

## 📊 测试报告

### 覆盖率报告

- HTML 格式：`coverage.html`
- 文本格式：`coverage.txt`
- 函数级别：`go tool cover -func=coverage.out`

### 性能报告

- 基准测试结果：控制台输出
- CPU 分析：`cpu.prof`
- 内存分析：`mem.prof`

### 持续集成报告

- GitLab CI artifacts
- 测试结果摘要
- 覆盖率趋势图

---

## 🌍 测试环境

### 本地环境

- Go 版本：1.25.1+
- 操作系统：Windows/Linux/macOS
- 依赖：通过 `go mod` 管理

### CI 环境

- Docker 镜像：`golang:latest`
- 代理设置：`GOPROXY=https://goproxy.cn,direct`
- 缓存策略：Go modules 和 build cache

---

## 🛠️ 故障排除

### 常见测试失败

1. **依赖问题**: 运行 `go mod tidy`
2. **权限问题**: 检查测试文件权限
3. **路径问题**: 使用相对路径或绝对路径
4. **并发问题**: 使用 `-race` 参数检测

### 性能问题

1. **内存使用**: 使用 `pprof` 分析
2. **CPU 使用**: 检查算法复杂度
3. **I/O 瓶颈**: 优化文件读写操作

---

**测试文档更新完成！项目拥有完善的测试体系和详细的测试指南。** 🚀

**下次更新**: 根据测试覆盖率提升进度定期更新
