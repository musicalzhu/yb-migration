# 0002. 模块化架构设计

## 状态

接受

## 背景

YB Migration 需要处理多种类型的输入（SQL 文件、日志文件、目录），执行多种检查（函数兼容性、数据类型、语法等），并生成多种格式的报告。为了确保代码的可维护性、可扩展性和可测试性，我们需要一个清晰的架构设计。

## 决策

采用模块化架构，将系统划分为以下核心模块：

1. **输入解析模块** (`internal/input-parser`)
2. **SQL 解析模块** (`internal/sql-parser`)
3. **检查器模块** (`internal/checker`)
4. **分析器模块** (`internal/analyzer`)
5. **报告生成模块** (`internal/report`)
6. **配置管理模块** (`internal/config`)
7. **数据模型模块** (`internal/model`)

## 后果

### 正面影响

1. **职责分离**: 每个模块有明确的职责，降低耦合度
2. **可扩展性**: 新增检查器或报告格式只需扩展相应模块
3. **可测试性**: 每个模块可以独立测试
4. **可维护性**: 修改某个功能不会影响其他模块
5. **代码复用**: 模块可以在不同场景下复用
6. **团队协作**: 不同开发者可以并行开发不同模块

### 负面影响

1. **复杂性**: 需要设计模块间的接口
2. **性能开销**: 模块间调用可能有轻微性能开销
3. **学习成本**: 新开发者需要理解架构设计

## 实施细节

### 模块职责

#### 输入解析模块 (`internal/input-parser`)
- 负责解析不同类型的输入文件
- 支持 SQL 文件、MySQL 日志文件、目录扫描
- 提供统一的输入接口

#### SQL 解析模块 (`internal/sql-parser`)
- 使用 TiDB parser 解析 SQL 语句
- 提供语法树和元数据信息
- 支持多种 SQL 方言

#### 检查器模块 (`internal/checker`)
- 实现各种兼容性检查器
- 支持函数、数据类型、语法、字符集检查
- 提供可扩展的检查器接口

#### 分析器模块 (`internal/analyzer`)
- 协调各个模块的工作
- 实现分析流程控制
- 提供分析器工厂模式

#### 报告生成模块 (`internal/report`)
- 生成多种格式的报告
- 支持 JSON、HTML、Markdown、SQL 格式
- 提供报告模板和样式

#### 配置管理模块 (`internal/config`)
- 管理配置文件加载和验证
- 提供默认配置和用户配置合并
- 支持配置热重载

#### 数据模型模块 (`internal/model`)
- 定义核心数据结构
- 提供数据验证和转换
- 统一数据格式

### 接口设计

```go
// 核心接口定义
type Parser interface {
    Parse(input string) ([]SQLStatement, error)
}

type Checker interface {
    Check(stmt SQLStatement) []Issue
    GetName() string
    GetCategory() string
}

type Reporter interface {
    Generate(result *AnalysisResult) ([]byte, error)
    GetFormat() string
}

type Analyzer interface {
    Analyze(input string, checkers []Checker) (*AnalysisResult, error)
}
```

## 替代方案

### 单体架构
- **优点**: 实现简单，性能可能更好
- **缺点**: 难以维护和扩展，测试困难

### 微服务架构
- **优点**: 高度解耦，独立部署
- **缺点**: 过度复杂，不适合 CLI 工具

### 插件架构
- **优点**: 极高的扩展性
- **缺点**: 实现复杂，安全性考虑多

## 相关决策

- [0001. 使用 Go 语言开发 CLI 工具](0001-use-go-for-cli.md)
- [0004. 插件化检查器架构](0004-plugin-checkers.md)

## 参考资料

- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go 项目布局最佳实践](https://github.com/golang-standards/project-layout)
- [模块化设计原则](https://en.wikipedia.org/wiki/Modular_programming)

---

*创建日期: 2026-02-03*  
*最后更新: 2026-02-03*  
*负责人: YB Migration Team*
