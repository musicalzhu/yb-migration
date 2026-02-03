# 0004. 插件化检查器架构

## 状态

接受

## 背景

YB Migration 需要支持多种类型的兼容性检查，包括：
1. 函数兼容性检查
2. 数据类型兼容性检查
3. 语法兼容性检查
4. 字符集兼容性检查

随着需求的发展，可能需要添加新的检查类型。为了确保系统的可扩展性和灵活性，我们需要一个可插拔的检查器架构。

## 决策

采用插件化检查器架构，定义标准的检查器接口，支持：
1. 动态加载检查器
2. 配置驱动的检查器启用/禁用
3. 检查器参数化配置
4. 检查器优先级和依赖管理

## 后果

### 正面影响

1. **可扩展性**: 可以轻松添加新的检查器
2. **可配置性**: 通过配置文件控制检查器行为
3. **可测试性**: 每个检查器可以独立测试
4. **可维护性**: 检查器之间解耦，易于维护
5. **灵活性**: 支持第三方检查器开发
6. **性能优化**: 可以选择性启用需要的检查器

### 负面影响

1. **复杂性**: 需要设计插件接口和管理机制
2. **性能开销**: 动态加载可能有性能开销
3. **调试困难**: 插件错误可能难以调试

## 实施细节

### 检查器接口定义

```go
// Checker 接口定义
type Checker interface {
    // 检查单个 SQL 语句
    Check(stmt SQLStatement) []Issue
    
    // 获取检查器名称
    GetName() string
    
    // 获取检查器类别
    GetCategory() string
    
    // 获取检查器描述
    GetDescription() string
    
    // 获取检查器版本
    GetVersion() string
    
    // 初始化检查器
    Initialize(config map[string]interface{}) error
    
    // 验证配置
    ValidateConfig(config map[string]interface{}) error
    
    // 获取默认配置
    GetDefaultConfig() map[string]interface{}
}

// Issue 表示检查发现的问题
type Issue struct {
    Checker    string    `json:"checker"`
    Category   string    `json:"category"`
    Severity   string    `json:"severity"`
    Message    string    `json:"message"`
    LineNumber int       `json:"line_number"`
    Column     int       `json:"column"`
    Suggestion string    `json:"suggestion"`
    RuleID     string    `json:"rule_id"`
}
```

### 检查器注册机制

```go
// 检查器注册表
type CheckerRegistry struct {
    checkers map[string]Checker
    mu       sync.RWMutex
}

var globalRegistry = &CheckerRegistry{
    checkers: make(map[string]Checker),
}

// 注册检查器
func RegisterChecker(name string, checker Checker) {
    globalRegistry.mu.Lock()
    defer globalRegistry.mu.Unlock()
    globalRegistry.checkers[name] = checker
}

// 获取检查器
func GetChecker(name string) (Checker, bool) {
    globalRegistry.mu.RLock()
    defer globalRegistry.mu.RUnlock()
    checker, exists := globalRegistry.checkers[name]
    return checker, exists
}

// 列出所有检查器
func ListCheckers() []string {
    globalRegistry.mu.RLock()
    defer globalRegistry.mu.RUnlock()
    
    names := make([]string, 0, len(globalRegistry.checkers))
    for name := range globalRegistry.checkers {
        names = append(names, name)
    }
    return names
}
```

### 内置检查器实现

```go
// 函数兼容性检查器
type FunctionChecker struct {
    name     string
    config   map[string]interface{}
    database string
}

func NewFunctionChecker() *FunctionChecker {
    return &FunctionChecker{
        name: "function_incompatibility",
    }
}

func (c *FunctionChecker) Check(stmt SQLStatement) []Issue {
    var issues []Issue
    
    // 检查函数调用
    for _, funcCall := range stmt.FunctionCalls {
        if !c.isCompatibleFunction(funcCall.Name) {
            issues = append(issues, Issue{
                Checker:    c.GetName(),
                Category:   c.GetCategory(),
                Severity:   "error",
                Message:    fmt.Sprintf("函数 %s 在目标数据库中不兼容", funcCall.Name),
                LineNumber: funcCall.LineNumber,
                Column:     funcCall.Column,
                Suggestion: c.getSuggestion(funcCall.Name),
                RuleID:     "FUNC_001",
            })
        }
    }
    
    return issues
}

func (c *FunctionChecker) GetName() string {
    return c.name
}

func (c *FunctionChecker) GetCategory() string {
    return "function"
}

func (c *FunctionChecker) GetDescription() string {
    return "检查函数兼容性"
}

func (c *FunctionChecker) Initialize(config map[string]interface{}) error {
    c.config = config
    if db, ok := config["target_database"].(string); ok {
        c.database = db
    }
    return nil
}

// 注册检查器
func init() {
    RegisterChecker("function_incompatibility", NewFunctionChecker())
}
```

### 检查器工厂

```go
// 检查器工厂
type CheckerFactory struct {
    registry *CheckerRegistry
}

func NewCheckerFactory() *CheckerFactory {
    return &CheckerFactory{
        registry: globalRegistry,
    }
}

func (f *CheckerFactory) CreateCheckersFromConfig(config *Config) ([]Checker, error) {
    var checkers []Checker
    
    for _, rule := range config.Rules {
        if !rule.Enabled {
            continue
        }
        
        checker, exists := f.registry.GetChecker(rule.Name)
        if !exists {
            return nil, fmt.Errorf("未找到检查器: %s", rule.Name)
        }
        
        // 初始化检查器
        if err := checker.Initialize(rule.Parameters); err != nil {
            return nil, fmt.Errorf("初始化检查器 %s 失败: %w", rule.Name, err)
        }
        
        checkers = append(checkers, checker)
    }
    
    return checkers, nil
}
```

## 替代方案

### 硬编码检查器
- **优点**: 实现简单，性能好
- **缺点**: 不易扩展，添加新检查器需要修改代码

### 配置文件驱动
- **优点**: 灵活配置
- **缺点**: 实现复杂，难以处理复杂逻辑

### 外部脚本检查器
- **优点**: 极高灵活性
- **缺点**: 性能差，安全性问题

## 相关决策

- [0002. 模块化架构设计](0002-modular-architecture.md)
- [0003. 使用 YAML 配置文件](0003-yaml-config.md)

## 参考资料

- [Go 插件模式](https://dave.cheney.net/2016/08/20/go-plugin)
- [策略模式](https://en.wikipedia.org/wiki/Strategy_pattern)
- [插件架构最佳实践](https://martinfowler.com/articles/plugin-architecture.html)

---

*创建日期: 2026-02-03*  
*最后更新: 2026-02-03*  
*负责人: YB Migration Team*
