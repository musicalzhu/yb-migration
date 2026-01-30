package checker

import (
	"fmt"
	"strings"

	"github.com/example/ybMigration/internal/config"
	"github.com/example/ybMigration/internal/model"
	"github.com/pingcap/tidb/pkg/parser/ast"
)

// FunctionChecker 函数检查器
// 专门用于检测和报告不兼容的 MySQL 函数用法
// 转换逻辑由统一的AST规则引擎处理
//
// 主要功能:
//   - 检测不兼容的函数调用
//   - 提供问题描述
//   - 生成兼容性报告
//   - 与AST规则引擎协同工作

type FunctionChecker struct {
	*RuleChecker
}

// NewFunctionChecker 创建函数检查器实例
// 返回:
//   - *FunctionChecker: 初始化后的函数检查器实例
//   - error: 错误信息
func NewFunctionChecker(cfg *config.Config) (*FunctionChecker, error) {
	ruleChecker, err := NewRuleChecker("FunctionChecker", "function", cfg)
	if err != nil {
		return nil, fmt.Errorf("创建函数检查器失败: %w", err)
	}
	return &FunctionChecker{
		RuleChecker: ruleChecker,
	}, nil
}

// Name 返回检查器名称
func (f *FunctionChecker) Name() string { return "FunctionChecker" }

// Inspect 检查 AST 节点中的函数调用
// Inspect 实现 Checker 接口
// 检查函数调用节点的兼容性并执行转换
//
// 参数:
//   - node: 当前检查的 AST 节点
//
// 返回值:
//   - ast.Node: 转换后的节点，如果不转换则返回原节点
//   - bool: 是否跳过检查子节点
//
// 检查逻辑:
// 1. 处理普通函数调用 (FuncCallExpr)
// 2. 处理聚合函数调用 (AggregateFuncExpr)
// 3. 处理窗口函数调用 (WindowFuncExpr)
// 4. 检查函数是否在规则中
// 5. 如果找到匹配规则，生成兼容性问题报告并执行转换
func (f *FunctionChecker) Inspect(n ast.Node) (w ast.Node, skipChildren bool) {
	if n == nil {
		return n, true
	}

	switch node := n.(type) {
	case *ast.FuncCallExpr:
		return f.handleFunctionNode(node, strings.ToUpper(node.FnName.L), "函数")

	case *ast.AggregateFuncExpr:
		return f.handleFunctionNode(node, strings.ToUpper(node.F), "聚合函数")

	case *ast.WindowFuncExpr:
		return f.handleFunctionNode(node, strings.ToUpper(node.Name), "窗口函数")
	}

	// 不处理其他类型的节点，返回原节点
	return n, false
}

// handleFunctionNode 统一处理函数节点
// 参数:
//   - node: AST节点（FuncCallExpr、AggregateFuncExpr、WindowFuncExpr）
//   - funcName: 函数名称
//   - funcType: 函数类型描述
//
// 返回值:
//   - ast.Node: 转换后的节点
//   - bool: 是否跳过子节点
func (f *FunctionChecker) handleFunctionNode(node ast.Node, funcName, funcType string) (ast.Node, bool) {
	rules := f.GetRules()

	if rule, exists := rules[funcName]; exists {
		// 使用RuleChecker的通用issue生成机制
		f.AddIssue(model.Issue{
			Checker: f.Name(),
			Message: fmt.Sprintf("函数 %s: %s (建议: %s)", funcName, rule.Description, rule.Then.Target),
			AutoFix: model.AutoFix{
				Available: true,
				Action:    rule.Then.Action,
				Code:      fmt.Sprintf("%s -> %s", funcName, rule.Then.Target),
			},
		})

		// 执行AST转换
		transformedNode := f.ApplyTransformation(node, rule)
		return transformedNode, false
	}

	return node, false
}
