// Package sqlparser 提供基于 TiDB 的 SQL 解析器封装，用于将 SQL 文本解析为 AST 节点。
// 空白导入 `test_driver` 用于使 TiDB 解析器加载必要的测试驱动。
package sqlparser

import (
	"fmt"

	pparser "github.com/pingcap/tidb/pkg/parser"
	"github.com/pingcap/tidb/pkg/parser/ast"

	// 空白导入 test_driver 是为了兼容 TiDB 的解析器实现。
	_ "github.com/pingcap/tidb/pkg/parser/test_driver"
)

// SqlParser 定义 SQL 解析器接口
// SqlParser 定义 SQL 解析器接口
// 注意：类型名保持 `SqlParser` 以兼容现有代码调用。
//
//revive:disable:var-naming
type SqlParser interface {
	// ParseSQL 解析 SQL 语句，返回 AST 节点
	ParseSQL(sql string) ([]ast.StmtNode, error)
}

//revive:enable:var-naming

// SQLParser 实现 Parser 接口
type SQLParser struct {
	parser *pparser.Parser
}

// NewSQLParser 创建新的 SQL 解析器实例
func NewSQLParser() *SQLParser {
	p := pparser.New()
	// 启用严格模式，确保SQL语法正确
	p.EnableWindowFunc(true)
	return &SQLParser{
		parser: p,
	}
}

// ParseSQL 解析 SQL 语句，返回 AST 节点
// 这是解析阶段，只负责将 SQL 文本转换为 AST
func (p *SQLParser) ParseSQL(sql string) ([]ast.StmtNode, error) {
	// 使用 TiDB 的解析器解析 SQL
	stmts, warns, err := p.parser.ParseSQL(sql)
	if err != nil {
		return nil, fmt.Errorf("SQL 解析错误: %w", err)
	}

	// 处理警告
	for _, warn := range warns {
		// 这里可以记录警告日志
		_ = warn
	}

	return stmts, nil
}

// parseError 封装解析错误
type parseError struct {
	err error
}

func (e *parseError) Error() string {
	return e.err.Error()
}

func (e *parseError) Unwrap() error {
	return e.err
}

// NewParseError 创建新的解析错误
func NewParseError(err error) error {
	return &parseError{err: err}
}
