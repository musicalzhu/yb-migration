# 测试文档

本文档描述了 YB Migration 项目的测试策略、测试用例和测试执行指南。

## 测试架构

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
│   ├── analyzer_test.go           # 分析器单元测试
│   
├── checker/
│   └── checker_test.go            # 检查器单元测试与基准测试
├── config/
│   └── config_test.go             # 配置模块测试
├── input-parser/
│   └── general_log_parser_test.go      # 输入解析器测试
├── report/
│   
├── sql-parser/
│   └── sql_parser_test.go        # SQL 解析器测试
└── testutils/
    └── testutils.go              # 测试工具函数
```

## 运行测试

### 本地测试

```bash
# 运行所有测试
go test -v ./...

# 运行特定模块测试
go test -v ./internal/analyzer
go test -v ./internal/checker

# 运行带覆盖率的测试
go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# 生成覆盖率报告
go tool cover -html=coverage.txt -o coverage.html
```

### 集成测试

```bash
# 运行集成测试
go test -v ./cmd

# 运行特定集成测试
go test -v ./cmd -run TestMain_Integration_ValidSQLFile
```

### 性能测试

```bash
# 运行所有性能测试
go test -bench=. -benchmem ./...

# 运行特定模块性能测试
go test -bench=. -benchmem ./internal/analyzer

# 运行性能测试并生成报告
go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof ./...
```

## 测试用例

### 分析器测试

#### 单元测试

文件：`internal/analyzer/analyzer_test.go`

测试覆盖：
- SQL 解析功能
- 错误处理
- 输入验证
- 边界条件

```go
func TestSQLAnalyzer_AnalyzeSQL(t *testing.T) {
    // 测试基本 SQL 分析功能
}

func TestSQLAnalyzer_AnalyzeSQL_EmptyInput(t *testing.T) {
    // 测试空输入处理
}

func TestSQLAnalyzer_AnalyzeSQL_InvalidSQL(t *testing.T) {
    // 测试无效 SQL 处理
}
```

### 检查器测试

文件：`internal/checker/checker_test.go`

测试覆盖：
- 检查器基础行为
- 多检查器组合
- 部分基准测试

```go
func TestDatatypeChecker_Check(t *testing.T) {
    // 测试数据类型检查
}

func TestDatatypeChecker_GetSuggestion(t *testing.T) {
    // 测试建议生成
}
```

### 集成测试

文件：`cmd/main_integration_test.go`

测试覆盖：
- 完整的分析流程
- 配置文件加载
- 报告生成
- 错误处理

```go
func TestMain_Integration_ValidSQLFile(t *testing.T) {
    // 测试完整流程：配置加载 → SQL 解析 → 报告生成
}

func TestMain_Integration_LogFile(t *testing.T) {
    // 测试日志文件分析
}

func TestMain_Integration_Directory(t *testing.T) {
    // 测试目录批量分析
}
```

## 测试数据

### 测试文件

项目包含以下测试数据文件：

- `testdata/mysql_queries.sql`：示例 SQL 语句
- `testdata/general_log_example.log`：MySQL General Log 示例

### 测试用例数据结构

```go
type TestCase struct {
    Name        string
    Input       string
    Expected    []model.Issue
    Description string
}
```

## CI/CD 测试

### GitLab CI 配置

文件：`.gitlab-ci.yml`

测试阶段：
1. **单元测试**：`go test -v -race -coverprofile=coverage.txt`
2. **集成测试**：`go test -v -race ./cmd`
3. **代码检查**：`golangci-lint run`
4. **性能测试**：`go test -bench=. -benchmem`

### 测试覆盖率

## 测试最佳实践

### API 测试指南

基于简化的 API 设计，推荐以下测试模式：

#### 分析器创建测试

```go
func TestSQLAnalyzer_Creation(t *testing.T) {
    // 测试字符串分析器创建
    factory, err := NewAnalyzerFactory("")
    require.NoError(t, err)
    
    checkers, err := factory.CreateCheckers("datatype", "function")
    require.NoError(t, err)
    
    sqlParser := sqlparser.NewSQLParser()
    
    analyzer, err := NewSQLAnalyzer(
        inputparser.NewStringParser(), 
        sqlParser, 
        checkers,
    )
    require.NoError(t, err)
    assert.NotNil(t, analyzer)
}

func TestSQLAnalyzer_FileAnalyzerCreation(t *testing.T) {
    // 测试文件分析器创建
    factory, err := NewAnalyzerFactory("")
    require.NoError(t, err)
    
    checkers, err := factory.CreateCheckers("datatype")
    require.NoError(t, err)
    
    sqlParser := sqlparser.NewSQLParser()
    
    analyzer, err := NewSQLAnalyzer(
        inputparser.NewSQLFileParser(), 
        sqlParser, 
        checkers,
    )
    require.NoError(t, err)
    assert.NotNil(t, analyzer)
}
```

#### 错误处理测试

```go
func TestSQLAnalyzer_ErrorHandling(t *testing.T) {
    factory, err := NewAnalyzerFactory("")
    require.NoError(t, err)
    
    checkers, err := factory.CreateCheckers("datatype")
    require.NoError(t, err)
    
    sqlParser := sqlparser.NewSQLParser()
    analyzer, err := NewSQLAnalyzer(
        inputparser.NewStringParser(), 
        sqlParser, 
        checkers,
    )
    require.NoError(t, err)
    
    t.Run("empty_sql_error", func(t *testing.T) {
        result, err := analyzer.AnalyzeSQL("", "test")
        require.Error(t, err)
        
        // 使用 errors.As 检查错误类型
        var analysisErr *model.AnalysisError
        require.True(t, errors.As(err, &analysisErr))
        assert.Equal(t, model.ErrorTypeNoSQL, analysisErr.Type)
        assert.Contains(t, analysisErr.Message, "未找到有效的 SQL 语句")
    })
    
    t.Run("invalid_sql_error", func(t *testing.T) {
        result, err := analyzer.AnalyzeSQL("INVALID SQL", "test")
        require.Error(t, err)
        
        var analysisErr *model.AnalysisError
        require.True(t, errors.As(err, &analysisErr))
        assert.Equal(t, model.ErrorTypeParse, analysisErr.Type)
        assert.Contains(t, analysisErr.Message, "SQL 解析失败")
    })
}

func TestSQLAnalyzer_ErrorTypeChecking(t *testing.T) {
    // 测试 errors.Is 使用
    factory, err := NewAnalyzerFactory("")
    require.NoError(t, err)
    
    checkers, err := factory.CreateCheckers("datatype")
    require.NoError(t, err)
    
    sqlParser := sqlparser.NewSQLParser()
    analyzer, err := NewSQLAnalyzer(
        inputparser.NewStringParser(), 
        sqlParser, 
        checkers,
    )
    require.NoError(t, err)
    
    result, err := analyzer.AnalyzeSQL("", "test")
    require.Error(t, err)
    
    // 使用 errors.Is 检查预定义错误
    assert.True(t, errors.Is(err, model.ErrNoSQL))
    assert.False(t, errors.Is(err, model.ErrParse))
}
```

#### AnalyzeInput 集成测试

```go
func TestAnalyzeInput_StringInput(t *testing.T) {
    sqlParser := sqlparser.NewSQLParser()
    factory, err := NewAnalyzerFactory("")
    require.NoError(t, err)
    
    checkers, err := factory.CreateCheckers("datatype")
    require.NoError(t, err)
    
    t.Run("valid_sql_string", func(t *testing.T) {
        sql := "CREATE TABLE test (id INT, name VARCHAR(255))"
        result, err := AnalyzeInput(sql, sqlParser, checkers)
        require.NoError(t, err)
        
        assert.Equal(t, sql, result.SQL)
        assert.Equal(t, "input_string", result.Source)
        assert.NotEmpty(t, result.Issues) // 应该检测到 TINYINT 等问题
    })
    
    t.Run("unsupported_type", func(t *testing.T) {
        result, err := AnalyzeInput(123, sqlParser, checkers)
        require.Error(t, err)
        
        assert.Equal(t, "unknown", result.Source)
        assert.Contains(t, err.Error(), "不支持的输入类型")
    })
}
```

#### 性能测试模板

当前仓库已包含部分基准测试（例如 `internal/checker/checker_test.go` 中的 `Benchmark...`）。如需新增 analyzer 相关基准测试，建议在 `internal/analyzer/analyzer_test.go` 中补充对应的 `Benchmark...` 函数。

### 编写测试用例

1. **命名规范**：使用 `Test[FunctionName]_[Scenario]` 格式
2. **测试结构**：使用 AAA 模式（Arrange, Act, Assert）
3. **断言使用**：使用 testify 库的断言方法
4. **错误处理**：测试错误路径和边界条件

### 测试数据管理

1. **测试文件**：放在 `testdata/` 目录下
2. **Mock 数据**：使用测试工具函数生成
3. **清理工作**：在 `teardown` 中清理测试产生的文件

### 性能测试指南

1. **基准测试**：使用 `Benchmark` 前缀
2. **内存分析**：使用 `-memprofile` 参数
3. **CPU 分析**：使用 `-cpuprofile` 参数
4. **结果对比**：使用 `benchcmp` 工具对比性能变化

## 调试测试

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

1. **测试超时**：增加 `-timeout` 参数
2. **并发测试**：使用 `-race` 参数检测竞态条件
3. **内存泄漏**：使用 `-memprofile` 分析内存使用

## 添加新测试

### 新功能测试

当添加新功能时，应该：

1. 编写单元测试覆盖核心逻辑
2. 编写集成测试验证完整流程
3. 添加性能测试确保不影响整体性能
4. 更新测试文档

### 测试模板

```go
func TestNewFeature_Functionality(t *testing.T) {
    // Arrange
    input := "test input"
    expected := "expected output"
    
    // Act
    result, err := NewFunction(input)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

## 测试报告

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

## 测试环境

### 本地环境

- Go 版本：1.25.1+
- 操作系统：Windows/Linux/macOS
- 依赖：通过 `go mod` 管理

### CI 环境

- Docker 镜像：`golang:latest`
- 代理设置：`GOPROXY=https://goproxy.cn,direct`
- 缓存策略：Go modules 和 build cache

## 故障排除

### 常见测试失败

1. **依赖问题**：运行 `go mod tidy`
2. **权限问题**：检查测试文件权限
3. **路径问题**：使用相对路径或绝对路径
4. **并发问题**：使用 `-race` 参数检测

### 性能问题

1. **内存使用**：使用 `pprof` 分析
2. **CPU 使用**：检查算法复杂度
3. **I/O 瓶颈**：优化文件读写操作

通过遵循这些测试指南，可以确保 YB Migration 项目的质量和可靠性。