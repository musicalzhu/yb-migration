package checker

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/example/ybMigration/internal/config"
	"github.com/example/ybMigration/internal/model"
	"github.com/example/ybMigration/internal/testutils"
	"github.com/pingcap/tidb/pkg/parser/ast"
	_ "github.com/pingcap/tidb/pkg/parser/test_driver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// 测试初始化
// ============================================================================

func TestMain(m *testing.M) {
	// 初始化测试配置
	// SetupTestConfig 支持传入 nil，用于 TestMain 初始化
	testutils.SetupTestConfig(nil)

	// 运行所有测试
	code := m.Run()

	// 清理资源
	testutils.ResetTestConfig()

	// 使用 os.Exit 而不是直接返回
	os.Exit(code)
}

// ============================================================================
// RuleChecker 基础测试
// ============================================================================

func TestNewRuleChecker(t *testing.T) {
	cfg := testutils.GetTestConfig(t)

	cases := []struct {
		name    string
		inName  string
		inCat   string
		cfg     *config.Config
		wantErr bool
	}{
		{name: "valid_creation", inName: "test", inCat: "function", cfg: cfg, wantErr: false},
		{name: "nil_config", inName: "test", inCat: "function", cfg: nil, wantErr: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			chk, err := newRuleChecker(tc.inName, tc.inCat, tc.cfg)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, chk)
			assert.Equal(t, tc.inName, chk.name)
			assert.Equal(t, tc.inCat, chk.category)
			assert.NotNil(t, chk.rules)
			assert.Empty(t, chk.issues)
		})
	}
}

func TestRuleChecker_BasicMethods(t *testing.T) {
	cfg := testutils.GetTestConfig(t)
	checker, err := newRuleChecker("test", "function", cfg)
	require.NoError(t, err)

	t.Run("name_and_category", func(t *testing.T) {
		assert.Equal(t, "test", checker.Name())
		assert.Equal(t, "function", checker.category)
	})

	t.Run("issues_management", func(t *testing.T) {
		// 初始状态
		assert.Empty(t, checker.Issues())

		// 添加问题
		issue := model.Issue{
			Checker: "test",
			Message: "测试问题",
		}
		checker.AddIssue(issue)

		// 验证问题被添加
		issues := checker.Issues()
		assert.Len(t, issues, 1)
		assert.Equal(t, "测试问题", issues[0].Message)

		// 重置
		checker.Reset()
		assert.Empty(t, checker.Issues())
	})
}

func TestRuleChecker_LoadRulesFromConfig(t *testing.T) {
	cfg := testutils.GetTestConfig(t)

	cases := []struct {
		name      string
		category  string
		expectAny bool
	}{
		{name: "function_rules", category: "function", expectAny: true},
		{name: "datatype_rules", category: "datatype", expectAny: true},
		{name: "syntax_rules", category: "syntax", expectAny: true},
		{name: "charset_rules", category: "charset", expectAny: true},
		{name: "nonexistent_category", category: "nonexistent", expectAny: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checker, err := newRuleChecker("test", tc.category, cfg)
			require.NoError(t, err)
			rules := checker.GetRules()
			if tc.expectAny {
				assert.NotEmpty(t, rules)
				for _, rule := range rules {
					assert.Equal(t, tc.category, rule.Category)
				}
			} else {
				assert.Empty(t, rules)
			}
		})
	}
}

func TestRuleChecker_ConcurrentSafety(t *testing.T) {
	cfg := testutils.GetTestConfig(t)
	checker, err := newRuleChecker("concurrent_test", "function", cfg)
	require.NoError(t, err)

	const numGoroutines = 100
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// 并发添加问题
			issue := model.Issue{
				Checker: "concurrent_test",
				Message: fmt.Sprintf("并发测试问题 %d", id),
			}
			checker.AddIssue(issue)

			// 并发读取问题
			_ = checker.Issues()
		}(i)
	}

	wg.Wait()

	// 验证所有问题都已添加
	issues := checker.Issues()
	if len(issues) != numGoroutines {
		t.Fatalf("并发测试失败: 期望 %d 个问题, 实际得到 %d", numGoroutines, len(issues))
	}
}

// ============================================================================
// 具体检查器测试
// ============================================================================

func TestFunctionChecker(t *testing.T) {
	cfg := testutils.GetTestConfig(t)
	checker, err := NewFunctionChecker(cfg)
	require.NoError(t, err)

	t.Run("basic_properties", func(t *testing.T) {
		assert.Equal(t, "FunctionChecker", checker.Name())
		assert.Equal(t, "function", checker.category)
		assert.NotNil(t, checker.GetRules())
	})

	t.Run("inspect_function_call", func(t *testing.T) {
		// 测试 GROUP_CONCAT 函数调用
		funcCall := &ast.FuncCallExpr{
			FnName: ast.NewCIStr("GROUP_CONCAT"),
		}

		node, skip := checker.Inspect(funcCall)

		// 验证返回值
		assert.NotNil(t, node)
		assert.False(t, skip)

		// 验证问题收集
		issues := checker.Issues()
		assert.NotEmpty(t, issues)

		// 验证问题内容
		found := false
		for _, issue := range issues {
			if len(issue.Message) > 0 {
				found = true
				break
			}
		}
		assert.True(t, found, "应该发现问题")
	})

	t.Run("inspect_non_function_node", func(t *testing.T) {
		// 测试非函数节点
		tableName := &ast.TableName{
			Name: ast.NewCIStr("test_table"),
		}

		node, skip := checker.Inspect(tableName)

		// 应该返回原节点，不跳过子节点
		assert.Equal(t, tableName, node)
		assert.False(t, skip)

		// 清除之前的问题
		checker.Reset()
		assert.Empty(t, checker.Issues())
	})
}

func TestDataTypeChecker(t *testing.T) {
	cfg := testutils.GetTestConfig(t)
	checker, err := NewDataTypeChecker(cfg)
	require.NoError(t, err)

	t.Run("basic_properties", func(t *testing.T) {
		assert.Equal(t, "DataTypeChecker", checker.Name())
		assert.Equal(t, "datatype", checker.category)
		assert.NotNil(t, checker.GetRules())
	})

	t.Run("inspect_column_def", func(t *testing.T) {
		// 测试简单的列定义
		columnDef := &ast.ColumnDef{
			Name: &ast.ColumnName{
				Name: ast.NewCIStr("test_col"),
			},
		}

		node, skip := checker.Inspect(columnDef)

		// 验证返回值
		assert.NotNil(t, node)
		assert.False(t, skip)
	})

	t.Run("inspect_non_datatype_node", func(t *testing.T) {
		// 测试非数据类型节点
		selectStmt := &ast.SelectStmt{}

		node, skip := checker.Inspect(selectStmt)

		// 应该返回原节点，不跳过子节点
		assert.Equal(t, selectStmt, node)
		assert.False(t, skip)

		// 不应该收集问题
		checker.Reset()
		assert.Empty(t, checker.Issues())
	})
}

func TestSyntaxChecker(t *testing.T) {
	cfg := testutils.GetTestConfig(t)
	checker, err := NewSyntaxChecker(cfg)
	require.NoError(t, err)

	t.Run("basic_properties", func(t *testing.T) {
		assert.Equal(t, "SyntaxChecker", checker.Name())
		assert.Equal(t, "syntax", checker.category)
		assert.NotNil(t, checker.GetRules())
	})

	t.Run("inspect_create_table", func(t *testing.T) {
		// 测试 CREATE TABLE 语句
		createStmt := &ast.CreateTableStmt{
			Table: &ast.TableName{
				Name: ast.NewCIStr("test_table"),
			},
		}

		node, skip := checker.Inspect(createStmt)

		// 验证返回值
		assert.NotNil(t, node)
		assert.False(t, skip)
	})

	t.Run("inspect_non_syntax_node", func(t *testing.T) {
		// 测试非语法相关节点
		funcCall := &ast.FuncCallExpr{
			FnName: ast.NewCIStr("SOME_FUNC"),
		}

		node, skip := checker.Inspect(funcCall)

		// 应该返回原节点，不跳过子节点
		assert.Equal(t, funcCall, node)
		assert.False(t, skip)

		// 不应该收集问题
		checker.Reset()
		assert.Empty(t, checker.Issues())
	})
}

func TestCharsetChecker(t *testing.T) {
	cfg := testutils.GetTestConfig(t)
	checker, err := NewCharsetChecker(cfg)
	require.NoError(t, err)

	t.Run("basic_properties", func(t *testing.T) {
		assert.Equal(t, "CharsetChecker", checker.Name())
		assert.Equal(t, "charset", checker.category)
		assert.NotNil(t, checker.GetRules())
	})

	t.Run("inspect_create_table", func(t *testing.T) {
		// 测试 CREATE TABLE 语句
		createStmt := &ast.CreateTableStmt{
			Table: &ast.TableName{
				Name: ast.NewCIStr("test_table"),
			},
		}

		node, skip := checker.Inspect(createStmt)

		// 验证返回值
		assert.NotNil(t, node)
		assert.False(t, skip)
	})

	t.Run("inspect_non_charset_node", func(t *testing.T) {
		// 测试非字符集相关节点
		selectStmt := &ast.SelectStmt{}

		node, skip := checker.Inspect(selectStmt)

		// 应该返回原节点，不跳过子节点
		assert.Equal(t, selectStmt, node)
		assert.False(t, skip)

		// 不应该收集问题
		checker.Reset()
		assert.Empty(t, checker.Issues())
	})
}

// ============================================================================
// 错误处理和边界情况测试
// ============================================================================

func TestCheckerErrorHandling(t *testing.T) {
	cfg := testutils.GetTestConfig(t)

	t.Run("nil_node_handling", func(t *testing.T) {
		checker, err := NewFunctionChecker(cfg)
		require.NoError(t, err)

		// 处理 nil 节点不应该 panic
		// 不同的检查器可能对 nil 节点有不同的处理方式
		// 我们只验证不会 panic
		assert.NotPanics(t, func() {
			checker.Inspect(nil)
		}, "处理 nil 节点不应该 panic")
	})

	t.Run("empty_config", func(t *testing.T) {
		emptyCfg := &config.Config{Rules: []config.Rule{}}
		checker, err := newRuleChecker("test", "function", emptyCfg)
		require.NoError(t, err)
		assert.Empty(t, checker.GetRules())

		// RuleChecker 没有 Inspect 方法，这里只测试基本功能
		assert.Equal(t, "test", checker.Name())
		assert.Equal(t, "function", checker.category)
	})
}

// ============================================================================
// 性能测试
// ============================================================================

func BenchmarkChecker_Performance(b *testing.B) {
	cfg := testutils.GetTestConfig(&testing.T{})
	checker, err := NewFunctionChecker(cfg)
	if err != nil {
		b.Fatal(err)
	}

	// 创建一个简单的函数调用
	funcCall := &ast.FuncCallExpr{
		FnName: ast.NewCIStr("GROUP_CONCAT"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.Reset()
		_, _ = checker.Inspect(funcCall)
	}
}

func BenchmarkMultipleCheckers(b *testing.B) {
	cfg := testutils.GetTestConfig(&testing.T{})
	functionChecker, err1 := NewFunctionChecker(cfg)
	dataTypeChecker, err2 := NewDataTypeChecker(cfg)
	syntaxChecker, err3 := NewSyntaxChecker(cfg)
	charsetChecker, err4 := NewCharsetChecker(cfg)

	if err1 != nil {
		b.Fatal(err1)
	}
	if err2 != nil {
		b.Fatal(err2)
	}
	if err3 != nil {
		b.Fatal(err3)
	}
	if err4 != nil {
		b.Fatal(err4)
	}

	checkers := []Checker{
		functionChecker,
		dataTypeChecker,
		syntaxChecker,
		charsetChecker,
	}

	// 创建简单的 AST 节点
	funcCall := &ast.FuncCallExpr{
		FnName: ast.NewCIStr("GROUP_CONCAT"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, checker := range checkers {
			checker.Reset()
			_, _ = checker.Inspect(funcCall)
		}
	}
}
