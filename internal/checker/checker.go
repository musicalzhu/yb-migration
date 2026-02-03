// Package checker 提供SQL检查器的实现，用于分析和验证SQL语句
package checker

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/pingcap/tidb/pkg/parser/mysql"
	"github.com/pingcap/tidb/pkg/parser/types"

	"github.com/example/ybMigration/internal/config"
	"github.com/example/ybMigration/internal/model"
)

// Checker 检查器接口
type Checker interface {
	// Name 返回检查器名称
	Name() string
	// Inspect 检查AST节点
	// 参数:
	//   - node: 当前检查的AST节点
	// 返回:
	//   - ast.Node: 如果返回非nil，则替换当前节点
	//   - bool: 是否跳过子节点的检查
	Inspect(node ast.Node) (w ast.Node, skipChildren bool)
	// Issues 返回发现的问题
	Issues() []model.Issue
	// Reset 重置检查器状态
	Reset()
}

// 并发语义说明:
// - `Check` 在遍历 AST 时会在单个 goroutine 中调用每个检查器的 `Inspect` 方法，
//   因此 `Inspect` 的实现通常不需要为并发调用提供额外保护（在同一次遍历中是串行调用）。
// - 但外部代码可能并发读取检查器的结果（`Issues()`）或者并发调用 `AddIssue`/`AddIssues`，
//   因此建议检查器实现提供对问题集合的并发保护。`RuleChecker` 已通过内部的 `sync.RWMutex`
//   保证 `AddIssue`/`AddIssues`/`Issues`/`Reset` 在并发场景下是安全的。

// RuleChecker 规则检查器统一实现
// 提供完整的检查器功能：规则管理、问题收集、状态重置
// 支持不同类别的SQL兼容性检查，是所有检查器的基础实现
type RuleChecker struct {
	name     string                 // 检查器名称
	category string                 // 规则类别：指定检查器处理的规则类型（datatype/function/syntax/charset）
	rules    map[string]config.Rule // 规则映射：存储从配置文件加载的规则，key为Pattern的大写形式
	issues   []model.Issue          // 发现的问题列表
	mu       sync.RWMutex           // 读写锁：保护并发访问 `issues` 字段，保证对问题集合的并发读写安全
}

// newRuleChecker 创建规则检查器（包内私有）
// 参数:
//   - name: 检查器名称，用于标识和日志输出
//   - category: 规则类别，决定从配置文件中加载哪类规则
//   - cfg: 配置实例，不能为空
//
// 返回:
//   - *RuleChecker: 初始化后的规则检查器实例
//   - error: 错误信息
func newRuleChecker(name, category string, cfg *config.Config) (*RuleChecker, error) {
	if cfg == nil {
		return nil, fmt.Errorf("配置实例不能为空")
	}

	checker := &RuleChecker{
		name:     name,
		category: category,
		issues:   make([]model.Issue, 0),
	}

	// 加载规则
	checker.LoadRulesFromConfig(cfg)

	return checker, nil
}

// Name 返回检查器名称
func (r *RuleChecker) Name() string {
	return r.name
}

// AddIssues 添加问题列表
func (r *RuleChecker) AddIssues(issues []model.Issue) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.issues = append(r.issues, issues...)
}

// AddIssue 添加问题到检查器
// 参数:
//   - issue: 要添加的问题
func (r *RuleChecker) AddIssue(issue model.Issue) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.issues = append(r.issues, issue)
}

// Issues 返回发现的问题（只读，请勿修改返回的切片）
// 返回:
//   - []model.Issue: 问题列表，如果没有问题返回nil
//
// 并发安全: 该方法使用读锁保护 `issues`，并发调用是安全的。
func (r *RuleChecker) Issues() []model.Issue {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.issues
}

// Reset 重置检查器状态
// 并发安全: 使用写锁保证在并发场景下清空 `issues` 的原子性。
func (r *RuleChecker) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.issues = r.issues[:0] // 清空切片但保留底层数组
}

// LoadRulesFromConfig 从配置中加载规则
// 参数:
//   - cfg: 配置实例
func (r *RuleChecker) LoadRulesFromConfig(cfg *config.Config) {
	if cfg == nil {
		log.Printf("配置实例为空，跳过规则加载")
		return
	}

	r.rules = make(map[string]config.Rule)
	for _, rule := range cfg.Rules {
		// 关键：只加载与当前检查器category匹配的规则
		// strings.EqualFold() 进行不区分大小写的比较
		if strings.EqualFold(rule.Category, r.category) {
			// 将Pattern转换为大写作为key，便于后续不区分大小写的快速查找
			patternKey := strings.ToUpper(rule.When.Pattern)

			// 添加重复检测
			if existingRule, exists := r.rules[patternKey]; exists {
				log.Printf("警告：规则 '%s' 的 pattern '%s' 与现有规则 '%s' 重复，将覆盖",
					rule.Name, patternKey, existingRule.Name)
			}

			r.rules[patternKey] = rule
		}
	}
}

// GetRules 返回检查器加载的规则
// 返回:
//   - map[string]config.Rule: 规则映射
func (r *RuleChecker) GetRules() map[string]config.Rule {
	return r.rules
}

// TableOption 类型常量 (对应 TiDB AST 中的 TableOption.Tp)
// 这些常量用于标识表选项类型，应在包级别定义以便所有检查器使用
const (
	TableOptionCharacterSet = 2 // CHARACTER SET 选项
	TableOptionCollate      = 3 // COLLATE 选项
)

// 确保 visitor 实现了 ast.Visitor 接口
var _ ast.Visitor = (*visitor)(nil)

// visitor 实现 ast.Visitor 接口，用于遍历SQL AST
//
// 字段说明:
//   - checkers: 检查器列表，用于处理AST节点
//   - skipChildren: 布尔标志，指示是否跳过当前节点的子节点
//
// 并发安全:
//   - 该结构体不是并发安全的，每个goroutine应该使用独立的实例
type visitor struct {
	checkers     []Checker
	skipChildren bool // 缓存是否跳过子节点
}

// Reset 重置访问者状态
//
// 功能:
//   - 重置 skipChildren 标志为 false
//   - 在开始新的AST遍历前调用
//
// 使用场景:
//   - 复用visitor实例进行多次遍历
//   - 确保每次遍历的初始状态一致
func (v *visitor) Reset() {
	v.skipChildren = false
}

// getCheckerName 获取检查器名称的辅助函数。
// 通过反射获取检查器的类型名称，用于日志和错误报告。
// 参数:
//   - c: 检查器实例，不能为 nil
//
// 返回:
//   - string: 检查器的类型名称（格式：*package.TypeName）
func getCheckerName(c Checker) string {
	return fmt.Sprintf("%T", c)
}

// Enter 实现 ast.Visitor 接口
// 当进入节点时调用，返回处理后的节点和是否跳过子节点
//
// 参数:
//   - node: 当前访问的AST节点
//
// 返回值:
//   - ast.Node: 处理后的节点（通常返回原节点）
//   - bool: 是否跳过子节点遍历
//
// 实现细节:
//  1. 检查节点是否为nil或已标记跳过
//  2. 遍历所有检查器处理当前节点
//  3. 收集检查器的跳过子节点请求
//
// 并发安全:
//   - 该方法不是并发安全的，应该单线程调用
func (v *visitor) Enter(node ast.Node) (ast.Node, bool) {
	// 如果节点为nil或已标记跳过子节点，则直接返回
	if node == nil || v.skipChildren {
		return node, true
	}

	var skip bool
	// 遍历所有检查器处理当前节点
	for _, checker := range v.checkers {
		// 添加 defer 保护，防止检查器中的 panic
		func() {
			defer func() {
				if r := recover(); r != nil {
					// 记录 panic 但不中断遍历
					log.Printf("检查器 %v 处理节点 %T 时发生 panic: %v",
						getCheckerName(checker), node, r)
				}
			}()

			if n, s := checker.Inspect(node); n != nil || s {
				if n != nil {
					node = n // 替换节点
				}
				skip = skip || s // 任一检查器要求跳过子节点则跳过
			}
		}()
	}

	v.skipChildren = skip // 缓存跳过状态
	return node, skip
}

// Leave 实现 ast.Visitor 接口
// 当离开节点时调用，返回原始节点而不是 nil
func (v *visitor) Leave(node ast.Node) (ast.Node, bool) {
	return node, true
}

// CheckResult 检查和转换结果
type CheckResult struct {
	Issues           []model.Issue  // 发现的问题
	TransformedStmts []ast.StmtNode // 转换后的语句
}

// Check 检查和转换SQL语句（一次遍历完成所有工作）
// 参数:
//   - stmts: 要检查的SQL语句AST节点列表
//   - checkers: 要应用的检查器列表
//
// 返回:
//   - CheckResult: 包含问题和转换结果
func Check(stmts []ast.StmtNode, checkers ...Checker) CheckResult {
	// 快速返回空结果
	if len(checkers) == 0 || len(stmts) == 0 {
		return CheckResult{}
	}

	// 初始化所有检查器
	for _, checker := range checkers {
		checker.Reset()
	}

	// 创建访问者
	v := &visitor{checkers: checkers}

	// 一次遍历AST，同时完成分析和转换
	transformedStmts := make([]ast.StmtNode, len(stmts))
	for i, stmt := range stmts {
		if stmt == nil {
			continue
		}

		// 重置访问者状态
		v.Reset()

		// 单次遍历AST，应用所有检查器并支持节点转换
		newNode, _ := stmt.Accept(v)
		if newNode != nil {
			if stmtNode, ok := newNode.(ast.StmtNode); ok {
				transformedStmts[i] = stmtNode
			} else {
				transformedStmts[i] = stmt
			}
		} else {
			transformedStmts[i] = stmt
		}
	}

	// 收集所有检查器发现的问题
	var allIssues []model.Issue
	for _, checker := range checkers {
		if issues := checker.Issues(); len(issues) > 0 {
			allIssues = append(allIssues, issues...)
		}
	}

	return CheckResult{
		Issues:           allIssues,
		TransformedStmts: transformedStmts,
	}
}

// ApplyTransformation 应用规则转换
// 参数:
//   - node: AST节点
//   - rule: 要应用的规则
//
// 返回:
//   - ast.Node: 转换后的AST节点
func (r *RuleChecker) ApplyTransformation(node ast.Node, rule config.Rule) ast.Node {
	switch rule.Then.Action {
	case "replace_function":
		return r.replaceFunction(node, rule)
	case "replace_type":
		return r.replaceType(node, rule)
	case "replace_constraint":
		return r.replaceConstraint(node, rule)
	case "replace_quotes":
		return r.replaceQuotes(node, rule)
	case "replace_clause":
		return r.replaceClause(node, rule)
	case "replace_charset":
		return r.replaceCharset(node, rule)
	case "replace_collation":
		return r.replaceCollation(node, rule)
	default:
		return node
	}
}

// replaceFunction 替换函数调用
func (r *RuleChecker) replaceFunction(node ast.Node, rule config.Rule) ast.Node {
	switch n := node.(type) {
	case *ast.FuncCallExpr:
		return &ast.FuncCallExpr{
			FnName: ast.NewCIStr(rule.Then.Target),
			Args:   n.Args,
		}
	case *ast.AggregateFuncExpr:
		return &ast.AggregateFuncExpr{
			F:        rule.Then.Target,
			Args:     n.Args,
			Distinct: n.Distinct,
			Order:    n.Order,
		}
	case *ast.WindowFuncExpr:
		return &ast.WindowFuncExpr{
			Name: rule.Then.Target,
			Args: n.Args,
		}
	default:
		return node
	}
}

// replaceType 替换数据类型
func (r *RuleChecker) replaceType(node ast.Node, rule config.Rule) ast.Node {
	switch n := node.(type) {
	case *ast.ColumnDef:
		// 根据目标类型创建新的FieldType
		var newTp byte
		switch rule.Then.Target {
		case "TINYINT":
			newTp = mysql.TypeTiny
		case "SMALLINT":
			newTp = mysql.TypeShort
		case "INT":
			newTp = mysql.TypeLong
		case "BIGINT":
			newTp = mysql.TypeLonglong
		case "FLOAT":
			newTp = mysql.TypeFloat
		case "DOUBLE":
			newTp = mysql.TypeDouble
		case "DECIMAL":
			newTp = mysql.TypeNewDecimal
		case "DATE":
			newTp = mysql.TypeDate
		case "DATETIME":
			newTp = mysql.TypeDatetime
		case "TIMESTAMP":
			newTp = mysql.TypeTimestamp
		default:
			return node
		}
		newType := types.NewFieldType(newTp)
		n.Tp = newType
		return n
	default:
		return node
	}
}

// replaceConstraint 替换约束
func (r *RuleChecker) replaceConstraint(node ast.Node, rule config.Rule) ast.Node {
	switch n := node.(type) {
	case *ast.ColumnDef:
		// 处理 AUTO_INCREMENT 等约束
		if rule.Then.Target == "SERIAL" {
			// 将 AUTO_INCREMENT 转换为 SERIAL
			if n.Options != nil {
				for i, option := range n.Options {
					if option.Tp == ast.ColumnOptionAutoIncrement {
						// 移除 AUTO_INCREMENT 选项
						n.Options = append(n.Options[:i], n.Options[i+1:]...)
						break
					}
				}
			}
			// SERIAL 是 PostgreSQL 的自增类型，需要修改数据类型
			if n.Tp != nil {
				serialType := types.NewFieldType(mysql.TypeLong)
				n.Tp = serialType
			}
		}
		return n
	default:
		return node
	}
}

// replaceQuotes 替换引号
func (r *RuleChecker) replaceQuotes(node ast.Node, _ config.Rule) ast.Node {
	switch n := node.(type) {
	case *ast.TableName:
		// 处理表名中的反引号
		originalName := n.Name.String()
		if strings.Contains(originalName, "`") {
			newName := strings.ReplaceAll(originalName, "`", "\"")
			n.Name = ast.NewCIStr(newName)
		}
		return n
	case *ast.ColumnName:
		// 处理列名中的反引号
		originalName := n.Name.String()
		if strings.Contains(originalName, "`") {
			newName := strings.ReplaceAll(originalName, "`", "\"")
			n.Name = ast.NewCIStr(newName)
		}
		return n
	default:
		return node
	}
}

// replaceClause 替换子句
func (r *RuleChecker) replaceClause(node ast.Node, _ config.Rule) ast.Node {
	switch n := node.(type) {
	case *ast.Limit:
		// 处理 LIMIT 子句转换为 OFFSET FETCH
		// 这里需要创建新的节点结构，暂时返回原节点
		// 实际实现需要根据具体的SQL语法来构造新的AST节点
		return n
	default:
		return node
	}
}

// replaceCharset 替换字符集
func (r *RuleChecker) replaceCharset(node ast.Node, rule config.Rule) ast.Node {
	switch n := node.(type) {
	case *ast.TableOption:
		// 处理表选项中的字符集
		if n.Tp == TableOptionCharacterSet && n.StrValue != "" {
			n.StrValue = rule.Then.Target
		}
		return n
	default:
		return node
	}
}

// replaceCollation 替换排序规则
func (r *RuleChecker) replaceCollation(node ast.Node, rule config.Rule) ast.Node {
	switch n := node.(type) {
	case *ast.TableOption:
		// 处理表选项中的排序规则
		if n.Tp == TableOptionCollate && n.StrValue != "" {
			n.StrValue = rule.Then.Target
		}
		return n
	default:
		return node
	}
}
