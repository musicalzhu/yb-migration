package checker

import (
	"fmt"
	"strings"

	"github.com/example/ybMigration/internal/config"
	"github.com/example/ybMigration/internal/model"
	"github.com/pingcap/tidb/pkg/parser/ast"
)

// SyntaxChecker 语法检查器
// 检查 SQL 语法兼容性问题，支持从配置文件加载规则
type SyntaxChecker struct {
	*RuleChecker
}

// NewSyntaxChecker 创建新的 SyntaxChecker 实例
// 返回:
//   - *SyntaxChecker: 初始化后的语法检查器实例
//   - error: 错误信息
func NewSyntaxChecker(cfg *config.Config) (*SyntaxChecker, error) {
	ruleChecker, err := newRuleChecker("SyntaxChecker", "syntax", cfg)
	if err != nil {
		return nil, fmt.Errorf("创建语法检查器失败: %w", err)
	}
	return &SyntaxChecker{
		RuleChecker: ruleChecker,
	}, nil
}

// Name 返回检查器名称
func (s *SyntaxChecker) Name() string { return "SyntaxChecker" }

// Inspect 实现 Checker 接口，处理 AST 节点
// 检查语法兼容性问题，如 AUTO_INCREMENT 等
func (s *SyntaxChecker) Inspect(n ast.Node) (w ast.Node, skipChildren bool) {
	switch node := n.(type) {
	case *ast.CreateTableStmt:
		// 检查并转换表级别的语法问题
		return s.checkCreateTableSyntax(node)

	case *ast.TableName:
		// 检查并转换表名中的反引号
		return s.checkTableNameQuotes(node)

	case *ast.ColumnName:
		// 检查并转换列名中的反引号
		return s.checkColumnNameQuotes(node)

	case *ast.LockTablesStmt:
		// 检查 LOCK TABLES 语法
		return s.checkLockTablesSyntax(node)

	case *ast.UnlockTablesStmt:
		// 检查 UNLOCK TABLES 语法
		return s.checkUnlockTablesSyntax(node)
	}
	return n, false
}

// checkCreateTableSyntax 检查 CREATE TABLE 语句中的语法问题并执行转换
// 参数:
//   - node: CREATE TABLE语句节点
//
// 返回值:
//   - ast.Node: 转换后的节点
//   - bool: 是否有转换发生
func (s *SyntaxChecker) checkCreateTableSyntax(node *ast.CreateTableStmt) (ast.Node, bool) {
	rules := s.GetRules()
	hasTransform := false

	// 检查列定义中的 AUTO_INCREMENT
	for _, col := range node.Cols {
		if col.Options != nil {
			for _, opt := range col.Options {
				if opt.Tp == ast.ColumnOptionAutoIncrement {
					rule, hasRule := rules["AUTO_INCREMENT"]
					if hasRule {
						// 生成兼容性问题
						s.AddIssue(model.Issue{
							Checker: "SyntaxChecker",
							Message: fmt.Sprintf("语法 %s: %s (建议: %s)", "AUTO_INCREMENT", rule.Description, rule.Then.Target),
							AutoFix: model.AutoFix{
								Available: true,
								Action:    rule.Then.Action,
								Code:      fmt.Sprintf("%s -> %s", "AUTO_INCREMENT", rule.Then.Target),
							},
						})

						// 执行AST转换
						transformedNode := s.ApplyTransformation(col, rule)
						if transformedNode != col {
							*col = *transformedNode.(*ast.ColumnDef)
							hasTransform = true
						}
					}
				}
			}
		}
	}

	// 检查表引擎选项 (MySQL 特有)
	for _, option := range node.Options {
		if option.Tp == ast.TableOptionEngine {
			rule, hasRule := rules["ENGINE"]
			if hasRule {
				// 生成兼容性问题
				s.AddIssue(model.Issue{
					Checker: "SyntaxChecker",
					Message: fmt.Sprintf("语法 %s: %s (建议: %s)", "ENGINE", rule.Description, rule.Then.Target),
					AutoFix: model.AutoFix{
						Available: true,
						Action:    rule.Then.Action,
						Code:      fmt.Sprintf("%s -> %s", "ENGINE", rule.Then.Target),
					},
				})

				// 执行AST转换
				transformedNode := s.ApplyTransformation(option, rule)
				if transformedNode != option {
					*option = *transformedNode.(*ast.TableOption)
					hasTransform = true
				}
			}
		}
	}

	return node, hasTransform
}

// checkTableNameQuotes 检查表名中的反引号并执行转换
// 参数:
//   - node: 表名节点
//
// 返回值:
//   - ast.Node: 转换后的节点
//   - bool: 是否有转换发生
func (s *SyntaxChecker) checkTableNameQuotes(node *ast.TableName) (ast.Node, bool) {
	rules := s.GetRules()

	originalName := node.Name.String()
	if strings.Contains(originalName, "`") {
		// 调试：打印所有加载的规则
		fmt.Printf("DEBUG: Loaded rules: %v\n", rules)

		rule, hasRule := rules["`"] // 使用 pattern 作为 key
		if hasRule {
			fmt.Printf("DEBUG: Found backtick rule: %v\n", rule)
			// 生成兼容性问题
			s.AddIssue(model.Issue{
				Checker: "SyntaxChecker",
				Message: fmt.Sprintf("语法 %s: %s (建议: %s)", "反引号", rule.Description, rule.Then.Target),
				AutoFix: model.AutoFix{
					Available: true,
					Action:    rule.Then.Action,
					Code:      fmt.Sprintf("`%s` -> \"%s\"", originalName, strings.ReplaceAll(originalName, "`", "\"")),
				},
			})

			// 执行AST转换
			transformedNode := s.ApplyTransformation(node, rule)
			return transformedNode, transformedNode != node
		} else {
			fmt.Printf("DEBUG: No backtick rule found in rules\n")
		}
	}

	return node, false
}

// checkLockTablesSyntax 检查 LOCK TABLES 语法
// 参数:
//   - node: LOCK TABLES语句节点
//
// 返回值:
//   - ast.Node: 转换后的节点
//   - bool: 是否有转换发生
func (s *SyntaxChecker) checkLockTablesSyntax(node *ast.LockTablesStmt) (ast.Node, bool) {
	issue := model.Issue{
		Checker: "SyntaxChecker",
		Message: "MySQL LOCK TABLES 语法不兼容: LOCK TABLES 是 MySQL 特有的表锁定语法，目标数据库可能使用不同的锁定机制或不支持此语法",
	}
	s.AddIssue(issue)
	return node, false
}

// checkUnlockTablesSyntax 检查 UNLOCK TABLES 语法
// 参数:
//   - node: UNLOCK TABLES语句节点
//
// 返回值:
//   - ast.Node: 转换后的节点
//   - bool: 是否有转换发生
func (s *SyntaxChecker) checkUnlockTablesSyntax(node *ast.UnlockTablesStmt) (ast.Node, bool) {
	issue := model.Issue{
		Checker: "SyntaxChecker",
		Message: "MySQL UNLOCK TABLES 语法不兼容: UNLOCK TABLES 是 MySQL 特有的表解锁语法，目标数据库可能使用不同的锁定机制或不支持此语法",
	}
	s.AddIssue(issue)
	return node, false
}

// checkColumnNameQuotes 检查列名中的反引号并执行转换
// 参数:
//   - node: 列名节点
//
// 返回值:
//   - ast.Node: 转换后的节点
//   - bool: 是否有转换发生
func (s *SyntaxChecker) checkColumnNameQuotes(node *ast.ColumnName) (ast.Node, bool) {
	rules := s.GetRules()

	originalName := node.Name.String()
	if strings.Contains(originalName, "`") {
		rule, hasRule := rules["`"] // 使用 pattern 作为 key
		if hasRule {
			// 生成兼容性问题
			s.AddIssue(model.Issue{
				Checker: "SyntaxChecker",
				Message: fmt.Sprintf("语法 %s: %s (建议: %s)", "反引号", rule.Description, rule.Then.Target),
				AutoFix: model.AutoFix{
					Available: true,
					Action:    rule.Then.Action,
					Code:      fmt.Sprintf("`%s` -> \"%s\"", originalName, strings.ReplaceAll(originalName, "`", "\"")),
				},
			})

			// 执行AST转换
			transformedNode := s.ApplyTransformation(node, rule)
			return transformedNode, transformedNode != node
		}
	}

	return node, false
}
