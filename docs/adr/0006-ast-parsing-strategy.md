# 0006. AST 解析策略选择

## 状态

接受

## 背景

YB Migration 需要解析 SQL 语句并进行多种检查和转换。在选择解析策略时，我们需要考虑以下因素：

1. **解析准确性**: 需要准确理解 SQL 语句的结构和语义
2. **性能要求**: 需要高效处理大量 SQL 文件
3. **扩展性**: 需要支持多种 SQL 方言和数据库特性
4. **维护性**: 解析器应该易于维护和扩展
5. **生态系统**: 需要良好的工具链和社区支持

## 决策

我们选择基于 TiDB Parser 的 AST 解析策略，原因如下：

1. **成熟稳定**: TiDB Parser 经过大量生产环境验证
2. **功能完整**: 支持 MySQL 8.0 的大部分语法特性
3. **性能优秀**: 基于 Go 实现，性能优异
4. **社区活跃**: 持续维护和更新
5. **易于集成**: Go 生态系统中集成简单

## 后果

### 正面影响

1. **解析准确**: 基于 AST 的解析比正则表达式更准确
2. **语义理解**: 可以理解 SQL 语句的语义结构
3. **扩展性强**: 易于添加新的语法支持
4. **性能优秀**: 解析速度快，内存占用合理
5. **维护简单**: 使用成熟的第三方库，减少维护负担

### 负面影响

1. **依赖外部**: 依赖 TiDB Parser 的更新和维护
2. **学习成本**: 需要理解 AST 结构和 API
3. **版本兼容**: 需要关注 TiDB Parser 的版本更新

## 实施细节

### AST 解析器设计

```go
// SQLParser 接口定义
type SQLParser interface {
    Parse(sql string) ([]SQLStatement, error)
    ParseFile(filePath string) ([]SQLStatement, error)
}

// TiDBParser 实现
type TiDBParser struct {
    parser *parser.Parser
}

func NewSQLParser() SQLParser {
    return &TiDBParser{
        parser: parser.New(),
    }
}

func (p *TiDBParser) Parse(sql string) ([]SQLStatement, error) {
    stmtNodes, _, err := p.parser.Parse(sql, "", "")
    if err != nil {
        return nil, fmt.Errorf("SQL 解析失败: %w", err)
    }
    
    var statements []SQLStatement
    for _, stmtNode := range stmtNodes {
        statement := p.convertASTToStatement(stmtNode)
        statements = append(statements, statement)
    }
    
    return statements, nil
}

func (p *TiDBParser) convertASTToStatement(stmtNode ast.StmtNode) SQLStatement {
    // 将 AST 节点转换为内部数据结构
    switch stmt := stmtNode.(type) {
    case *ast.CreateTableStmt:
        return p.convertCreateTable(stmt)
    case *ast.InsertStmt:
        return p.convertInsert(stmt)
    case *ast.SelectStmt:
        return p.convertSelect(stmt)
    default:
        return SQLStatement{
            Type:    stmt.Text(),
            Content: stmt.Text(),
            AST:     stmt,
        }
    }
}
```

### 数据结构设计

```go
// SQLStatement 表示解析后的 SQL 语句
type SQLStatement struct {
    Type         string                 `json:"type"`
    Content      string                 `json:"content"`
    AST          ast.StmtNode           `json:"-"`
    Columns      []Column               `json:"columns,omitempty"`
    Tables       []Table                `json:"tables,omitempty"`
    Functions    []FunctionCall         `json:"functions,omitempty"`
    Constraints  []Constraint           `json:"constraints,omitempty"`
    LineNumber   int                    `json:"line_number"`
    Column       int                    `json:"column"`
}

// Column 表示列定义
type Column struct {
    Name       string `json:"name"`
    Type       string `json:"type"`
    Nullable   bool   `json:"nullable"`
    Default    string `json:"default,omitempty"`
    LineNumber int    `json:"line_number"`
    Column     int    `json:"column"`
}

// FunctionCall 表示函数调用
type FunctionCall struct {
    Name       string   `json:"name"`
    Arguments  []string `json:"arguments"`
    LineNumber int      `json:"line_number"`
    Column     int      `json:"column"`
}
```

## 替代方案

### 正则表达式解析
- **优点**: 实现简单，无外部依赖
- **缺点**: 解析不准确，难以处理复杂语法，扩展性差

### 自定义解析器
- **优点**: 完全可控，针对性强
- **缺点**: 开发成本高，维护负担重，容易出错

### ANTLR 解析器
- **优点**: 功能强大，支持多种语法
- **缺点**: Go 集成复杂，性能相对较差

## 相关决策

- [0002. 模块化架构设计](0002-modular-architecture.md)
- [0004. 插件化检查器架构](0004-plugin-checkers.md)
- [0007. 一次遍历AST完成所有检查](0007-single-pass-ast-traversal.md)

## 参考资料

- [TiDB Parser 文档](https://github.com/pingcap/tidb/tree/master/pkg/parser)
- [AST 设计模式](https://en.wikipedia.org/wiki/Abstract_syntax_tree)
- [SQL 解析最佳实践](https://github.com/pingcap/parser)

---

*创建日期: 2026-02-03*  
*最后更新: 2026-02-03*  
*负责人: YB Migration Team*
