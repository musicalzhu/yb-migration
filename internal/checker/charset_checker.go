package checker

import (
	"fmt"
	"strings"

	"github.com/pingcap/tidb/pkg/parser/ast"

	"github.com/example/ybMigration/internal/config"
	"github.com/example/ybMigration/internal/model"
)

// CharsetChecker 字符集检查器
// 专门用于检测和报告不兼容的字符集和排序规则用法
// 转换逻辑由统一的AST规则引擎处理
//
// 主要功能:
//   - 检测不兼容的字符集设置
//   - 检测不兼容的排序规则
//   - 提供问题描述
//   - 生成兼容性报告
//   - 与AST规则引擎协同工作
//
// 注意: TableOptionCharacterSet 和 TableOptionCollate 常量已在 checker 包中定义

// CharsetChecker 字符集检查器实现
// 检查SQL字符集和排序规则兼容性问题
// 支持从default.yaml配置文件加载规则，实现MySQL字符集到目标数据库的转换
type CharsetChecker struct {
	*RuleChecker
}

// NewCharsetChecker 创建字符集检查器实例
// 返回:
//   - *CharsetChecker: 初始化后的字符集检查器实例
//   - error: 错误信息
func NewCharsetChecker(cfg *config.Config) (*CharsetChecker, error) {
	ruleChecker, err := newRuleChecker("CharsetChecker", "charset", cfg)
	if err != nil {
		return nil, fmt.Errorf("创建字符集检查器失败: %w", err)
	}
	return &CharsetChecker{
		RuleChecker: ruleChecker,
	}, nil
}

// Name 返回检查器名称
func (c *CharsetChecker) Name() string { return "CharsetChecker" }

// Inspect 实现 Checker 接口，处理 AST 节点
// 检查字符集和排序规则兼容性并执行转换
func (c *CharsetChecker) Inspect(n ast.Node) (w ast.Node, skipChildren bool) {
	switch node := n.(type) {
	case *ast.CreateTableStmt:
		// 检查并转换表级别的字符集设置
		return c.checkTableCharset(node)

	case *ast.ColumnDef:
		// 检查并转换列级别的字符集设置
		return c.checkColumnCharset(node)

	case *ast.AlterTableStmt:
		// 检查并转换ALTER TABLE中的字符集设置
		return c.checkAlterTableCharset(node)
	}
	return n, false
}

// checkTableCharset 检查表级别的字符集设置并执行转换
// 参数:
//   - node: CREATE TABLE语句节点
//
// 返回值:
//   - ast.Node: 转换后的节点
//   - bool: 是否有转换发生
func (c *CharsetChecker) checkTableCharset(node *ast.CreateTableStmt) (ast.Node, bool) {
	if node.Options == nil {
		return node, false
	}

	hasTransform := false
	for _, opt := range node.Options {
		if c.processCharsetOption(opt) {
			hasTransform = true
		}
	}

	return node, hasTransform
}

// checkColumnCharset 检查列级别的字符集设置并执行转换
// 参数:
//   - node: 列定义节点
//
// 返回值:
//   - ast.Node: 转换后的节点
//   - bool: 是否有转换发生
func (c *CharsetChecker) checkColumnCharset(node *ast.ColumnDef) (ast.Node, bool) {
	if node.Tp == nil {
		return node, false
	}

	hasTransform := false
	// 检查FieldType中的字符集信息
	charset := node.Tp.GetCharset()
	collate := node.Tp.GetCollate()

	if charset != "" {
		if transformedNode, transformed := c.checkCharsetRule(node, charset, TableOptionCharacterSet); transformed {
			if colDef, ok := transformedNode.(*ast.ColumnDef); ok {
				*node = *colDef
				hasTransform = true
			}
		}
	}

	if collate != "" {
		if transformedNode, transformed := c.checkCharsetRule(node, collate, TableOptionCollate); transformed {
			if colDef, ok := transformedNode.(*ast.ColumnDef); ok {
				*node = *colDef
				hasTransform = true
			}
		}
	}

	return node, hasTransform
}

// checkAlterTableCharset 检查ALTER TABLE语句中的字符集变更并执行转换
// 参数:
//   - node: ALTER TABLE语句节点
//
// 返回值:
//   - ast.Node: 转换后的节点
//   - bool: 是否有转换发生
func (c *CharsetChecker) checkAlterTableCharset(node *ast.AlterTableStmt) (ast.Node, bool) {
	hasTransform := false

	// 遍历ALTER TABLE中的所有变更项
	for _, spec := range node.Specs {
		if transformed := c.processAlterTableSpec(spec); transformed {
			hasTransform = true
		}
	}

	return node, hasTransform
}

// processAlterTableSpec 处理单个 ALTER TABLE 规范
// 参数:
//   - spec: ALTER TABLE 规范节点
//
// 返回值:
//   - bool: 是否有转换发生
func (c *CharsetChecker) processAlterTableSpec(spec *ast.AlterTableSpec) bool {
	switch spec.Tp {
	case ast.AlterTableAddColumns:
		return c.processAddColumns(spec)
	case ast.AlterTableModifyColumn, ast.AlterTableChangeColumn:
		return c.processModifyColumn(spec)
	case ast.AlterTableOption:
		return c.processTableOptions(spec)
	default:
		return false
	}
}

// processAddColumns 处理添加列操作
// 参数:
//   - spec: ALTER TABLE 规范节点
//
// 返回值:
//   - bool: 是否有转换发生
func (c *CharsetChecker) processAddColumns(spec *ast.AlterTableSpec) bool {
	if spec.NewColumns == nil {
		return false
	}
	return c.processColumns(spec.NewColumns)
}

// processModifyColumn 处理修改列操作
// 参数:
//   - spec: ALTER TABLE 规范节点
//
// 返回值:
//   - bool: 是否有转换发生
func (c *CharsetChecker) processModifyColumn(spec *ast.AlterTableSpec) bool {
	if len(spec.NewColumns) == 0 {
		return false
	}
	return c.processColumns(spec.NewColumns)
}

// processTableOptions 处理表选项操作
// 参数:
//   - spec: ALTER TABLE 规范节点
//
// 返回值:
//   - bool: 是否有转换发生
func (c *CharsetChecker) processTableOptions(spec *ast.AlterTableSpec) bool {
	if spec.Options == nil {
		return false
	}

	hasTransform := false
	for _, opt := range spec.Options {
		if c.processCharsetOption(opt) {
			hasTransform = true
		}
	}
	return hasTransform
}

// isCharsetOption 检查是否为字符集相关选项
// 参数:
//   - opt: 表选项节点
//
// 返回值:
//   - bool: 是否为字符集选项
func (c *CharsetChecker) isCharsetOption(opt *ast.TableOption) bool {
	return opt.Tp == TableOptionCharacterSet || opt.Tp == TableOptionCollate
}

// processCharsetOption 处理单个字符集选项
// 参数:
//   - opt: 表选项节点
//
// 返回值:
//   - bool: 是否有转换发生
func (c *CharsetChecker) processCharsetOption(opt *ast.TableOption) bool {
	if !c.isCharsetOption(opt) {
		return false
	}

	if transformedNode, transformed := c.checkCharsetRule(opt, opt.StrValue, int(opt.Tp)); transformed {
		if tableOption, ok := transformedNode.(*ast.TableOption); ok {
			*opt = *tableOption
			return true
		}
	}
	return false
}

// processColumns 处理列数组中的字符集设置
// 参数:
//   - columns: 列定义数组
//
// 返回值:
//   - bool: 是否有转换发生
func (c *CharsetChecker) processColumns(columns []*ast.ColumnDef) bool {
	hasTransform := false
	for _, col := range columns {
		if transformedNode, transformed := c.checkColumnCharset(col); transformed {
			if colDef, ok := transformedNode.(*ast.ColumnDef); ok {
				*col = *colDef
				hasTransform = true
			}
		}
	}
	return hasTransform
}

// checkCharsetRule 检查字符集规则并执行转换
// 参数:
//   - node: AST节点（TableOption或ColumnDef）
//   - value: 字符集或排序规则值
//   - optType: TableOptionCharacterSet=CHARACTER SET, TableOptionCollate=COLLATE
//
// 返回值:
//   - ast.Node: 转换后的节点
//   - bool: 是否有转换发生
func (c *CharsetChecker) checkCharsetRule(node ast.Node, value string, optType int) (ast.Node, bool) {
	if value == "" {
		return node, false
	}

	// 获取规则
	rules := c.GetRules()

	// 使用大写进行匹配
	upperValue := strings.ToUpper(value)
	rule, hasRule := rules[upperValue]

	if !hasRule {
		return node, false
	}

	// 确定问题类型
	var issueType string
	if optType == TableOptionCharacterSet {
		issueType = "字符集"
	} else {
		issueType = "排序规则"
	}

	// 生成兼容性问题
	c.AddIssue(model.Issue{
		Checker: "CharsetChecker",
		Message: fmt.Sprintf("%s %s: %s (建议: %s)", issueType, value, rule.Description, rule.Then.Target),
		AutoFix: model.AutoFix{
			Available: true,
			Action:    rule.Then.Action,
			Code:      fmt.Sprintf("%s -> %s", value, rule.Then.Target),
		},
	})

	// 执行AST转换
	transformedNode := c.ApplyTransformation(node, rule)
	return transformedNode, true
}
